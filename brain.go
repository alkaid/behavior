package behavior

// IBrain 大脑=行为树+记忆. 也可以理解为上下文
type IBrain interface {
	// Tree() (*BehaviorTree, error)
	Blackboard() IBlackboard
}

// IDelegate 任务委托
type IDelegate interface {
	// OnTaskExecute 任务触发时
	//  @param brain
	//  @return Result
	OnTaskExecute(brain IBrain) Result
}
