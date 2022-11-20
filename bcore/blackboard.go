package bcore

import (
	stderr "errors"
	"reflect"
	"sync"
	"time"

	"go.uber.org/zap"

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

// Blackboard
//
//	@implement IBlackboard
//	@implement IBlackboardInternal
//	get和set为线程安全,其他方法均为非线程安全,请在树自己的线程内调用
//	黑板的[生命周期调用,监听添加移除,监听函数的执行]必须派发到AI独立线程
//	黑板的kv读写可以在任意线程
//	黑板为树形结构,实例化时可指定父黑板,将继承父黑板的KV.父黑板,一般来说是AI集群的共享黑板。想实现AI间通信时这将很有用.
type Blackboard struct {
	memoryMutex sync.RWMutex
	threadID    int                    // 监听函数执行的线程ID
	treesMemory map[string]Memory      // 索引为行为树ID(rootID),元素为对应行为树的数据<行为树域>.仅允许框架内部CRUD
	nodesData   map[string]*NodeMemory // 索引为节点ID,元素为对应节点的数据<节点域>.仅允许节点内部CRUD
	userMemory  Memory                 // 作用域为<用户域>的数据.仅允许业务方CRUD
	enable      bool                   // 是否开启
	observers   map[string][]Observer  // 监听列表
	parent      *Blackboard            // 父黑板,一般来说是AI集群的共享黑板
	children    []*Blackboard          // 子黑板
}

func (b *Blackboard) ThreadID() int {
	return b.threadID
}

// TreeMemory
//
//	@implement IBlackboardInternal.TreeMemory
//	@receiver b
//	@param rootID
//	@return Memory
func (b *Blackboard) TreeMemory(rootID string) Memory {
	mem, ok := b.treesMemory[rootID]
	if !ok {
		mem = make(Memory)
		b.treesMemory[rootID] = mem
	}
	return mem
}

// NodeExt
//
//	@implement IBlackboardInternal.NodeExt
//	@receiver b
//	@param nodeID
//	@return Memory
func (b *Blackboard) NodeExt(nodeID string) Memory {
	return b.NodeMemory(nodeID).Ext
}

// NodeMemory
//
//	@implement IBlackboardInternal.NodeMemory
//	@receiver b
//	@param nodeID
//	@return *NodeMemory
func (b *Blackboard) NodeMemory(nodeID string) *NodeMemory {
	mem, ok := b.nodesData[nodeID]
	if !ok {
		mem = NewNodeMemory()
		b.nodesData[nodeID] = mem
	}
	return mem
}

// NewBlackboard 实例化一个黑板
//
//	@param threadID AI工作线程ID
//	@param parent 父黑板,一般来说是AI集群的共享黑板
//	@return *Blackboard
func NewBlackboard(threadID int, parent *Blackboard) *Blackboard {
	if threadID <= 0 {
		logger.Log.Fatal("threadID cannot <=0")
	}
	b := &Blackboard{
		threadID:    threadID,
		treesMemory: make(map[string]Memory),
		nodesData:   make(map[string]*NodeMemory),
		userMemory:  make(Memory),
		observers:   make(map[string][]Observer),
		parent:      parent,
		children:    make([]*Blackboard, 0),
	}
	return b
}

// Start
//
//	@implement IBlackboardInternal.Start
//	@receiver b
func (b *Blackboard) Start() {
	if b.enable {
		logger.Log.Warn("blackboard already stared")
		return
	}
	b.enable = true
	if b.parent != nil {
		b.parent.children = append(b.parent.children, b)
	}
}

// Stop
//
//	@implement IBlackboardInternal.Stop
//	@receiver b
func (b *Blackboard) Stop() {
	if !b.enable {
		logger.Log.Warn("blackboard already stopped")
		return
	}
	b.enable = false
	// 销毁所有树和节点数据,销毁所有监听
	b.observers = map[string][]Observer{}
	b.treesMemory = map[string]Memory{}
	b.nodesData = map[string]*NodeMemory{}
	// 从父黑板中移除自己
	if b.parent != nil {
		b.parent.children = lo.DropWhile(b.parent.children, func(v *Blackboard) bool { return v == b })
	}
}

// AddObserver
//
//	@implement IBlackboardInternal.AddObserver
//	@receiver b
//	@param key
//	@param observer
func (b *Blackboard) AddObserver(key string, observer Observer) {
	b.addOrRmObserver(true, key, observer)
}

// RemoveObserver
//
//	@implement IBlackboardInternal.RemoveObserver
//	@receiver b
//	@param key
//	@param observer
func (b *Blackboard) RemoveObserver(key string, observer Observer) {
	b.addOrRmObserver(false, key, observer)
}

// addOrRmObserver
//
//	@receiver b
//	@param add
//	@param key
//	@param observer
func (b *Blackboard) addOrRmObserver(add bool, key string, observer Observer) {
	// 无论调用方是否在AI线程里,都兜底派发到AI线程,避免和监听函数并行
	thread.GoByID(b.threadID, func() {
		if !b.enable {
			return
		}
		observers, ok := b.observers[key]
		contains := false
		if !ok {
			observers = make([]Observer, 0)
			b.observers[key] = observers
		} else {
			observerPtr := reflect.ValueOf(observer).Pointer()
			contains = lo.ContainsBy(observers, func(v Observer) bool { return observerPtr == reflect.ValueOf(v).Pointer() })
		}
		if !add {
			if contains {
				observerPtr := reflect.ValueOf(observer).Pointer()
				b.observers[key] = lo.DropWhile(observers, func(v Observer) bool {
					return observerPtr == reflect.ValueOf(v).Pointer()
				})
				return
			}
			logger.Log.Debug("[blackboard]remove observer failed. observers not contained this observer", zap.String("key", key))
			return
		}
		if !contains {
			b.observers[key] = append(b.observers[key], observer)
			return
		}
		logger.Log.Debug("[blackboard]add observer failed. observers already contained this observer", zap.String("key", key))
	})
}

// notify 黑板数据(用户域)改变时通知监听函数执行
//
//	@receiver b
//	@param op
//	@param key
//	@param oldVal
//	@param newVal
func (b *Blackboard) notify(op OpType, key string, oldVal any, newVal any) {
	// 无论调用方是否在AI线程里,都兜底派发到AI线程,使监听函数在AI线程里串行
	thread.GoByID(b.threadID, func() {
		if !b.enable {
			return
		}
		for _, ob := range b.observers[key] {
			ob(op, key, oldVal, newVal)
		}
	})
}

// Get
//
//	@implement IBlackboard.Get
//	@receiver b
//	@param key
//	@return any
//	@return bool
func (b *Blackboard) Get(key string) (any, bool) {
	b.memoryMutex.RLock()
	defer b.memoryMutex.RUnlock()
	val, ok := b.userMemory[key]
	if ok || b.parent == nil {
		return val, ok
	}
	// 找不到时回溯父黑板
	val, ok = b.parent.Get(key)
	return val, ok
}

// GetDuration
//
//	@implement IBlackboard.GetDuration
//	@receiver b
//	@param key
//	@return time.Duration
//	@return bool
func (b *Blackboard) GetDuration(key string) (time.Duration, bool) {
	val, ok := b.Get(key)
	switch v := val.(type) {
	case time.Duration:
		return v, ok
	case int64:
		return time.Duration(v), ok
	case string:
		tm, err := time.ParseDuration(v)
		if err != nil {
			logger.Log.Error("convert duration error", zap.Error(err), zap.String("key", key), zap.Any("value", val))
			return tm, false
		}
		return tm, true
	}
	logger.Log.Error("convert duration error,unsupport type", zap.String("key", key), zap.Any("value", val))
	return 0, false
}

// Set
//
//	@implement IBlackboard.Set
//	@receiver b
//	@param key
//	@param val
func (b *Blackboard) Set(key string, val any) {
	b.memoryMutex.Lock()
	defer b.memoryMutex.Unlock()
	// 优先设置父黑板
	if b.parent != nil {
		_, ok := b.parent.Get(key)
		if ok {
			b.parent.Set(key, val)
			return
		}
	}
	op := OpAdd
	oldVal, ok := b.userMemory[key]
	if ok {
		op = OpChange
	}
	b.userMemory[key] = val
	b.notify(op, key, oldVal, val)
}

// Del
//
//	@implement IBlackboard.Del
//	@receiver b
//	@param key
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

// IBlackboard 黑板,AI行为树实例的记忆
type IBlackboard interface {
	// Get 获取Value,可链式传给 ConvertAnyValue[T]来转换类型
	//  @receiver b
	//  @param key
	//  @return any
	//  @return bool
	Get(key string) (any, bool)
	// GetDuration 获取 time.Duration 类型的值,支持 int64 | string | time.Duration. string 格式参考 time.ParseDuration
	//  @receiver b
	//  @param key
	//  @return time.Duration
	//  @return bool
	GetDuration(key string) (time.Duration, bool)
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
//
//	含有私有API,业务层请勿调用,避免引发不可预期的后果
type IBlackboardInternal interface {
	IBlackboard
	// ThreadID 获取线程ID
	//  @return int
	ThreadID() int
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
	AddObserver(key string, observer Observer)
	RemoveObserver(key string, observer Observer)
	// TreeMemory 树数据
	//  @param rootID
	//  @return Memory
	TreeMemory(rootID string) Memory
	// NodeExt 节点的数据(扩展)
	//  @param nodeID
	//  @return Memory
	NodeExt(nodeID string) Memory
	// NodeMemory 节点的数据(固定)
	//  @param nodeID
	//  @return *NodeMemory
	NodeMemory(nodeID string) *NodeMemory
}
