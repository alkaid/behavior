package decorator

import (
	"time"

	"github.com/alkaid/behavior/util"

	"github.com/alkaid/behavior/bcore"
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
	RandomDeviation     util.Duration `json:"randomDeviation"`     // 随机离差:冷却时间 ICooldownBaseWorker.CooldownTime = CooldownTime + RandomDeviation*[-0.5,0.5)
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
	return p.RandomDeviation.Duration
}

// CooldownBase cd等待装饰器
//
//	每次子节点完成后将等待一段时间才能再次执行
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
//
//	@override Node.InitNodeWorker
//	@receiver c
//	@param worker
func (b *CooldownBase) InitNodeWorker(worker bcore.INodeWorker) error {
	b.ICooldownBaseWorker = worker.(ICooldownBaseWorker)
	return b.Decorator.InitNodeWorker(worker)
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (b *CooldownBase) PropertiesClassProvider() any {
	return &CooldownBaseProperties{}
}

func (b *CooldownBase) CooldownProperties() ICooldownBaseProperties {
	return b.Properties().(ICooldownBaseProperties)
}

// OnStart
//
//	@override Node.OnStart
//	@receiver n
//	@param brain
func (b *CooldownBase) OnStart(brain bcore.IBrain) {
	b.Decorator.OnStart(brain)
	if !b.Memory(brain).Cooling {
		b.Memory(brain).Cooling = true
		if !b.CooldownProperties().GetStartAfterDecorated() {
			b.startTimer(brain)
		}
		b.Decorated(brain).Start(brain)
		return
	}
	if b.CooldownProperties().GetFailOnCoolDown() {
		b.Finish(brain, false)
	}
}

// OnAbort
//
//	@override Node.OnAbort
//	@receiver n
//	@param brain
func (b *CooldownBase) OnAbort(brain bcore.IBrain) {
	b.Memory(brain).Cooling = false
	b.stopTimer(brain)
	b.Decorator.OnAbort(brain)
}

// OnChildFinished
//
//	@override bcore.Decorator .OnChildFinished
//	@receiver s
//	@param brain
//	@param child
//	@param succeeded
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
		if b.IsActive(brain) && !b.Decorated(brain).IsActive(brain) {
			b.startTimer(brain)
			b.Decorated(brain).Start(brain)
		} else {
			b.Memory(brain).Cooling = false
		}
	}
}
func (b *CooldownBase) startTimer(brain bcore.IBrain) {
	b.Memory(brain).CronTask = brain.After(b.CooldownTime(brain),
		b.CooldownProperties().GetRandomDeviation(),
		b.getTaskFun(brain))
}
func (b *CooldownBase) stopTimer(brain bcore.IBrain) {
	if b.Memory(brain).CronTask != nil {
		b.Memory(brain).CronTask.Stop()
		b.Memory(brain).CronTask = nil
	}
}
