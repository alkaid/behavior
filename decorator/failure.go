package decorator

import (
	"github.com/alkaid/behavior/bcore"
)

// Failure 强制失败装饰器
//  子节点完成时强制返回失败
type Failure struct {
	SimpleDecorator
}

// OnChildFinished
//  @override bcore.Decorator .OnChildFinished
//  @receiver s
//  @param brain
//  @param child
//  @param succeeded
func (f *Failure) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	f.Decorator.OnChildFinished(brain, child, succeeded)
	f.Finish(brain, false)
}
