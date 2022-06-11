package decorator

import (
	"github.com/alkaid/behavior/bcore"
)

// Succeeded 强制成功装饰器
//  子节点完成时强制返回失败
type Succeeded struct {
	SimpleDecorator
}

// OnChildFinished
//  @override bcore.Decorator .OnChildFinished
//  @receiver s
//  @param brain
//  @param child
//  @param succeeded
func (s *Succeeded) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	s.Decorator.OnChildFinished(brain, child, succeeded)
	s.Finish(brain, true)
}
