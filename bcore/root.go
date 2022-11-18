package bcore

import (
	"time"

	"github.com/alkaid/behavior/logger"

	"github.com/alkaid/behavior/util"

	"github.com/alkaid/behavior/thread"
	"github.com/samber/lo"

	"go.uber.org/zap"
)

type iRootProperties interface {
	IsOnce() bool
	GetInterval() time.Duration
	GetLoopInterval() time.Duration
	GetRandomDeviation() time.Duration
}

// rootProperties 根节点属性
type rootProperties struct {
	BaseProperties
	Once            bool          `json:"once"`            // 是否仅运行一次,反之永远循环
	Interval        util.Duration `json:"interval"`        // 默认帧率(每帧更新间隔,默认50ms)
	LoopInterval    util.Duration `json:"loopInterval"`    // 每次运行之间的间隔
	RandomDeviation util.Duration `json:"randomDeviation"` // 每次运行随机偏差: LoopInterval = LoopInterval + RandomDeviation*[-0.5,0.5)
}

func (r *rootProperties) IsOnce() bool {
	return r.Once
}
func (r *rootProperties) GetInterval() time.Duration {
	return r.Interval.Duration
}
func (r *rootProperties) GetLoopInterval() time.Duration {
	return r.LoopInterval.Duration
}
func (r *rootProperties) GetRandomDeviation() time.Duration {
	return r.RandomDeviation.Duration
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
	// SafeStart  启动行为树.
	//
	//		线程安全
	// @receiver r
	// @param brain
	// SafeStart  启动行为树.
	//
	// @param brain
	// @param force 是否强制启动.若root是激活状态,true则先终止再启动,false则原tree继续运行只报错不重启.
	SafeStart(brain IBrain, force bool)
	// SafeAbort 终止行为树
	//
	//		线程安全
	// @param brain
	// @param abortChan
	SafeAbort(brain IBrain, abortChan chan *FinishEvent)
}

var _ IRoot = (*Root)(nil)

type Root struct {
	Decorator
}

// CanMountTo
//
//	@override Node.CanMountTo
//	@receiver r
//	@return bool
func (r *Root) CanMountTo() bool {
	return true
}

// IsSubTree
//
//	@implement IRoot.IsSubTree
//	@receiver r
//	@param brain
//	@return bool
func (r *Root) IsSubTree(brain IBrain) bool {
	// 父节点不为空就是子树
	return r.Parent(brain) != nil
}

