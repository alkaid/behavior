package decorator

import (
	"time"

	"github.com/alkaid/behavior/util"

	"github.com/alkaid/behavior/bcore"
)

type ITimeMaxProperties interface {
	GetLimit() time.Duration
	GetRandomDeviation() time.Duration
	GetWaitForChildButFail() bool
}

// TimeMaxProperties 最大时限装饰器属性
type TimeMaxProperties struct {
	Limit               util.Duration `json:"limit"`               // 最大时限
	RandomDeviation     util.Duration `json:"randomDeviation"`     // 随机离差:将一个随机范围数值添加至服务节点的 Limit 值。Limit = Limit + RandomDeviation * [-0.5,0.5)
	WaitForChildButFail bool          `json:"waitForChildButFail"` // true:超时后依然等待子节点完成,但将修改结果为失败 false:超时后立即关闭子节点
}

func (t TimeMaxProperties) GetLimit() time.Duration {
	return t.Limit.Duration
}

func (t TimeMaxProperties) GetRandomDeviation() time.Duration {
	return t.RandomDeviation.Duration
}
func (t TimeMaxProperties) GetWaitForChildButFail() bool {
	return t.WaitForChildButFail
}

// TimeMax 最大时限装饰器
//
//	限制最大时限内必须返回结果
type TimeMax struct {
	bcore.Decorator
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (m *TimeMax) PropertiesClassProvider() any {
	return &TimeMaxProperties{}
}
func (m *TimeMax) TimeMaxProperties() ITimeMaxProperties {
	return m.Properties().(ITimeMaxProperties)
}

// OnStart
//
//	@override Node.OnStart
//	@receiver n
//	@param brain
func (m *TimeMax) OnStart(brain bcore.IBrain) {
	m.Decorator.OnStart(brain)
	m.Memory(brain).LimitReached = false
	m.startTimer(brain)
	m.Decorated(brain).Start(brain)
}

// OnAbort
//
//	@override Node.OnAbort
//	@receiver n
//	@param brain
func (m *TimeMax) OnAbort(brain bcore.IBrain) {
	m.stopTimer(brain)
	m.Decorator.OnAbort(brain)
}

// OnChildFinished
//
//	@override bcore.Decorator .OnChildFinished
//	@receiver s
//	@param brain
//	@param child
//	@param succeeded
func (m *TimeMax) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	m.Decorator.OnChildFinished(brain, child, succeeded)
	m.stopTimer(brain)
	if m.Memory(brain).LimitReached {
		m.Finish(brain, false)
	} else {
		m.Finish(brain, succeeded)
	}
}

func (b *TimeMax) getTaskFun(brain bcore.IBrain) func() {
	return func() {
		if !b.IsActive(brain) {
			return
		}
		if !b.TimeMaxProperties().GetWaitForChildButFail() {
			b.Decorated(brain).SetUpstream(brain, b)
			b.Decorated(brain).Abort(brain)
		} else {
			b.Memory(brain).LimitReached = true
			if !b.Decorated(brain).IsActive(brain) {
				b.Log(brain).Error("decorated must be active")
				return
			}
		}
	}
}
func (b *TimeMax) startTimer(brain bcore.IBrain) {
	b.Memory(brain).CronTask = brain.After(b.TimeMaxProperties().GetLimit(),
		b.TimeMaxProperties().GetRandomDeviation(),
		b.getTaskFun(brain))
}
func (b *TimeMax) stopTimer(brain bcore.IBrain) {
	if b.Memory(brain).CronTask != nil {
		b.Memory(brain).CronTask.Stop()
		b.Memory(brain).CronTask = nil
	}
}
