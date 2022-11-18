package bcore

import (
	"context"
	"reflect"
	"time"

	"github.com/alkaid/timingwheel"
)

type DelegateMeta struct {
	Delegate     any
	ReflectValue reflect.Value
}

type FinishEvent struct {
	IsAbort   bool // 是否是终止的,false则为正常完成
	Succeeded bool // 运行结果是否成功: Result is ResultSucceeded
	IsActive  bool // 停止前是否活跃
}

// IBrain 大脑=行为树+记忆+委托对象集合. 也可以理解为上下文
type IBrain interface {
	// FinishChan 树运行结束的信道
	//  @return <-chan
	FinishChan() <-chan *FinishEvent
	SetFinishChan(finishChan chan *FinishEvent)
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
	//
	// 非线程安全
	// @return bool
	Running() bool
	// Abort 终止正在运行的行为树
	//  线程安全
	//
	// @param abortChan
	Abort(abortChan chan *FinishEvent)
	// Run 异步启动
	//
	// @param tag
	// @param force 是否强制终止正在运行的树
	Run(tag string, force bool)
	// DynamicDecorate 给正在运行的树动态挂载子树
	//
	//	非线程安全,调用方自己保证
	//
	// @receiver b
	// @param containerTag 动态子树容器的tag
	// @param subtreeTag 子树的tag
	// @return error
	DynamicDecorate(containerTag string, subtreeTag string) error
	Cron(interval time.Duration, randomDeviation time.Duration, task func()) *timingwheel.Timer
	After(interval time.Duration, randomDeviation time.Duration, task func()) *timingwheel.Timer
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
	SetRunningTree(root IRoot)
	Context() context.Context
	SetContext(ctx context.Context)
}
