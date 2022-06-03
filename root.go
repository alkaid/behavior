package behavior

import (
	"encoding/json"

	"github.com/alkaid/behavior/thread"

	"github.com/alkaid/behavior/logger"
	"go.uber.org/zap"
)

type rootProperties struct {
	Once bool `json:"once"` // 是否仅运行一次,反之永远循环
}

type IRoot interface {
	IDecorator
	// IsSubTree 是否父节点
	//  @param brain
	//  @return bool
	IsSubTree(brain IBrain) bool
}

var _ IRoot = (*Root)(nil)

type Root struct {
	Decorator
}

// IsSubTree
//  @implement IRoot.IsSubTree
//  @receiver r
//  @param brain
//  @return bool
func (r *Root) IsSubTree(brain IBrain) bool {
	// 父节点不为空就是子树
	return r.Parent(brain) != nil
}

// OnParseProperties
//  @override Node.OnParseProperties
//  @receiver r
//  @param properties
//  @return any
func (r *Root) OnParseProperties(properties json.RawMessage) any {
	var prop rootProperties
	err := json.Unmarshal(properties, &prop)
	if err != nil {
		logger.Log.Error("", zap.Error(err))
	}
	return prop
}

// Parent
//  @override Node.Parent
//  @receiver n
//  @return IContainer
//
func (r *Root) Parent(brain IBrain) IContainer {
	// root节点的parent获取须通过blackboard
	p := brain.Blackboard().(IBlackboardInternal).NodeData(r.Id()).RootParent
	return p
}

// SetRoot
//  @override Decorator.SetRoot
//  @receiver n
//  @param root
func (r *Root) SetRoot(root IRoot) {
	if root.Id() != r.Id() {
		logger.Log.Fatal("root's root must be self", zap.String("selfID", r.Id()), zap.String("inID", root.Id()))
	}
	r.Decorator.SetRoot(root)
}

// SetParent
//  @implement Node.SetParent
//  @receiver n
//  @param parent
//
func (r *Root) SetParent(brain IBrain, parent IContainer) {
	// 子树的root设置parent须通过黑板
	if parent != nil {
		brain.Blackboard().(IBlackboardInternal).NodeData(r.id).RootParent = parent
	}
}

// OnStart
//  @override Node.OnStart
//  @receiver n
//  @param brain
func (r *Root) OnStart(brain IBrain) {
	r.Decorator.OnStart(brain)
	// 非子树要开启黑板监听
	if !r.IsSubTree(brain) {
		brain.Blackboard().(IBlackboardInternal).Start()
	}
	// 开启子节点
	r.decorated.Start(brain)
}

// OnCancel
//  @override Node.OnCancel
//  @receiver r
//  @param brain
func (r *Root) OnCancel(brain IBrain) {
	if brain.Blackboard().(IBlackboardInternal).NodeData(r.decorated.Id()).State == NodeStateActive {
		r.decorated.Cancel(brain)
	}
	// TODO 这里是否有异步异常情况未考虑
}

// OnChildFinished
//  @override Container.OnChildFinished
//  @receiver r
//  @param brain
//  @param child
//  @param succeeded
func (r *Root) OnChildFinished(brain IBrain, child INode, succeeded bool) {
	r.Decorator.OnChildFinished(brain, child, succeeded)
	// 如果是外部触发取消的,结束运行
	if brain.Blackboard().(IBlackboardInternal).NodeData(r.id).State == NodeStateCanceling {
		// 非子树要关闭黑板监听
		if !r.IsSubTree(brain) {
			brain.Blackboard().(IBlackboardInternal).Stop()
		}
		// 无论是否子树都要结束root,若是子树将回溯parent,否则整棵行为树终止运行。
		r.Finish(brain, succeeded)
	} else {
		// 如果是正常结束，判断是否子树,子树则正常结束当前root
		// 如果配置了一次性运行,同上
		if r.IsSubTree(brain) {
			r.Finish(brain, succeeded)
		} else if r.properties.(rootProperties).Once {
			// 若是主树且是一次性,结束root并关闭监听
			brain.Blackboard().(IBlackboardInternal).Stop()
			r.Finish(brain, succeeded)
		} else {
			// 若是主树且可以循环，开启下一轮
			// 不能直接 Start(),会堆栈溢出,应该重新异步派发
			thread.GoByID(brain.Blackboard().(IBlackboardInternal).ThreadID(), func() {
				r.decorated.Start(brain)
			})
		}
	}
}

// SafeCancelTree 若是主树,取消运行.线程安全
//  @receiver r
//  @param brain
func (r *Root) SafeCancelTree(brain IBrain) {
	if r.IsSubTree(brain) {
		logger.Log.Error("root is subtree,cannot cancel", zap.String("id", r.id))
		return
	}
	thread.GoByID(brain.Blackboard().(IBlackboardInternal).ThreadID(), func() {
		r.Cancel(brain)
	})
}
