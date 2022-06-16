package bcore

import (
	"reflect"
	"time"
)

type DelegateMeta struct {
	Delegate     any
	ReflectValue reflect.Value
}

// IBrain 大脑=行为树+记忆+委托对象集合. 也可以理解为上下文
type IBrain interface {
	// Tree() (*BehaviorTree, error)

	// Blackboard 获取黑板
	//  @return IBlackboard
	Blackboard() IBlackboard
	// GetDelegate 获取委托
	//  @param name
	//  @return delegate
	//  @return ok
	GetDelegate(name string) (delegate any, ok bool)
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
}
