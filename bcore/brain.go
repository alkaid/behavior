package bcore

// IBrain 大脑=行为树+记忆+委托对象集合. 也可以理解为上下文
type IBrain interface {
	// Tree() (*BehaviorTree, error)

	// Blackboard 获取黑板
	//  @return IBlackboard
	Blackboard() IBlackboard
	// Execute 执行委托方法
	//  @param target 委托对象名
	//  @param method 委托方法名
	//  @param args 参数列表,注意若多个请加上...
	//  @return []any
	//  @return error
	Execute(target string, method string, args ...any) ([]any, error)
}
