package task

import (
	"time"

	"github.com/alkaid/behavior/util"

	"go.uber.org/zap"

	"github.com/alkaid/behavior/bcore"
)

type IWaitBBProperties interface {
	GetKey() string
	GetRandomDeviation() time.Duration
}

// WaitBBProperties 等待黑板时间属性
type WaitBBProperties struct {
	Key             string        `json:"key"`             // 引用的黑板键，确定等待时间
	RandomDeviation util.Duration `json:"randomDeviation"` // 随机偏差:允许向 等待时间（WaitTime）属性添加随机时间 WaitTime=WaitTime+RandomDeviation*[-0.5,0.5)
}

func (w *WaitBBProperties) GetKey() string {
	return w.Key
}

func (w *WaitBBProperties) GetRandomDeviation() time.Duration {
	return w.RandomDeviation.Duration
}

// WaitBB 等待黑板时间
//
//	与 等待 WaitBB 任务节点的原理类似，但该节点会拉取等待时间黑板值。
type WaitBB struct {
	bcore.Task
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (w *WaitBB) PropertiesClassProvider() any {
	return &WaitBBProperties{}
}
func (w *WaitBB) WaitBBProperties() IWaitBBProperties {
	return w.Properties().(IWaitBBProperties)
}

// OnStart
//
//	@override Node.OnStart
//	@receiver n
//	@param brain
func (w *WaitBB) OnStart(brain bcore.IBrain) {
	w.Task.OnStart(brain)
	w.Memory(brain).Elapsed = 0
	delay, ok := brain.Blackboard().GetDuration(w.WaitBBProperties().GetKey())
	if !ok {
		w.Log(brain).Error("not found wait time in blackboard", zap.String("key", w.WaitBBProperties().GetKey()))
		// 取值失败则默认为不等待
		w.Finish(brain, true)
	}
	w.Memory(brain).CronTask = brain.After(delay, w.WaitBBProperties().GetRandomDeviation(), func() {
		if !w.IsActive(brain) {
			return
		}
		w.Finish(brain, true)
	})
}

// OnAbort
//
//	@override Node.OnAbort
//	@receiver n
//	@param brain
func (w *WaitBB) OnAbort(brain bcore.IBrain) {
	w.Task.OnAbort(brain)
	w.stopTimer(brain)
	w.Finish(brain, false)
}

func (w *WaitBB) stopTimer(brain bcore.IBrain) {
	if w.Memory(brain).CronTask != nil {
		w.Memory(brain).CronTask.Stop()
		w.Memory(brain).CronTask = nil
	}
}
