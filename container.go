package behavior

import "github.com/alkaid/behavior/logger"

// IContainer 容器:可以挂载子节点的节点
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

var _ IContainer = (*Container)(nil)
var _ IContainerWorker = (*Container)(nil)

// Container 容器基类
//  @implement IContainer
//  @implement IContainerWorker
type Container struct {
	*Node
}

func (c *Container) OnChildFinished(brain IBrain, child INode, succeeded bool) {
	logger.Log.Debug(c.String() + " OnChildFinished")
}

func (c *Container) ChildFinished(brain IBrain, child INode, succeeded bool) {
	state := brain.Blackboard().(IBlackboardInternal).NodeData(c.id).State
	if state == NodeStateInactive {
		logger.Log.Error("A Child of a Container was stopped while the container was inactive!")
		return
	}
	c.OnChildFinished(brain, child, succeeded)
}
