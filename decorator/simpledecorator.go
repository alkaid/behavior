package decorator

import (
	"github.com/alkaid/behavior/bcore"
)

// SimpleDecorator 简单装饰器基类
//
//	子节点完成时强制返回失败
type SimpleDecorator struct {
	bcore.Decorator
}

// OnStart
//
//	@override Node.OnStart
//	@receiver n
//	@param brain
func (f *SimpleDecorator) OnStart(brain bcore.IBrain) {
	f.Decorator.OnStart(brain)
	f.Decorated(brain).Start(brain)
}

// OnAbort
//
//	@override Node.OnAbort
//	@receiver n
//	@param brain
func (f *SimpleDecorator) OnAbort(brain bcore.IBrain) {
	f.Decorator.OnAbort(brain)
	f.Decorated(brain).SetUpstream(brain, f)
	f.Decorated(brain).Abort(brain)
}

// OnChildFinished
//
//	@override bcore.Decorator .OnChildFinished
//	@receiver s
//	@param brain
//	@param child
//	@param succeeded
func (f *SimpleDecorator) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	f.Decorator.OnChildFinished(brain, child, succeeded)
	f.Finish(brain, succeeded)
}
