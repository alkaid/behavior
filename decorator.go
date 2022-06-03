package behavior

// IDecorator 装饰器,修饰子节点
type IDecorator interface {
	INode
	// Decorate 装饰子节点
	//  @param decorated
	Decorate(decorated INode)
}

var _ IDecorator = (*Decorator)(nil)

// Decorator 装饰器基类
//  @implement IDecorator
type Decorator struct {
	Container
	decorated INode // 被装饰节点,也是子节点
}

// Decorate
//  @implement IDecorator.Decorate
//  @receiver n
//  @param decorated
func (n *Decorator) Decorate(decorated INode) {
	n.decorated = decorated
}

// SetRoot
//  @override Node.SetRoot
//  @receiver n
//  @param root
func (n *Decorator) SetRoot(root IRoot) {
	n.Container.SetRoot(root)
	n.decorated.SetRoot(root)
}

// CompositeAncestorFinished
//  @override Node.CompositeAncestorFinished
//  @receiver n
//  @param brain
//  @param composite
func (n *Decorator) CompositeAncestorFinished(brain IBrain, composite IComposite) {
	n.Container.CompositeAncestorFinished(brain, composite)
	n.decorated.CompositeAncestorFinished(brain, composite)
}
