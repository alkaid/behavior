package bcore

import (
	"reflect"
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
	// OnExecute 执行委托方法，Node 框架内部调用
	//  @param target 委托对象名
	//  @param method 委托方法名
	//  @param args 参数列表,注意若多个请加上...
	//  @return []any
	//  @return error
	OnExecute(target string, method string, args ...any) ([]any, error)
}
