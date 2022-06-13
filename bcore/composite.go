package bcore

// IComposite 复合节点
type IComposite interface {
	IContainer
	AddChild(child INode)
	AddChildren(children []INode)
	Children() []INode
	// AbortLowerPriorityChildrenForChild 中断低优先级分支
	//  子类必须覆写
	//  @param childAbortBy 发起中断请求的子节点
	AbortLowerPriorityChildrenForChild(brain IBrain, childAbortBy INode)
}

var _ IComposite = (*Composite)(nil)

// Composite 复合节点基类
//  @implement IComposite
type Composite struct {
	Container
	children []INode
}

func (c *Composite) Children() []INode {
	return c.children
}
func (c *Composite) AddChild(child INode) {
	child.SetParent(c)
	c.children = append(c.children, child)
}

func (c *Composite) AddChildren(children []INode) {
	for _, child := range children {
		child.SetParent(c)
	}
	c.children = append(c.children, children...)
}

// SetRoot
//  @override Node.SetRoot
//  @receiver n
//  @param root
func (c *Composite) SetRoot(root IRoot) {
	c.Container.SetRoot(root)
	for _, child := range c.children {
		child.SetRoot(root)
	}
}

// Finish
//  @override Node.Finish
//  @receiver n
//  @param brain
//  @param succeeded
//
func (c *Composite) Finish(brain IBrain, succeeded bool) {
	// 广播所有子节点
	for _, child := range c.children {
		child.CompositeAncestorFinished(brain, c)
	}
	c.Container.Finish(brain, succeeded)
}

// AbortLowerPriorityChildrenForChild panic实现不可继承,子类须自行实现 IComposite.AbortLowerPriorityChildrenForChild
//  @receiver c
//  @param childAbortBy
func (c *Composite) AbortLowerPriorityChildrenForChild(brain IBrain, childAbortBy INode) {
	panic("implement me")
}
