package decorator

import (
	"time"

	"github.com/alkaid/behavior/bcore"
	"github.com/alkaid/behavior/timer"
	"github.com/alkaid/timingwheel"
)

type ICooldownBaseProperties interface {
	GetStartAfterDecorated() bool
	GetResetOnFailure() bool
	GetFailOnCoolDown() bool
	GetRandomDeviation() time.Duration
}

// CooldownBaseProperties cd等待装饰器属性
type CooldownBaseProperties struct {
	StartAfterDecorated bool          `json:"startAfterDecorated"` // 是否在子节点完成后才进入cd
	ResetOnFailure      bool          `json:"resetOnFailure"`      // 子节点取消后是否重置cd
	FailOnCoolDown      bool          `json:"failOnCoolDown"`      // true:cd时间未到却被请求执行时停用当前节点并返回false false:什么都不做
	RandomDeviation     time.Duration `json:"randomDeviation"`     // 随机偏差:将一个随机范围数值添加至服务节点的 冷却时间（cooldownTime） 值。
}

func (p *CooldownBaseProperties) GetStartAfterDecorated() bool {
	return p.StartAfterDecorated
}

func (p *CooldownBaseProperties) GetResetOnFailure() bool {
	return p.ResetOnFailure
}

func (p *CooldownBaseProperties) GetFailOnCoolDown() bool {
	return p.FailOnCoolDown
}

func (p *CooldownBaseProperties) GetRandomDeviation() time.Duration {
	return p.RandomDeviation
}

// CooldownBase cd等待装饰器
//  每次子节点完成后将等待一段时间才能再次执行
type CooldownBase struct {
	bcore.Decorator
	ICooldownBaseWorker
}
type ICooldownBaseWorker interface {
	// CooldownTime 获取冷却时间
	//  @param brain
	//  @return time.Duration
	CooldownTime(brain bcore.IBrain) time.Duration
}

// InitNodeWorker
//  @override Node.InitNodeWorker
//  @receiver c
//  @param worker
func (b *CooldownBase) InitNodeWorker(worker bcore.INodeWorker) {
	b.Decorator.InitNodeWorker(worker)
	b.ICooldownBaseWorker = worker.(ICooldownBaseWorker)
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (b *CooldownBase) PropertiesClassProvider() any {
	return &CooldownBaseProperties{}
}

func (b *CooldownBase) CooldownProperties() ICooldownBaseProperties {
	return b.Properties().(ICooldownBaseProperties)
}

// OnStart
//  @override Node.OnStart
//  @receiver n
//  @param brain
func (b *CooldownBase) OnStart(brain bcore.IBrain) {
	b.Decorator.OnStart(brain)
	if !b.Memory(brain).Cooling {
		b.Memory(brain).Cooling = true
		if !b.CooldownProperties().GetStartAfterDecorated() {
			b.startTimer(brain)
		}
		b.Decorated().Start(brain)
		return
	}
	if b.CooldownProperties().GetFailOnCoolDown() {
		b.Finish(brain, false)
	}
}

// OnAbort
//  @override Node.OnAbort
//  @receiver n
//  @param brain
func (b *CooldownBase) OnAbort(brain bcore.IBrain) {
	b.Decorator.OnAbort(brain)
	b.Memory(brain).Cooling = false
	b.stopTimer(brain)
	if b.Decorated().IsActive(brain) {
		b.Decorated().Abort(brain)
	} else {
		b.Finish(brain, false)
	}
}

// OnChildFinished
//  @override bcore.Decorator .OnChildFinished
//  @receiver s
//  @param brain
//  @param child
//  @param succeeded
func (b *CooldownBase) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	b.Decorator.OnChildFinished(brain, child, succeeded)
	if !succeeded && b.CooldownProperties().GetResetOnFailure() {
		b.Memory(brain).Cooling = false
		b.stopTimer(brain)
	} else if b.CooldownProperties().GetStartAfterDecorated() {
		b.startTimer(brain)
	}
	b.Finish(brain, succeeded)
}

func (b *CooldownBase) getTaskFun(brain bcore.IBrain) func() {
	return func() {
		if b.IsActive(brain) && !b.Decorated().IsActive(brain) {
			b.startTimer(brain)
			b.Decorated().Start(brain)
		} else {
			b.Memory(brain).Cooling = false
		}
	}
}
func (b *CooldownBase) startTimer(brain bcore.IBrain) {
	b.Memory(brain).CronTask = timer.After(b.CooldownTime(brain),
		b.CooldownProperties().GetRandomDeviation(),
		b.getTaskFun(brain),
		timingwheel.WithGoID(brain.Blackboard().(bcore.IBlackboardInternal).ThreadID()))
}
func (b *CooldownBase) stopTimer(brain bcore.IBrain) {
	if b.Memory(brain).CronTask != nil {
		b.Memory(brain).CronTask.Stop()
		b.Memory(brain).CronTask = nil
	}
}
