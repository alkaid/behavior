package decorator

import (
	"github.com/alkaid/behavior/bcore"
	"github.com/alkaid/behavior/thread"
)

type IRepeaterProperties interface {
	GetTimes() int
}

// RepeaterProperties 黑板条件节点属性
type RepeaterProperties struct {
	Times int `json:"times"` // 循环次数 0或负值将永远循环
}

func (r *RepeaterProperties) GetTimes() int {
	return r.Times
}

// Repeater 条件装饰器
//
//	将根据配置定时检查委托方法，并根据结果（等于或不等）阻止或允许节点的执行。
type Repeater struct {
	bcore.Decorator
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (r *Repeater) PropertiesClassProvider() any {
	return &RepeaterProperties{}
}

// OnStart
//
//	@override Node.OnStart
//	@receiver n
//	@param brain
func (r *Repeater) OnStart(brain bcore.IBrain) {
	r.Decorator.OnStart(brain)
	r.Memory(brain).CurrIndex = 0
	r.Decorated(brain).Start(brain)
}

// OnAbort
//
//	@override Node.OnAbort
//	@receiver n
//	@param brain
func (r *Repeater) OnAbort(brain bcore.IBrain) {
	r.Decorator.OnAbort(brain)
}

// OnChildFinished
//
//	@override bcore.Decorator .OnChildFinished
//	@receiver s
//	@param brain
//	@param child
//	@param succeeded
func (r *Repeater) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	r.Decorator.OnChildFinished(brain, child, succeeded)
	if !succeeded {
		r.Finish(brain, false)
		return
	}
	r.Memory(brain).CurrIndex++
	if r.IsAborting(brain) || (r.Properties().(IRepeaterProperties).GetTimes() > 0 && r.Memory(brain).CurrIndex >= r.Properties().(IRepeaterProperties).GetTimes()) {
		r.Finish(brain, true)
		return
	}
	// 不能直接 Finish(),会堆栈溢出且阻塞其他分支,应该重新异步派发
	thread.GoByID(brain.Blackboard().(bcore.IBlackboardInternal).ThreadID(), func() {
		if !r.IsActive(brain) {
			return
		}
		r.Decorated(brain).Start(brain)
	})
}
