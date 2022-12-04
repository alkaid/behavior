package behavior

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/alkaid/behavior/timer"
	"github.com/alkaid/timingwheel"

	"github.com/pkg/errors"

	"github.com/alkaid/behavior/thread"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/alkaid/behavior/handle"
	handle2 "github.com/alkaid/behavior/handle"

	"github.com/alkaid/behavior/bcore"

	"github.com/alkaid/behavior/logger"
)

var _ bcore.IBrain = (*Brain)(nil)

// ExecutorFun 委托方法签名例
type ExecutorFun = func(eventType bcore.EventType, delta time.Duration) bcore.Result

type Brain struct {
	blackboard    bcore.IBlackboardInternal
	delegatesMeta map[string]*bcore.DelegateMeta
	finishChan    chan *bcore.FinishEvent // 供上层业务方使用的完成通知
	root          bcore.IRoot
	ctx           context.Context
}

func (b *Brain) ID() int {
	return b.blackboard.ThreadID()
}

func (b *Brain) SetFinishChan(finishChan chan *bcore.FinishEvent) {
	b.finishChan = finishChan
}

func (b *Brain) RWFinishChan() chan *bcore.FinishEvent {
	return b.finishChan
}
func (b *Brain) FinishChan() <-chan *bcore.FinishEvent {
	return b.finishChan
}
func (b *Brain) RunningTree() bcore.IRoot {
	return b.root
}
func (b *Brain) Running() bool {
	return b.root != nil
}
func (b *Brain) SetRunningTree(root bcore.IRoot) {
	b.root = root
}
func (b *Brain) Context() context.Context {
	return b.ctx
}
func (b *Brain) SetContext(ctx context.Context) {
	b.ctx = ctx
}

// NewBrain bcore.IBrain 实例
//
//	@param blackboard
//	@param delegates 要注册的委托对象
//	@return bcore.IBrain
func NewBrain(blackboard bcore.IBlackboard, delegates map[string]any, finishChan chan *bcore.FinishEvent) bcore.IBrain {
	b := &Brain{
		blackboard:    blackboard.(bcore.IBlackboardInternal),
		delegatesMeta: map[string]*bcore.DelegateMeta{},
	}
	b.SetDelegates(delegates)
	b.finishChan = finishChan
	b.ctx = context.Background()
	return b
}

func (b *Brain) Blackboard() bcore.IBlackboard {
	return b.blackboard
}

// RegisterDelegate 注册委托对象
//
//	@receiver b
//	@param name
//	@param delegate
func (b *Brain) RegisterDelegate(name string, delegate any) {
	if delegate == nil {
		logger.Log.Fatal("delegate can't be nil")
	}
	b.delegatesMeta[name] = &bcore.DelegateMeta{
		Delegate:     delegate,
		ReflectValue: reflect.ValueOf(delegate),
	}
}

// SetDelegates 注册委托对象
//
//	@receiver b
//	@param delegatesMeta
func (b *Brain) SetDelegates(delegates map[string]any) {
	for name, d := range delegates {
		b.RegisterDelegate(name, d)
	}
}

// GetDelegates 获取委托map拷贝
//
//	@receiver b
//	@return map[string]any
func (b *Brain) GetDelegates() map[string]any {
	delegates := map[string]any{}
	for name, meta := range b.delegatesMeta {
		delegates[name] = meta.Delegate
	}
	return delegates
}
func (b *Brain) GetDelegate(name string) (delegate any, ok bool) {
	meta, ok := b.delegatesMeta[name]
	if ok {
		return meta.Delegate, ok
	}
	return nil, false
}

func (b *Brain) Go(task func()) {
	thread.GoByID(b.blackboard.ThreadID(), task)
}

// Abort @implement bcore.IBrain .Abort
//
// @receiver b
// @param abortChan
func (b *Brain) Abort(abortChan chan *bcore.FinishEvent) {
	// 派发到自己的线程
	b.Go(func() {
		event := &bcore.FinishEvent{
			ID:        b.ID(),
			IsAbort:   true,
			Succeeded: false,
			IsActive:  true,
		}
		if !b.Running() || !b.RunningTree().IsActive(b) {
			logger.Log.Warn("brain not running,can not abort")
			if abortChan != nil {
				event.IsActive = false
				abortChan <- event
			}
			return
		}
		b.RunningTree().SetUpstream(b, nil)
		b.RunningTree().Abort(b)
		if abortChan != nil {
			abortChan <- event
		}
	})
}
func (b *Brain) Run(tag string, force bool) {
	b.Go(func() {
		if b.Running() && b.RunningTree().IsActive(b) {
			if force {
				b.RunningTree().SetUpstream(b, nil)
				b.RunningTree().Abort(b)
			} else {
				return
			}
		} else {
			tree := GlobalTreeRegistry().GetNotParentTreeWithoutClone(tag)
			if tree == nil || tree.Root == nil {
				logger.Log.Warn("can not find main tree for tag", zap.String("tag", tag))
			}
			tree.Root.Start(b)
		}
	})
}

