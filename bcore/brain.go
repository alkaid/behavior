package bcore

import (
	"reflect"
	"time"
)

type DelegateMeta struct {
	Delegate     any
	ReflectValue reflect.Value
}

type FinishEvent struct {
	IsAbort   bool // 是否是终止的,false则为正常完成
	Succeeded bool // 是否成功
}

// IBrain 大脑=行为树+记忆+委托对象集合. 也可以理解为上下文
type IBrain interface {
	// FinishChan 树运行结束的信道
	//  @return <-chan
	FinishChan() <-chan *FinishEvent
	// Blackboard 获取黑板
	//  @return IBlackboard
	Blackboard() IBlackboard
	// GetDelegate 获取委托
	//  @param name
	//  @return delegate
	//  @return ok
	GetDelegate(name string) (delegate any, ok bool)
	// Go 在 IBrain 的独立线程里执行任务
	//  @param task
	Go(task func())
	// RunningTree 正在执行的树
	//  @return IRoot
	RunningTree() IRoot
	// Running 是否有正在执行的树
	//  @return bool
	Running() bool
	// ForceRun 执行指定树,若已经有运行中的树则终止后再启动
	//  @param root
	ForceRun(root IRoot)
}

// IBrainInternal 框架内部使用的 Brain
type IBrainInternal interface {
	// OnNodeUpdate 供节点回调执行委托
	//  @receiver b
	//  @param target
	//  @param method
	//  @param brain
	//  @param eventType
	//  @param delta
	//  @return bcore.Result
	OnNodeUpdate(target string, method string, brain IBrain, eventType EventType, delta time.Duration) Result

	// GetDelegates 获取委托map拷贝
	//  @receiver b
	//  @return map[string]any
	GetDelegates() map[string]any
	RWFinishChan() chan *FinishEvent
	SetFinishChan(finishChan chan *FinishEvent)
	SetRunningTree(root IRoot)
}
