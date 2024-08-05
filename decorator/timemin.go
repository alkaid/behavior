package decorator

import (
	"time"

	"github.com/alkaid/behavior/bcore"
	"github.com/alkaid/behavior/util"
)

type ITimeMinProperties interface {
	GetLimit() time.Duration
	GetRandomDeviation() time.Duration
	GetFinishOnChildFailure() bool
}

// TimeMinProperties 最小时限装饰器属性
type TimeMinProperties struct {
	Limit                util.Duration `json:"Limit"`                // 最小时限
	RandomDeviation      util.Duration `json:"randomDeviation"`      // 随机离差:将一个随机范围数值添加至服务节点的 Limit 值。Limit = Limit + RandomDeviation * [-0.5,0.5)
	FinishOnChildFailure bool          `json:"finishOnChildFailure"` // true:子节点返回false时，当前节点会立即停用并返回false false:子节点返回时，当前节点会等到直到达到时间限制后才停用
}

func (t TimeMinProperties) GetLimit() time.Duration {
	return t.Limit.Duration
}

func (t TimeMinProperties) GetRandomDeviation() time.Duration {
	return t.RandomDeviation.Duration
}
func (t TimeMinProperties) GetFinishOnChildFailure() bool {
	return t.FinishOnChildFailure
}

// TimeMin 最小时限装饰器
//
//	限制子节点至少须执行xx时限
type TimeMin struct {
	bcore.Decorator
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (m *TimeMin) PropertiesClassProvider() any {
	return &TimeMinProperties{}
}
func (m *TimeMin) TimeMinProperties() ITimeMinProperties {
	return m.Properties().(ITimeMinProperties)
}

// OnStart
//
//	@override Node.OnStart
//	@receiver n
//	@param brain
func (m *TimeMin) OnStart(brain bcore.IBrain) {
	m.Decorator.OnStart(brain)
	m.Memory(brain).LimitReached = false
	m.Memory(brain).DecoratedDone = false
	m.Memory(brain).DecoratedSuccess = false
	m.startTimer(brain)
	m.Decorated(brain).Start(brain)
}

// OnAbort
//
//	@override Node.OnAbort
//	@receiver n
//	@param brain
func (m *TimeMin) OnAbort(brain bcore.IBrain) {
	m.stopTimer(brain)
	m.Memory(brain).LimitReached = true
	m.Decorator.OnAbort(brain)
}

// OnChildFinished
//
//	@override bcore.Decorator .OnChildFinished
//	@receiver s
//	@param brain
//	@param child
//	@param succeeded
func (m *TimeMin) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	m.Decorator.OnChildFinished(brain, child, succeeded)
	m.Memory(brain).DecoratedDone = true
	m.Memory(brain).DecoratedSuccess = succeeded
	if m.Memory(brain).LimitReached || (!succeeded && m.TimeMinProperties().GetFinishOnChildFailure()) {
		m.stopTimer(brain)
		m.Finish(brain, succeeded)
		return
	}
	if m.Memory(brain).CronTask == nil {
		m.Log(brain).Error("timer task cannot be nil")
		return
	}
}

func (b *TimeMin) getTaskFun(brain bcore.IBrain) func() {
	return func() {
		if !b.IsActive(brain) {
			return
		}
		b.Memory(brain).LimitReached = true
		if b.Memory(brain).DecoratedDone {
			b.Finish(brain, b.Memory(brain).DecoratedSuccess)
			return
		}
		if !b.Decorated(brain).IsActive(brain) {
			b.Log(brain).Error("decorated must be active")
			return
		}
	}
}
func (b *TimeMin) startTimer(brain bcore.IBrain) {
	b.Memory(brain).CronTask = brain.After(b.TimeMinProperties().GetLimit(),
		b.TimeMinProperties().GetRandomDeviation(),
		b.getTaskFun(brain))
}
func (b *TimeMin) stopTimer(brain bcore.IBrain) {
	if b.Memory(brain).CronTask != nil {
		b.Memory(brain).CronTask.Stop()
		b.Memory(brain).CronTask = nil
	}
}
