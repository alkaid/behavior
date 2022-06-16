package task

import (
	"time"

	"github.com/alkaid/behavior/bcore"
	"github.com/alkaid/behavior/timer"
	"github.com/alkaid/timingwheel"
)

type Action struct {
	bcore.Task
}

// OnStart
//  @override Node.OnStart
//  @receiver n
//  @param brain
func (a *Action) OnStart(brain bcore.IBrain) {
	a.Task.OnStart(brain)
	if !a.HasDelegatorOrScript() {
		a.Finish(brain, false)
		return
	}
	result := a.Update(brain, bcore.EventTypeOnStart, 0)
	if result != bcore.ResultInProgress {
		a.Finish(brain, result == bcore.ResultSucceeded)
		return
	}
	// 按root节点时钟频率定时调用
	interval := a.Root().Interval()
	a.stopTimer(brain)
	lastTime := time.Now()
	// 默认投递到黑板保存的线程ID
	a.Memory(brain).CronTask = timer.Cron(interval, 0, func() {
		currTime := time.Now()
		delta := currTime.Sub(lastTime)
		lastTime = currTime
		result := a.Update(brain, bcore.EventTypeOnUpdate, delta)
		if result != bcore.ResultInProgress {
			a.stopTimer(brain)
			a.Finish(brain, result == bcore.ResultSucceeded)
			return
		}
	}, timingwheel.WithGoID(brain.Blackboard().(bcore.IBlackboardInternal).ThreadID()))
}

// OnAbort
//  @override Node.OnAbort
//  @receiver n
//  @param brain
func (a *Action) OnAbort(brain bcore.IBrain) {
	a.Task.OnAbort(brain)
	// 中断时最后调用一次委托
	result := a.Update(brain, bcore.EventTypeOnAbort, 0)
	if result == bcore.ResultInProgress {
		a.Log().Error("action executor result cannot be ResultInProgress")
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
