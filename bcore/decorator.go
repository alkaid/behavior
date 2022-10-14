package bcore

import "go.uber.org/zap"

// IDecorator 装饰器,修饰子节点
type IDecorator interface {
	INode
	// Decorate 装饰子节点
	//  @param decorated
	Decorate(decorated INode)
	// Decorated 获取被装饰节点
	//  @receiver d
	//  @return INode
	Decorated(brain IBrain) INode
	IDecoratorWorker
}

type IDecoratorWorker interface {
	// CanDynamicDecorate 能否动态更换子节点
	//  @return bool
	CanDynamicDecorate() bool
}

var _ IDecorator = (*Decorator)(nil)

// Decorator 装饰器基类
//
//	@implement IDecorator
type Decorator struct {
	Container
	IDecoratorWorker
	decorated INode // 被装饰节点,也是子节点
}

// InitNodeWorker
//
//	@override Node.InitNodeWorker
//	@receiver c
//	@param worker
func (d *Decorator) InitNodeWorker(worker INodeWorker) error {
	err := d.Container.InitNodeWorker(worker)
	// 强转,由框架本身保证实例化时传进来的worker是自己(自己实现了IContainerWorker接口,故强转不会panic)
	d.IDecoratorWorker = worker.(IDecoratorWorker)
	return err
}

// CanDynamicDecorate
//
//	@implement IDecoratorWorker.CanDynamicDecorate
//	@receiver d
//	@return bool
func (d *Decorator) CanDynamicDecorate() bool {
	return false
}

// Decorate
//
//	@implement IDecorator.Decorate
//	@receiver n
//	@param decorated
func (d *Decorator) Decorate(decorated INode) {
	if d.IDecoratorWorker.CanDynamicDecorate() {
		d.Log(nil).Info("Warn:you are mount child to a dynamic decorator as static decorator", zap.String("childName", decorated.Name()), zap.String("childTitle", decorated.Title()))
	}
	d.decorated = decorated
	decorated.SetParent(d.INodeWorker.(IContainer))
}

// Decorated
//
//	@implement IDecorator.Decorated
//	@receiver d
//	@return INode
func (d *Decorator) Decorated(brain IBrain) INode {
	if brain == nil || !d.IDecoratorWorker.CanDynamicDecorate() {
		return d.decorated
	}
	dynamicDecorated := d.Memory(brain).DynamicChild
	if dynamicDecorated != nil {
		return dynamicDecorated
	}
	return d.decorated
}

// SetRoot
//
//	@override Node.SetRoot
//	@receiver n
//	@param root
func (d *Decorator) SetRoot(brain IBrain, root IRoot) {
	d.Container.SetRoot(brain, root)
	// 子节点非动态或者不为空时
	if d.decorated != nil {
		d.decorated.SetRoot(brain, root)
	}
}

// CompositeAncestorFinished
//
//	@override Node.CompositeAncestorFinished
//	@receiver n
//	@param brain
//	@param composite
func (d *Decorator) CompositeAncestorFinished(brain IBrain, composite IComposite) {
	d.Container.CompositeAncestorFinished(brain, composite)
	// 向下传播
	decorated := d.Decorated(brain)
	if decorated != nil {
		decorated.CompositeAncestorFinished(brain, composite)
	}
}

// Finish
//
//	@override Node.Finish
//	@receiver d
//	@param brain
//	@param succeeded
func (d *Decorator) Finish(brain IBrain, succeeded bool) {
	// 动态更换子节点
	if d.IDecoratorWorker.CanDynamicDecorate() && d.Memory(brain).RequestDynamicChild != nil {
		d.Memory(brain).DynamicChild = d.Memory(brain).RequestDynamicChild
		d.Memory(brain).RequestDynamicChild = nil
		d.Memory(brain).DynamicChild.DynamicMount(brain, d)
	}
	d.Container.Finish(brain, succeeded)
}
