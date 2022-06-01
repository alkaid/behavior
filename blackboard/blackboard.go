package blackboard

import (
	stderr "errors"
	"sync"

	"github.com/alkaid/behavior/logger"

	"github.com/alkaid/behavior/thread"
	"github.com/samber/lo"
)

var ErrBlackboardAlreadyStarted = stderr.New("blackboard already started")
var ErrBlackboardAlreadyStopped = stderr.New("blackboard already stopped")

type OpType int // 操作类型

const (
	OpAdd    OpType = iota + 1 // 添加kv
	OpDel                      // 删除kv
	OpChange                   // 修改value
)

// Blackboard.observers 的元素无法存储范型方法,只能先改用any，未来go支持以后再改为范型方法
// type Observer[T any] func(op OpType, key string, oldValue T, newValue T)

type Observer func(op OpType, key string, oldValue any, newValue any)

// Blackboard 黑板,AI行为树实例的记忆
//  get和set为线程安全,其他方法均为非线程安全,请在树自己的线程内调用
//  黑板的[生命周期调用,监听添加移除,监听函数的执行]必须派发到AI独立线程
//  黑板的kv读写可以在任意线程
type Blackboard struct {
	memoryMutex sync.Mutex
	goID        int                    // 监听函数执行的线程ID
	treeMemory  map[string]*TreeMemory // 索引为行为树ID,元素为对应行为数的数据.仅允许框架内部CRUD
	userMemory  Memory                 // 作用域为<用户域>的数据.仅允许业务方CRUD
	enable      bool                   // 是否开启
	observers   map[string][]Observer  // 监听列表
}

// NewBlackboard 实例化一个黑板
//  @param goID AI工作线程ID
//  @return *Blackboard
//
func NewBlackboard(goID int) *Blackboard {
	b := &Blackboard{
		goID:       goID,
		treeMemory: make(map[string]*TreeMemory),
		userMemory: make(Memory),
		observers:  make(map[string][]Observer),
	}
	return b
}

// Start 启动,将会开始监听kv
//  私有,框架内部使用
//  非线程安全
//  @receiver b
//
func (b *Blackboard) Start() {
	b.enable = true
}

// Stop 停止,将会停止监听kv
//  私有,框架内部使用
//  非线程安全
//  @receiver b
//
func (b *Blackboard) Stop() {
	b.enable = false
}

// memory 根据树ID和节点ID获取内部集合
//  私有,框架内部使用
//  非线程安全
//  @receiver b
//  @param treeID
//  @param nodeID
//  @return *Memory
//
func (b *Blackboard) memory(treeID, nodeID string) Memory {
	var memory = b.userMemory
	if len(treeID) > 0 {
		treeMem, ok := b.treeMemory[treeID]
		if !ok {
			treeMem = newTreeMemory()
			b.treeMemory[treeID] = treeMem
		}
		memory = treeMem.memory
		if len(nodeID) > 0 {
			memory, ok = treeMem.nodes[nodeID]
			if !ok {
				memory = make(Memory)
				treeMem.nodes[nodeID] = memory
			}
		}
	}
	return memory
}

// AddOrRmObserver 添加或删除监听(异步)
//  私有,框架内部使用
//  @receiver b
//  @param add 是否添加
//  @param key 黑板上的key
//  @param observer 监听
func (b *Blackboard) AddOrRmObserver(add bool, key string, observer Observer) {
	// 无论调用方是否在AI线程里,都兜底派发到AI线程,避免和监听函数并行
	thread.GoByID(b.goID, func() {
		observers, ok := b.observers[key]
		contains := false
		if !ok {
			observers = make([]Observer, 0)
			b.observers[key] = observers
		} else {
			contains = lo.ContainsBy(observers, func(v Observer) bool { return &v == &observer })
		}
		if !add {
			if contains {
				b.observers[key] = lo.DropWhile(observers, func(v Observer) bool {
					return &v == &observer
				})
				return
			}
			logger.Log.Warn("remove observer failed. observers not contained this observer")
			return
		}
		if !contains {
			b.observers[key] = append(b.observers[key], observer)
			return
		}
		logger.Log.Warn("add observer failed. observers already contained this observer")
	})
}

// notify 黑板数据(用户域)改变时通知监听函数执行
//  @receiver b
//  @param op
//  @param key
//  @param oldVal
//  @param newVal
//
func (b *Blackboard) notify(op OpType, key string, oldVal any, newVal any) {
	// 无论调用方是否在AI线程里,都兜底派发到AI线程,使监听函数在AI线程里串行
	thread.GoByID(b.goID, func() {
		if !b.enable {
			return
		}
		for _, ob := range b.observers[key] {
			ob(op, key, oldVal, newVal)
		}
	})
}

// Set 设置KV(用户域)
//  线程安全
//  @receiver b
//  @param key
//  @param val
//
func (b *Blackboard) Set(key string, val any) {
	b.memoryMutex.Lock()
	defer b.memoryMutex.Unlock()
	op := OpAdd
	oldVal, ok := b.userMemory[key]
	if ok {
		op = OpChange
	}
	b.userMemory[key] = val
	b.notify(op, key, oldVal, val)
}

// Del 删除KV
//  线程安全
//  @receiver b
//  @param key
//
func (b *Blackboard) Del(key string) {
	b.memoryMutex.Lock()
	defer b.memoryMutex.Unlock()
	op := OpDel
	oldVal, ok := b.userMemory[key]
	if ok {
		delete(b.userMemory, key)
		b.notify(op, key, oldVal, nil)
	}
}

// SetTreeValue 设置KV(行为树域)
//  私有,框架内部使用
//  @receiver b
//  @param treeID
//  @param nodeID
//  @param key
//  @param val
//
func (b *Blackboard) SetTreeValue(treeID string, nodeID string, key string, val any) {
	b.memory(treeID, nodeID)[key] = val
}

// DelTreeValue 删除KV(行为树域)
//  @receiver b
//  @param treeID
//  @param nodeID
//  @param key
//
func (b *Blackboard) DelTreeValue(treeID string, nodeID string, key string) {
	delete(b.memory(treeID, nodeID), key)
}

// Get [T any] 获取黑板上的值
//  @param blackboard
//  @param key
//  @return T
//
func Get[T any](blackboard *Blackboard, key string) (T, bool) {
	return get[T](blackboard.userMemory, key)
}

// GetTreeValue [T any] 获取黑板上行为树域内的值,框架内部使用,业务方请勿调用
//  @param blackboard
//  @param treeID
//  @param nodeID
//  @param key
//  @return T
//
func GetTreeValue[T any](blackboard *Blackboard, treeID string, nodeID string, key string) (T, bool) {
	return get[T](blackboard.memory(treeID, nodeID), key)
}
