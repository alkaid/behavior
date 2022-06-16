package bcore

import (
	"time"

	"github.com/alkaid/behavior/util"

	"github.com/alkaid/behavior/thread"
	"github.com/samber/lo"

	"go.uber.org/zap"
)

type iRootProperties interface {
	IsOnce() bool
	GetInterval() time.Duration
}

// rootProperties 根节点属性
type rootProperties struct {
	BaseProperties
	Once     bool          `json:"once"`     // 是否仅运行一次,反之永远循环
	Interval util.Duration `json:"interval"` // 默认更新间隔
}

func (r *rootProperties) IsOnce() bool {
	return r.Once
}
func (r *rootProperties) GetInterval() time.Duration {
	return r.Interval.Duration
}

type IRoot interface {
	IDecorator
	// IsSubTree 是否父节点
	//  @param brain
	//  @return bool
	IsSubTree(brain IBrain) bool
	// Interval 该树默认刷新间隔时间
	//  @return time.Duration
	Interval() time.Duration
}

var _ IRoot = (*Root)(nil)

type Root struct {
	Decorator
}

// CanMounted
//  @override Node.CanMounted
//  @receiver r
//  @return bool
func (r *Root) CanMounted() bool {
	return true
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

func (r *Root) Interval() time.Duration {
	interval := r.properties.(iRootProperties).GetInterval()
	return lo.If(interval <= 0, DefaultInterval).Else(interval)
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (r *Root) PropertiesClassProvider() any {
	return &rootProperties{}
}

// SetRoot
//  @override Decorator.SetRoot
//  @receiver n
//  @param root
func (r *Root) SetRoot(root IRoot) {
	if root.ID() != r.ID() {
		r.Log().Fatal("root's root must be self", zap.String("selfID", r.ID()), zap.String("inID", root.ID()))
	}
	r.Decorator.SetRoot(root)
}

// OnStart
//  @override Node.OnStart
//  @receiver n
//  @param brain
func (r *Root) OnStart(brain IBrain) {
	// 开启子节点,保证单线程执行
	thread.GoByID(brain.Blackboard().(IBlackboardInternal).ThreadID(), func() {
		r.Decorator.OnStart(brain)
		// 非子树要开启黑板监听
		if !r.IsSubTree(brain) {
			brain.Blackboard().(IBlackboardInternal).Start()
		}
		r.Decorated(brain).Start(brain)
	})
}

// OnAbort
//  @override Node.OnAbort
//  @receiver r
//  @param brain
func (r *Root) OnAbort(brain IBrain) {
	thread.GoByID(brain.Blackboard().(IBlackboardInternal).ThreadID(), func() {
		r.Decorator.OnAbort(brain)
		if r.IsActive(brain) {
			r.Decorated(brain).Abort(brain)
			return
		}
		// TODO 这里是否有异步异常情况未考虑
		r.Log().Warn("can only abort active root")
	})
}

// OnChildFinished
//  @override Container.OnChildFinished
//  @receiver r
//  @param brain
//  @param child
//  @param succeeded
func (r *Root) OnChildFinished(brain IBrain, child INode, succeeded bool) {
	r.Decorator.OnChildFinished(brain, child, succeeded)
	// 如果是外部触发中断的,结束运行
	if r.Memory(brain).State == NodeStateAborting {
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
		} else if r.properties.(iRootProperties).IsOnce() {
			// 若是主树且是一次性,结束root并关闭监听
			brain.Blackboard().(IBlackboardInternal).Stop()
			r.Finish(brain, succeeded)
		} else {
			// 若是主树且可以循环，开启下一轮
			// 不能直接 Start(),会堆栈溢出且阻塞其他分支,应该重新异步派发
			thread.GoByID(brain.Blackboard().(IBlackboardInternal).ThreadID(), func() {
				r.Decorated(brain).Start(brain)
			})
		}
	}
}

// SafeAbortTree 若是主树,中断运行. 是线程安全的
//  @receiver r
//  @param brain
func (r *Root) SafeAbortTree(brain IBrain) {
	thread.GoByID(brain.Blackboard().(IBlackboardInternal).ThreadID(), func() {
		if r.IsSubTree(brain) {
			r.Log().Error("root is subtree,cannot abort", zap.String("id", r.id))
			return
		}
		r.Abort(brain)
	})
}
