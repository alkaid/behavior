package decorator

import (
	"time"

	"github.com/alkaid/behavior/util"

	"github.com/alkaid/timingwheel"

	"github.com/alkaid/behavior/timer"

	"github.com/alkaid/behavior/bcore"
)

type IServiceProperties interface {
	GetInterval() time.Duration
	GetRandomDeviation() time.Duration
}

// ServiceProperties 服务节点属性
type ServiceProperties struct {
	Interval        util.Duration `json:"interval"`        // 执行间隔，配0则为行为树默认时间轮间隔
	RandomDeviation util.Duration `json:"randomDeviation"` // 随机偏差:将一个随机范围数值添加至服务节点的 时间间隔（Interval） 值。
}

func (s *ServiceProperties) GetInterval() time.Duration {
	return s.Interval.Duration
}

func (s *ServiceProperties) GetRandomDeviation() time.Duration {
	return s.RandomDeviation.Duration
}

// Service 服务.
//  节点通常连接至合成（Composite）节点或任务（Task）节点，只要其分支被执行，它们就会以定义的频率执行。
//  这些节点常用于检查和更新黑板。它们取代了其他行为树系统中的传统平行（Parallel）节点。
type Service struct {
	bcore.Decorator
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (s *Service) PropertiesClassProvider() any {
	return &ServiceProperties{}
}

// OnStart
//  @override bcore.Decorator .OnStart
//  @receiver s
//  @param brain
func (s *Service) OnStart(brain bcore.IBrain) {
	s.Decorator.OnStart(brain)
	interval := s.Properties().(IServiceProperties).GetInterval()
	randomDeviation := s.Properties().(IServiceProperties).GetRandomDeviation()
	if interval <= 0 {
		interval = s.Root().Interval()
	}
	lastTime := time.Now()
	s.stopTimer(brain)
	// 默认投递到黑板保存的线程ID
	s.Memory(brain).CronTask = timer.Cron(interval, randomDeviation, func() {
		currTime := time.Now()
		delta := currTime.Sub(lastTime)
		s.Execute(brain, bcore.EventTypeOnStart, delta)
		lastTime = currTime
	}, timingwheel.WithGoID(brain.Blackboard().(bcore.IBlackboardInternal).ThreadID()))
	s.Execute(brain, bcore.EventTypeOnStart, 0)
	s.Decorated(brain).Start(brain)
}

// OnChildFinished
//  @override bcore.Decorator .OnChildFinished
//  @receiver s
//  @param brain
//  @param child
//  @param succeeded
func (s *Service) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	s.Decorator.OnChildFinished(brain, child, succeeded)
	s.stopTimer(brain)
	s.Finish(brain, succeeded)
}

func (s *Service) stopTimer(brain bcore.IBrain) {
	if s.Memory(brain).CronTask != nil {
		s.Memory(brain).CronTask.Stop()
		s.Memory(brain).CronTask = nil
	}
}
