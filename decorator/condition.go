package decorator

import (
	"time"

	"github.com/alkaid/behavior/bcore"
)

type IConditionProperties interface {
	IServiceProperties
}

// ConditionProperties 黑板条件节点属性
type ConditionProperties struct {
	bcore.ObservingProperties
	ServiceProperties
}

// Condition 条件装饰器
//
//	将根据配置定时检查委托方法，并根据结果（等于或不等）阻止或允许节点的执行。
type Condition struct {
	bcore.ObservingDecorator
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (c *Condition) PropertiesClassProvider() any {
	return &ConditionProperties{}
}

func (c *Condition) ConditionProperties() IConditionProperties {
	return c.Properties().(IConditionProperties)
}

// StartObserving
//
//	@override bcore.ObservingDecorator .StartObserving
//	@receiver o
//	@param brain
func (c *Condition) StartObserving(brain bcore.IBrain) {
	c.ObservingDecorator.StartObserving(brain)
	interval := c.Properties().(IServiceProperties).GetInterval()
	randomDeviation := c.Properties().(IServiceProperties).GetRandomDeviation()
	if interval <= 0 {
		interval = c.Root(brain).Interval()
	}
	c.StopObserving(brain)
	lastTime := time.Now()
	// 默认投递到黑板保存的线程ID
	c.Memory(brain).CronTask = brain.Cron(interval, randomDeviation, func() {
		currTime := time.Now()
		delta := currTime.Sub(lastTime)
		c.Evaluate(brain, delta)
		lastTime = currTime
	})
}

// StopObserving
//
//	@override bcore.ObservingDecorator .StopObserving
//	@receiver o
//	@param brain
func (c *Condition) StopObserving(brain bcore.IBrain) {
	memory := c.Memory(brain)
	if memory.CronTask != nil {
		memory.CronTask.Stop()
		memory.CronTask = nil
	}
}

// ConditionMet
//
//	@implement bcore.IObservingWorker .ConditionMet
//	@receiver o
//	@param brain
//	@param args
//	@return bool
func (c *Condition) ConditionMet(brain bcore.IBrain, args ...any) bool {
	var delta time.Duration
	if len(args) > 0 {
		delta = args[0].(time.Duration)
	}
	if c.HasDelegatorOrScript() {
		ret := c.Update(brain, bcore.EventTypeOnUpdate, delta)
		return ret == bcore.ResultSucceeded
	}
	c.Log(brain).Error("must set delegator method")
	return false
}
