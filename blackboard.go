package behavior

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
	goID        int                   // 监听函数执行的线程ID
	treesMemory map[string]Memory     // 索引为行为树ID(rootID),元素为对应行为树的数据.仅允许框架内部CRUD
	nodesData   map[string]*NodeData  // 索引为节点ID,元素为对应节点的数据(固定).仅允许框架内部CRUD
	nodesMemory map[string]Memory     // 索引为节点ID,元素为对应节点的数据(扩展).仅允许框架内部CRUD
	userMemory  Memory                // 作用域为<用户域>的数据.仅允许业务方CRUD
	enable      bool                  // 是否开启
	observers   map[string][]Observer // 监听列表
}

// TreeMemory
//  @implement IBlackboardInternal.TreeMemory
//  @receiver b
//  @param rootID
//  @return Memory
//
func (b *Blackboard) TreeMemory(rootID string) Memory {
	mem, ok := b.treesMemory[rootID]
	if !ok {
		mem = make(Memory)
		b.treesMemory[rootID] = mem
	}
	return mem
}

// NodeMemory
//  @implement IBlackboardInternal.NodeMemory
//  @receiver b
//  @param nodeID
//  @return Memory
//
func (b *Blackboard) NodeMemory(nodeID string) Memory {
	mem, ok := b.nodesMemory[nodeID]
	if !ok {
		mem = make(Memory)
		b.treesMemory[nodeID] = mem
	}
	return mem
}

// NodeData
//  @implement IBlackboardInternal.NodeData
//  @receiver b
//  @param nodeID
//  @return *NodeData
//
func (b *Blackboard) NodeData(nodeID string) *NodeData {
	mem, ok := b.nodesData[nodeID]
	if !ok {
		mem = &NodeData{}
		b.nodesData[nodeID] = mem
	}
	return mem
}

// NewBlackboard 实例化一个黑板
//  @param goID AI工作线程ID
//  @return *Blackboard
//
func NewBlackboard(goID int) *Blackboard {
	b := &Blackboard{
		goID:        goID,
		treesMemory: make(map[string]Memory),
		nodesData:   make(map[string]*NodeData),
		nodesMemory: make(map[string]Memory),
		userMemory:  make(Memory),
		observers:   make(map[string][]Observer),
	}
	return b
}

// Start
//  @implement IBlackboardInternal.Start
//  @receiver b
//
func (b *Blackboard) Start() {
	b.enable = true
}

// Stop
//  @implement IBlackboardInternal.Stop
//  @receiver b
//
func (b *Blackboard) Stop() {
	b.enable = false
}

// AddOrRmObserver
//  @implement IBlackboardInternal.AddOrRmObserver
//  @receiver b
//  @param add
//  @param key
//  @param observer
//
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

// Get
//  @implement IBlackboard.Get
//  @receiver b
//  @param key
//  @return any
//  @return bool
//
func (b *Blackboard) Get(key string) (any, bool) {
	val, ok := b.userMemory[key]
	return val, ok
}

// Set
//  @implement IBlackboard.Set
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

// Del
//  @implement IBlackboard.Del
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

var _ IBlackboard = (*Blackboard)(nil)
var _ IBlackboardInternal = (*Blackboard)(nil)

// IBlackboard 黑板
type IBlackboard interface {
	// Get 获取Value,可链式传给 ConvertAnyValue[T]来转换类型
	//  @receiver b
	//  @param key
	//  @return any
	//  @return bool
	Get(key string) (any, bool)
	// Set 设置KV(用户域)
	//  线程安全
	//  @receiver b
	//  @param key
	//  @param val
	Set(key string, val any)
	// Del 删除KV
	//  线程安全
	//  @receiver b
	//  @param key
	Del(key string)
}

// IBlackboardInternal 框架内或自定义节点时使用的黑板,从 IBlackboard 转化来
//  含有私有API,业务层请勿调用,避免引发不可预期的后果
type IBlackboardInternal interface {
	IBlackboard
	// Start 启动,将会开始监听kv
	//  私有,框架内部使用
	//  非线程安全
	//  @receiver b
	Start()
	// Stop 停止,将会停止监听kv
	//  私有,框架内部使用
	//  非线程安全
	//  @receiver b
	Stop()
	// AddOrRmObserver 添加或删除监听(异步)
	//  私有,框架内部使用
	//  @receiver b
	//  @param add 是否添加
	//  @param key 黑板上的key
	//  @param observer 监听
	AddOrRmObserver(add bool, key string, observer Observer)
	// TreeMemory 树数据
	//  @param rootID
	//  @return Memory
	TreeMemory(rootID string) Memory
	// NodeMemory 节点的数据(扩展)
	//  @param nodeID
	//  @return Memory
	NodeMemory(nodeID string) Memory
	// NodeData 节点的数据(固定)
	//  @param nodeID
	//  @return *NodeData
	NodeData(nodeID string) *NodeData
}
