package task

import (
	"github.com/alkaid/behavior/internal"
	"time"

	"github.com/alkaid/behavior/bcore"
)

// type IActionProperties interface {
// 	GetIsExeOnAbort() bool
// }
//
// // ActionProperties 子树容器属性
// type ActionProperties struct {
// 	IsExeOnAbort bool `json:"isExeOnAbort"` // 被中断时是否执行一次委托
// }
//
// func (s *ActionProperties) GetIsExeOnAbort() bool {
// 	return s.IsExeOnAbort
// }

type Action struct {
	bcore.Task
}

// // PropertiesClassProvider
// //
// //	@implement INodeWorker.PropertiesClassProvider
// //	@receiver n
// //	@return any
// func (a *Action) PropertiesClassProvider() any {
// 	return &ActionProperties{}
// }
//
// func (a *Action) ActionProperties() IActionProperties {
// 	return a.Properties().(IActionProperties)
// }

// OnStart
//
//	@override Node.OnStart
//	@receiver n
//	@param brain
func (a *Action) OnStart(brain bcore.IBrain) {
	a.Task.OnStart(brain)
	if !a.HasDelegatorOrScript() {
		a.Finish(brain, internal.GlobalConfig.ActionSuccessIfNotDelegate)
		return
	}
	result := a.Update(brain, bcore.EventTypeOnStart, 0)
	if result != bcore.ResultInProgress {
		a.Finish(brain, result == bcore.ResultSucceeded)
		return
	}
	// 按root节点时钟频率定时调用
	interval := a.Root(brain).Interval()
	a.stopTimer(brain)
	lastTime := time.Now()
	// 默认投递到黑板保存的线程ID
	a.Memory(brain).CronTask = brain.Cron(interval, 0, func() {
		if !a.IsActive(brain) {
			return
		}
		currTime := time.Now()
		delta := currTime.Sub(lastTime)
		lastTime = currTime
		result = a.Update(brain, bcore.EventTypeOnUpdate, delta)
		if result != bcore.ResultInProgress {
			a.stopTimer(brain)
			a.Finish(brain, result == bcore.ResultSucceeded)
			return
		}
	})
}

// OnAbort
//
//	@override Node.OnAbort
//	@receiver n
//	@param brain
func (a *Action) OnAbort(brain bcore.IBrain) {
	a.Task.OnAbort(brain)
	a.stopTimer(brain)
	// 中断时最后调用一次委托
	result := a.Update(brain, bcore.EventTypeOnAbort, 0)
	if result == bcore.ResultInProgress {
		a.Log(brain).Error("action executor result cannot be ResultInProgress")
		a.Finish(brain, false)
		return
	}
	a.Finish(brain, result == bcore.ResultSucceeded)
}

func (a *Action) stopTimer(brain bcore.IBrain) {
	if a.Memory(brain).CronTask != nil {
		a.Memory(brain).CronTask.Stop()
		a.Memory(brain).CronTask = nil
	}
}