func (r *Root) Interval() time.Duration {
	interval := r.properties.(iRootProperties).GetInterval()
	return lo.If(interval <= 0, DefaultInterval).Else(interval)
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (r *Root) PropertiesClassProvider() any {
	return &rootProperties{}
}

// SetRoot
//
//	@override Decorator.SetRoot
//	@receiver n
//	@param root
func (r *Root) SetRoot(brain IBrain, root IRoot) {
	if root.ID() != r.ID() {
		r.Log(brain).Fatal("root's root must be self", zap.String("selfID", r.ID()), zap.String("inID", root.ID()))
		return
	}
	r.Decorator.SetRoot(brain, root)
}

// Start 启动行为树.
//
//		非线程安全,要线程安全请使用 SafeStart
//	 @implement INode.Start
//	 @receiver n
//	 @param brain
func (r *Root) Start(brain IBrain) {
	if r.IsSubTree(brain) {
		r.Decorator.Start(brain)
		return
	}
	brain.(IBrainInternal).SetRunningTree(r)
	r.Decorator.Start(brain)
}

// SafeStart
//
// @implement IRoot.SafeStart
// @receiver r
// @param brain
func (r *Root) SafeStart(brain IBrain, force bool) {
	if r.IsSubTree(brain) {
		logger.Log.Error("SafeStart unSupport subtree")
		return
	}
	thread.GoByID(brain.Blackboard().(IBlackboardInternal).ThreadID(), func() {
		if force && r.IsActive(brain) {
			r.SetUpstream(brain, nil)
			r.Abort(brain)
		}
		r.Start(brain)
	})
}

// OnStart
//
//	@override Node.OnStart
//	@receiver n
//	@param brain
func (r *Root) OnStart(brain IBrain) {
	r.Decorator.OnStart(brain)
	// 非子树要开启黑板监听
	if !r.IsSubTree(brain) {
		brain.Blackboard().(IBlackboardInternal).Start()
	}
	r.Decorated(brain).Start(brain)
}

// SafeAbort
//
// @implement IRoot.SafeAbort
// @receiver r
// @param brain
// @param abortChan
func (r *Root) SafeAbort(brain IBrain, abortChan chan *FinishEvent) {
	if r.IsSubTree(brain) {
		logger.Log.Error("SafeAbort unSupport subtree")
		return
	}
	thread.GoByID(brain.Blackboard().(IBlackboardInternal).ThreadID(), func() {
		event := &FinishEvent{
			IsAbort:   true,
			Succeeded: false,
			IsActive:  true,
		}
		if !r.IsActive(brain) {
			if abortChan != nil {
				event.IsActive = false
				abortChan <- event
			}
			return
		}
		r.Decorator.SetUpstream(brain, r)
		r.Decorator.Abort(brain)
		if abortChan != nil {
			abortChan <- event
		}
	})
}

// OnAbort
//
//	@override Node.OnAbort
//	@receiver r
//	@param brain
func (r *Root) OnAbort(brain IBrain) {
	r.Decorator.OnAbort(brain)
	r.Decorated(brain).SetUpstream(brain, r)
	r.Decorated(brain).Abort(brain)
}

// OnChildFinished
//
//	@override Container.OnChildFinished
//	@receiver r
//	@param brain
//	@param child
//	@param succeeded
func (r *Root) OnChildFinished(brain IBrain, child INode, succeeded bool) {
	r.Decorator.OnChildFinished(brain, child, succeeded)
	r.stopTimer(brain)
	// 如果是外部触发中断的,结束运行
	if r.Memory(brain).State == NodeStateAborting {
		// 无论是否子树都要结束root,若是子树将回溯parent,否则整棵行为树终止运行。
		r.Finish(brain, succeeded)
		// 非子树要关闭黑板监听
		if !r.IsSubTree(brain) {
			brain.Blackboard().(IBlackboardInternal).Stop()
		}
	} else {
		// 如果是正常结束，判断是否子树,子树则正常结束当前root
		// 如果配置了一次性运行,同上
		if r.IsSubTree(brain) {
			r.Finish(brain, succeeded)
		} else if r.properties.(iRootProperties).IsOnce() {
			// 若是主树且是一次性,结束root并关闭监听
			r.Finish(brain, succeeded)
			brain.Blackboard().(IBlackboardInternal).Stop()
		} else {
			// 若是主树且可以循环，开启下一轮
			// 不能直接 Start(),会堆栈溢出且阻塞其他分支,应该重新异步派发
			r.startTimer(brain)
		}
	}
}

// Finish
//
//	@override Node.Finish
//	@receiver d
//	@param brain
//	@param succeeded
func (r *Root) Finish(brain IBrain, succeeded bool) {
	isAborting := r.IsAborting(brain)
	r.Decorator.Finish(brain, succeeded)
	if r.IsSubTree(brain) {
		return
	}
	// 若是主树 通知brain运行完成
	brain.(IBrainInternal).SetRunningTree(nil)
	event := &FinishEvent{
		IsAbort:   isAborting,
		Succeeded: succeeded,
		IsActive:  true,
	}
	if brain.FinishChan() != nil {
		brain.(IBrainInternal).RWFinishChan() <- event
	}
}

func (r *Root) startTimer(brain IBrain) {
	r.Memory(brain).CronTask = brain.After(r.properties.(iRootProperties).GetLoopInterval(),
		r.properties.(iRootProperties).GetRandomDeviation(),
		func() {
			if r.IsActive(brain) {
				r.Decorated(brain).Start(brain)
			}
		})
}
func (r *Root) stopTimer(brain IBrain) {
	if r.Memory(brain).CronTask != nil {
		r.Memory(brain).CronTask.Stop()
		r.Memory(brain).CronTask = nil
	}
}
