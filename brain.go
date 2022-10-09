package behavior

import (
	"reflect"
	"time"

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
	blackboard    bcore.IBlackboard
	delegatesMeta map[string]*bcore.DelegateMeta
	finishChan    chan *bcore.FinishEvent // 供上层业务方使用的完成通知
	root          bcore.IRoot
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

// NewBrain bcore.IBrain 实例
//
//	@param blackboard
//	@param delegates 要注册的委托对象
//	@return bcore.IBrain
func NewBrain(blackboard bcore.IBlackboard, delegates map[string]any, finishChan chan *bcore.FinishEvent) bcore.IBrain {
	b := &Brain{
		blackboard:    blackboard,
		delegatesMeta: map[string]*bcore.DelegateMeta{},
	}
	b.SetDelegates(delegates)
	b.finishChan = finishChan
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
	thread.GoByID(b.blackboard.(bcore.IBlackboardInternal).ThreadID(), task)
}

func (b *Brain) ForceRun(root bcore.IRoot) {
	notEnableChan := make(chan struct{})
	restartChan := make(chan *bcore.FinishEvent)
	originFinishChan := b.finishChan
	thread.Go(func() {
		for {
			select {
			case <-notEnableChan:
				return
			case e := <-restartChan:
				b.finishChan = originFinishChan
				if b.finishChan != nil {
					b.finishChan <- e
				}
				root.Start(b)
				return
			}
		}
	})
	thread.GoByID(b.Blackboard().(bcore.IBlackboardInternal).ThreadID(), func() {
		if !b.Running() {
			root.Start(b)
			notEnableChan <- struct{}{}
			return
		}
		if b.RunningTree().IsInactive(b) {
			logger.Log.Error("brain's tree state error,cannot be inactive")
			notEnableChan <- struct{}{}
			return
		}
		b.finishChan = restartChan
		if b.RunningTree().IsActive(b) {
			// tree finish时会通知 brain.finishChan
			b.RunningTree().Abort(b)
		}
	})
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
		_, rets, err = GlobalHandlerPool().ProcessHandler(handler, meta.ReflectValue, brain, eventType, delta)
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
