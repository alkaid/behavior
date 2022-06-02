package behavior

type IContainer interface {
	INode

	// ChildFinished 收到子节点停止结束的调用
	//  IContainer 的子节点须在 INode.Finish 中调用 父节点的 ChildFinished 以通知父节点,父节点再根据自己的控制逻辑决定是否回溯
	//  停止链路请参看 Cancel
	//  @param child
	//  @param succeeded
	ChildFinished(brain IBrain, child INode, succeeded bool)
}

type IContainerWorker interface {
	INodeWorker
	// OnChildFinished IContainer.ChildFinished 的回调
	//  @param child
	//  @param succeeded
	OnChildFinished(brain IBrain, child INode, succeeded bool)
}
