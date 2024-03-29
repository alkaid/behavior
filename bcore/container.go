package bcore

import "go.uber.org/zap"

// IContainer 容器:可以挂载子节点的节点
type IContainer interface {
	INode
	// ChildFinished 收到子节点停止结束的调用
	//  IContainer 的子节点须在 INode.Finish 中调用 父节点的 ChildFinished 以通知父节点,父节点再根据自己的控制逻辑决定是否回溯
	//  停止链路请参看 Abort
	//  @param child
	//  @param succeeded
	ChildFinished(brain IBrain, child INode, succeeded bool)
}

// IContainerWorker Container 的回调,Container 的子类必须实现该接口
type IContainerWorker interface {
	// OnChildFinished IContainer.ChildFinished 的回调
	//  @param child
	//  @param succeeded
	OnChildFinished(brain IBrain, child INode, succeeded bool)
}

var _ IContainer = (*Container)(nil)
var _ IContainerWorker = (*Container)(nil)

// Container 容器基类
//
//	@implement IContainer
//	@implement IContainerWorker
type Container struct {
	Node
	IContainerWorker
}

// InitNodeWorker
//
//	@override Node.InitNodeWorker
//	@receiver c
//	@param worker
func (c *Container) InitNodeWorker(worker INodeWorker) error {
	err := c.Node.InitNodeWorker(worker)
	// 强转,由框架本身保证实例化时传进来的worker是自己(自己实现了IContainerWorker接口,故强转不会panic)
	c.IContainerWorker = worker.(IContainerWorker)
	return err
}

// OnChildFinished
//
//	@implement IContainerWorker.OnChildFinished
//	@receiver c
//	@param brain
//	@param child
//	@param succeeded
func (c *Container) OnChildFinished(brain IBrain, child INode, succeeded bool) {
	c.Log(brain).Debug("OnChildFinished", zap.String("child", child.Title()), zap.Bool("succeeded", succeeded), zap.String("content", c.String(brain)))
}

// ChildFinished
//
//	@implement IContainer.ChildFinished
//	@receiver c
//	@param brain
//	@param child
//	@param succeeded
func (c *Container) ChildFinished(brain IBrain, child INode, succeeded bool) {
	if c.IsInactive(brain) {
		c.Log(brain).Error("A ChildID of a Container was stopped while the container was inactive!")
		return
	}
	c.IContainerWorker.OnChildFinished(brain, child, succeeded)
}
