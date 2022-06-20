package task

import (
	"github.com/alkaid/behavior/bcore"
)

type ISubtreeProperties interface {
	GetChild() string
	GetIsSuccessWhenNotChild() bool
}

// SubtreeProperties 子树容器属性
type SubtreeProperties struct {
	Child                 string `json:"child"`                 // 默认子节点ID
	IsSuccessWhenNotChild bool   `json:"isSuccessWhenNotChild"` // 无子节点时执行是否返回成功
}

func (s *SubtreeProperties) GetChild() string {
	return s.Child
}
func (s *SubtreeProperties) GetIsSuccessWhenNotChild() bool {
	return s.IsSuccessWhenNotChild
}

// ISubtree 静态子树容器
//  只能在初始化时添加子节点,不允许在运行时改变子节点
type ISubtree interface {
	bcore.IDecorator
	// GetPropChildID 获取配置中的子节点ID
	//  @return string
	GetPropChildID() string
}

var _ ISubtree = (*Subtree)(nil)

// Subtree 静态子树容器
//  只能在初始化时添加子节点,不允许在运行时改变子节点
type Subtree struct {
	bcore.Decorator
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (t *Subtree) PropertiesClassProvider() any {
	return &SubtreeProperties{}
}

func (t *Subtree) SubtreeProperties() ISubtreeProperties {
	return t.Properties().(ISubtreeProperties)
}

// OnStart
//  @override Node.OnStart
//  @receiver n
//  @param brain
func (t *Subtree) OnStart(brain bcore.IBrain) {
	t.Decorator.OnStart(brain)
	// 无子节点时默认返回失败
	if t.Decorated(brain) == nil {
		if t.GetPropChildID() != "" {
			t.Log().Fatal("child not empty in properties,you must decorate child")
		}
		t.Finish(brain, t.SubtreeProperties().GetIsSuccessWhenNotChild())
		return
	}
	t.Decorated(brain).Start(brain)
}

// OnAbort
//  @override Node.OnAbort
//  @receiver n
//  @param brain
func (t *Subtree) OnAbort(brain bcore.IBrain) {
	t.Decorator.OnAbort(brain)
	// 向下传播
	decorated := t.Decorated(brain)
	if decorated != nil {
		decorated.Abort(brain)
		return
	}
	// 无子节点时默认返回失败
	t.Finish(brain, false)
}

// OnChildFinished
//  @override bcore.Decorator .OnChildFinished
//  @receiver s
//  @param brain
//  @param child
//  @param succeeded
func (t *Subtree) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	t.Decorator.OnChildFinished(brain, child, succeeded)
	t.Finish(brain, succeeded)
}

// GetPropChildID
//  @implement ISubtree.GetPropChildID
//  @receiver s
//  @return string
func (s *Subtree) GetPropChildID() string {
	return s.Properties().(ISubtreeProperties).GetChild()
}