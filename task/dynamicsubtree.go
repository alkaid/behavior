package task

import (
	"github.com/alkaid/behavior/bcore"
)

type IDynamicSubtreeProperties interface {
	GetTag() string
	GetRunMode() bcore.DynamicBehaviorMode
}

// DynamicSubtreeProperties 子树容器属性
type DynamicSubtreeProperties struct {
	SubtreeProperties
	Tag     string                    `json:"tag"`     // 标签,用于识别子树
	RunMode bcore.DynamicBehaviorMode `json:"runMode"` // 动态子树中断模式
}

func (p *DynamicSubtreeProperties) GetTag() string {
	return p.Tag
}
func (p *DynamicSubtreeProperties) GetRunMode() bcore.DynamicBehaviorMode {
	return p.RunMode
}

// IDynamicSubtree 动态子树容器
//
//	可以在运行时动态更换子节点
type IDynamicSubtree interface {
	ISubtree
	Tag() string
	// DynamicDecorate 动态装饰子节点
	//
	// 非线程安全,由调用方自己保证
	// @param brain
	// @param decorated
	DynamicDecorate(brain bcore.IBrain, decorated bcore.IRoot)
}

var _ IDynamicSubtree = (*DynamicSubtree)(nil)

// DynamicSubtree 动态子树容器
//
//	可以在运行时动态更换子节点
type DynamicSubtree struct {
	Subtree
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (t *DynamicSubtree) PropertiesClassProvider() any {
	return &DynamicSubtreeProperties{}
}

// CanDynamicDecorate 标志可以运行时动态挂载子节点
//
//	@implement bcore.IDecoratorWorker .CanDynamicDecorate
//	@receiver t
//	@return bool
func (t *DynamicSubtree) CanDynamicDecorate() bool {
	return true
}
func (t *DynamicSubtree) Tag() string {
	return t.Properties().(IDynamicSubtreeProperties).GetTag()
}

// DynamicDecorate
//
//	@implement IDynamicSubtree.DynamicDecorate
//	@receiver d
//	@param brain
//	@param decorated
//	@param abort
func (t *DynamicSubtree) DynamicDecorate(brain bcore.IBrain, decorated bcore.IRoot) {
	if t.IsInactive(brain) {
		t.Memory(brain).DynamicChild = decorated
		decorated.DynamicMount(brain, t)
		return
	}
	// 如果已经激活,需要等待子树完成或强制中断
	t.Memory(brain).RequestDynamicChild = decorated
	if t.IsAborting(brain) {
		return
	}
	switch t.Properties().(IDynamicSubtreeProperties).GetRunMode() {
	case bcore.DynamicBehaviorModeContinue:
		return
	case bcore.DynamicBehaviorModeAbort:
		t.SetUpstream(brain, t)
		t.Abort(brain)
	case bcore.DynamicBehaviorModeRestart:
		t.Memory(brain).Restarting = true
		t.SetUpstream(brain, t)
		t.Abort(brain)
	}
}

// OnChildFinished
//
//	@override Decorator.OnChildFinished
//	@receiver s
//	@param brain
//	@param child
//	@param succeeded
func (t *DynamicSubtree) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	t.Decorator.OnChildFinished(brain, child, succeeded)
	if t.Memory(brain).Restarting {
		t.Start(brain)
		return
	}
	t.Finish(brain, succeeded)
}
