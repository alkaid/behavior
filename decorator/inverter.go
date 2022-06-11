package decorator

import (
	"github.com/alkaid/behavior/bcore"
)

// Inverter 结果反转装饰器
//  子节点完成时强制反转结果
type Inverter struct {
	SimpleDecorator
}

// OnChildFinished
//  @override bcore.Decorator .OnChildFinished
//  @receiver s
//  @param brain
//  @param child
//  @param succeeded
func (i *Inverter) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	i.Decorator.OnChildFinished(brain, child, succeeded)
	i.Finish(brain, !succeeded)
}
