package bcore

import "github.com/alkaid/behavior/thread"

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
	// DynamicDecorate 运行时动态更换子节点
	//  @param brain
	//  @param decorated 子节点
	//  @param abort 若在运行中是否要中断
	DynamicDecorate(brain IBrain, decorated INode, abort bool)
}

type IDecoratorWorker interface {
	// CanDynamicDecorate 能否动态更换子节点
	//  @return bool
	CanDynamicDecorate() bool
}

var _ IDecorator = (*Decorator)(nil)

// Decorator 装饰器基类
//  @implement IDecorator
type Decorator struct {
	Container
	IDecoratorWorker
	decorated INode // 被装饰节点,也是子节点
}

func (d *Decorator) DynamicDecorate(brain IBrain, decorated INode, abort bool) {
	if !d.CanDynamicDecorate() {
		d.Log().Fatal("cannot dynamic decorate child")
	}
	// 保证线程安全
	thread.GoByID(brain.Blackboard().(IBlackboardInternal).ThreadID(), func() {
		if d.IsInactive(brain) {
			d.Memory(brain).DynamicChild = decorated
			decorated.DynamicMount(brain, d)
			return
		}
		// 如果已经激活,需要等待子树完成或强制中断
		d.Memory(brain).RequestDynamicChild = decorated
		if abort && d.IsActive(brain) {
			d.Abort(brain)
		}
	})
}

// InitNodeWorker
//  @override Node.InitNodeWorker
//  @receiver c
//  @param worker
func (d *Decorator) InitNodeWorker(worker INodeWorker) error {
	err := d.Node.InitNodeWorker(worker)
	// 强转,由框架本身保证实例化时传进来的worker是自己(自己实现了IContainerWorker接口,故强转不会panic)
	d.IDecoratorWorker = worker.(IDecoratorWorker)
	return err
}

// CanDynamicDecorate
//  @implement IDecoratorWorker.CanDynamicDecorate
//  @receiver d
//  @return bool
func (d *Decorator) CanDynamicDecorate() bool {
	return false
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
func (d *Decorator) Decorated(brain IBrain) INode {
	if !d.IDecoratorWorker.CanDynamicDecorate() {
		return d.decorated
	}
	dynamicDecorated := d.Memory(brain).DynamicChild
	if dynamicDecorated != nil {
		return dynamicDecorated
	}
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
	decorated := d.Decorated(brain)
	if decorated != nil {
		decorated.CompositeAncestorFinished(brain, composite)
	}
}

// Finish
//  @override Node.Finish
//  @receiver d
//  @param brain
//  @param succeeded
func (d *Decorator) Finish(brain IBrain, succeeded bool) {
	// 动态更换子节点
	if d.CanDynamicDecorate() && d.Memory(brain).RequestDynamicChild != nil {
		d.Memory(brain).DynamicChild = d.Memory(brain).RequestDynamicChild
		d.Memory(brain).RequestDynamicChild = nil
		d.Memory(brain).DynamicChild.DynamicMount(brain, d)
	}
	d.Container.Finish(brain, succeeded)
}
