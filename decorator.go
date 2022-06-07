package behavior

// IDecorator 装饰器,修饰子节点
type IDecorator interface {
	INode
	// Decorate 装饰子节点
	//  @param decorated
	Decorate(decorated INode)
	// Decorated 获取被装饰节点
	//  @receiver d
	//  @return INode
	Decorated() INode
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
func (d *Decorator) Decorate(decorated INode) {
	d.decorated = decorated
}

// Decorated
//  @implement IDecorator.Decorated
//  @receiver d
//  @return INode
func (d *Decorator) Decorated() INode {
	return d.decorated
}

// SetRoot
//  @override Node.SetRoot
//  @receiver n
//  @param root
func (d *Decorator) SetRoot(root IRoot) {
	d.Container.SetRoot(root)
	d.decorated.SetRoot(root)
}

// CompositeAncestorFinished
//  @override Node.CompositeAncestorFinished
//  @receiver n
//  @param brain
//  @param composite
func (d *Decorator) CompositeAncestorFinished(brain IBrain, composite IComposite) {
	d.Container.CompositeAncestorFinished(brain, composite)
	// 向下传播
	d.decorated.CompositeAncestorFinished(brain, composite)
}

// OnAbort
//  @override Node.OnAbort
//  @receiver r
//  @param brain
func (d *Decorator) OnAbort(brain IBrain) {
	d.Container.OnAbort(brain)
	// 向下传播
	d.decorated.Abort(brain)
}