// DynamicDecorate 给正在运行的树动态挂载子树.暂时只支持主树上的动态容器
//
//	非线程安全,调用方自己保证
//
// @receiver b
// @param containerTag 动态子树容器的tag
// @param subtreeTag 子树的tag
// @return error
func (b *Brain) DynamicDecorate(containerTag string, subtreeTag string) error {
	if !b.Running() || !b.RunningTree().IsActive(b) {
		return errors.New(fmt.Sprintf("brain can not dynamic decorate cause not running tree,containerTag=%s,subtreeTag=%s", containerTag, subtreeTag))
	}
	registry := GlobalTreeRegistry()
	maintree := registry.TreesByID[b.RunningTree().ID()]
	if maintree == nil {
		return errors.New(fmt.Sprintf("brain can not dynamic decorate cause not main tree,containerTag=%s,subtreeTag=%s", containerTag, subtreeTag))
	}
	container := maintree.DynamicSubtrees[containerTag]
	if container == nil {
		return errors.New(fmt.Sprintf("brain can not dynamic decorate cause not dynamic container,containerTag=%s,subtreeTag=%s", containerTag, subtreeTag))
	}
	// 当前子树就是想要挂载的子树,不再执行动态替换
	childRoot := container.Decorated(b)
	if childRoot != nil && registry.TreesByID[childRoot.ID()].Tag == subtreeTag {
		return nil
	}
	subtree, _, err := registry.getNotDynamicParentTree(subtreeTag, container, b)
	if err != nil {
		return err
	}
	if subtree == nil {
		return errors.New(fmt.Sprintf("brain can not dynamic decorate cause not enough subtree,containerTag=%s,subtreeTag=%s", containerTag, subtreeTag))
	}

	container.DynamicDecorate(b, subtree.Root)
	return nil
}

// OnNodeUpdate 供节点回调执行委托 会在 Brain 的独立线程里运行
//
//	@receiver b
//	@param target
//	@param method
//	@param brain
//	@param eventType
//	@param delta
//	@return bcore.Result
func (b *Brain) OnNodeUpdate(target string, method string, brain bcore.IBrain, eventType bcore.EventType, delta time.Duration) bcore.Result {
	log := logger.Log.With(zap.String("target", target), zap.String("method", method), zap.Int("eventType", int(eventType)))
	meta := b.delegatesMeta[target]
	if meta == nil {
		log.Error("target is nil,please register delegate before run behavior tree")
		return bcore.ResultFailed
	}
	handler := GlobalHandlerPool().GetHandle(target, method)
	if handler == nil || handler.MethodType == handle2.MtNone {
		log.Error("handler is nil,please register target to GlobalHandlerPool() before run behavior tree")
		return bcore.ResultFailed
	}
	var rets []any
	var err error
	log = log.With(zap.Int("methodType", int(handler.MethodType)))
	switch handler.MethodType {
	case handle.MtFullStyle:
		_, rets, err = GlobalHandlerPool().ProcessHandler(handler, meta.ReflectValue, eventType, delta)
	default:
		_, rets, err = GlobalHandlerPool().ProcessHandler(handler, meta.ReflectValue)
	}
	// 出错默认返回失败
	if err != nil {
		log.Error("handler reflect method call error", zap.Error(err))
		return bcore.ResultFailed
	}
	// 出参只能0个或1个
	switch len(rets) {
	case 0:
		return bcore.ResultSucceeded
	case 1:
		// 1个时判断是error还是result
		if rets[0] == nil {
			return bcore.ResultSucceeded
		}
		if err, ok := rets[0].(error); ok {
			// 不打印堆栈,堆栈由上层业务方打印
			log.WithOptions(zap.AddStacktrace(zapcore.FatalLevel)).Error("delegator method return error", zap.Error(err))
			return bcore.ResultFailed
		}
		if result, ok := rets[0].(bcore.Result); ok {
			return result
		}
		// 出参类型超限
		log.Error("delegator method return type illegal")
		return bcore.ResultFailed
	}
	// 出参数量超限
	log.Error("delegator method return value's number illegal")
	return bcore.ResultFailed
}

// Cron wrap timingwheel.TimingWheel .Cron
//
//	@param interval 间隔
//	@param randomDeviation 随机方差范围,interval=interval+randomDeviation*[-0.5,0.5)
//	@param task
//	@param opts
func (b *Brain) Cron(interval time.Duration, randomDeviation time.Duration, task func()) *timingwheel.Timer {
	return timer.Cron(interval, randomDeviation, task, timingwheel.WithGoID(b.blackboard.ThreadID()), timingwheel.WithPool(thread.PoolInstance()))
}

// After wrap timingwheel.TimingWheel .AfterFunc
//
//	@param interval 间隔
//	@param randomDeviation 随机方差范围 interval = interval + randomDeviation*[-0.5,0.5)
//	@param task
//	@param opts
func (b *Brain) After(interval time.Duration, randomDeviation time.Duration, task func()) *timingwheel.Timer {
	return timer.After(interval, randomDeviation, task, timingwheel.WithGoID(b.blackboard.ThreadID()), timingwheel.WithPool(thread.PoolInstance()))
}
