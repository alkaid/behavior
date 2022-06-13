package task

import (
	"time"

	"github.com/alkaid/behavior/util"

	"github.com/alkaid/behavior/timer"
	"github.com/alkaid/timingwheel"

	"github.com/alkaid/behavior/bcore"
)

type IWaitProperties interface {
	GetWaitTime() time.Duration
	GetRandomDeviation() time.Duration
	GetForever() bool
}

// WaitProperties 等待属性
type WaitProperties struct {
	WaitTime        util.Duration `json:"waitTime"`        // 等待时间
	RandomDeviation util.Duration `json:"randomDeviation"` // 随机偏差:允许向 等待时间（Wait Time） 属性添加随机时间（正或负）
	Forever         bool          `json:"forever"`         // 永久等待直到被外界打断
}

func (w *WaitProperties) GetWaitTime() time.Duration {
	return w.WaitTime.Duration
}

func (w *WaitProperties) GetRandomDeviation() time.Duration {
	return w.RandomDeviation.Duration
}

func (w *WaitProperties) GetForever() bool {
	return w.Forever
}

// Wait 等待
//  任务节点可以在行为树中使用，使树在此节点上等待，直至指定的 等待时间（Wait Time） 结束。
type Wait struct {
	bcore.Task
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (w *Wait) PropertiesClassProvider() any {
	return &WaitProperties{}
}
func (w *Wait) WaitProperties() IWaitProperties {
	return w.Properties().(IWaitProperties)
}

// OnStart
//  @override Node.OnStart
//  @receiver n
//  @param brain
func (w *Wait) OnStart(brain bcore.IBrain) {
	w.Task.OnStart(brain)
	w.Memory(brain).Elapsed = 0
	if w.WaitProperties().GetForever() {
		return
	}
	w.Memory(brain).CronTask = timer.After(w.WaitProperties().GetWaitTime(), w.WaitProperties().GetRandomDeviation(), func() {
		w.Finish(brain, true)
	}, timingwheel.WithGoID(brain.Blackboard().(bcore.IBlackboardInternal).ThreadID()))
}

// OnAbort
//  @override Node.OnAbort
//  @receiver n
//  @param brain
func (w *Wait) OnAbort(brain bcore.IBrain) {
	w.Task.OnAbort(brain)
	w.stopTimer(brain)
	w.Finish(brain, false)
}

func (w *Wait) stopTimer(brain bcore.IBrain) {
	if w.Memory(brain).CronTask != nil {
		w.Memory(brain).CronTask.Stop()
		w.Memory(brain).CronTask = nil
	}
}
