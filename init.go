package behavior

import (
	"math/rand"
	"time"

	"github.com/alkaid/behavior/decorator"

	"github.com/alkaid/behavior/bcore"

	"github.com/alkaid/behavior/composite"
	"github.com/alkaid/behavior/task"

	"github.com/alkaid/behavior/logger"
	"go.uber.org/zap/zapcore"

	"github.com/alkaid/behavior/thread"
	"github.com/alkaid/behavior/timer"

	"github.com/panjf2000/ants/v2"
)

// InitSystem 系统初始化
//  要使用该库必须先初始化
//  @param option
func InitSystem(opts ...Option) {
	rand.Seed(time.Now().UnixNano())
	option := &InitialOption{
		ThreadPool:     nil,
		TimerPoolSize:  1,
		TimerInterval:  10 * time.Millisecond,
		TimerNumSlots:  100,
		LogLevel:       zapcore.ErrorLevel,
		LogDevelopment: false,
	}
	for _, opt := range opts {
		opt(option)
	}
	thread.InitPool(option.ThreadPool)
	timer.InitPoolInstance(option.TimerPoolSize, option.TimerInterval, option.TimerNumSlots)
	logger.Manager.SetDevelopment(option.LogDevelopment)
	logger.Manager.SetLevel(option.LogLevel)
	// built in class register
	GlobalClassLoader().Register(&bcore.Root{})

	GlobalClassLoader().Register(&composite.Sequence{})
	GlobalClassLoader().Register(&composite.Selector{})
	GlobalClassLoader().Register(&composite.RandomSequence{})
	GlobalClassLoader().Register(&composite.RandomSelector{})
	GlobalClassLoader().Register(&composite.Parallel{})

	GlobalClassLoader().Register(&decorator.BBCondition{})
	GlobalClassLoader().Register(&decorator.BBCooldown{})
	GlobalClassLoader().Register(&decorator.BBEntries{})
	GlobalClassLoader().Register(&decorator.Condition{})
	GlobalClassLoader().Register(&decorator.Cooldown{})
	GlobalClassLoader().Register(&decorator.Failure{})
	GlobalClassLoader().Register(&decorator.Inverter{})
	GlobalClassLoader().Register(&decorator.Random{})
	GlobalClassLoader().Register(&decorator.Repeater{})
	GlobalClassLoader().Register(&decorator.Service{})
	GlobalClassLoader().Register(&decorator.Succeeded{})
	GlobalClassLoader().Register(&decorator.TimeMax{})
	GlobalClassLoader().Register(&decorator.TimeMin{})
	GlobalClassLoader().Register(&decorator.WaitCondition{})

	GlobalClassLoader().Register(&task.Action{})
	GlobalClassLoader().Register(&task.Wait{})
	GlobalClassLoader().Register(&task.WaitBB{})
	GlobalClassLoader().Register(&task.Subtree{})
	GlobalClassLoader().Register(&task.DynamicSubtree{})
	// 注册自定义节点
	for _, class := range option.CustomNodeClass {
		GlobalClassLoader().Register(class)
	}
}

type InitialOption struct {
	ThreadPool      *ants.Pool    // 线程池 为空则使用默认
	TimerPoolSize   int           // 时间轮池子容量 为0则使用默认
	TimerInterval   time.Duration // 时间轮帧间隔 为0则使用默认
	TimerNumSlots   int           // 时间槽数量 时间轮第一层总时长=interval*numSlots 为0则使用默认
	LogLevel        zapcore.Level // 日志级别
	LogDevelopment  bool          // 日志模式是否开发模式
	CustomNodeClass []bcore.INode // 用于注册自定义节点类
}

type Option func(option *InitialOption)

func WithThreadPool(pool *ants.Pool) Option {
	return func(option *InitialOption) {
		option.ThreadPool = pool
	}
}
func WithTimerPoolSize(size int) Option {
	return func(o *InitialOption) {
		o.TimerPoolSize = size
	}
}

func WithTimerInterval(interval time.Duration) Option {
	return func(o *InitialOption) {
		o.TimerInterval = interval
	}
}
func WithTimerNumSlots(slots int) Option {
	return func(o *InitialOption) {
		o.TimerNumSlots = slots
	}
}
func WithLogLevel(level zapcore.Level) Option {
	return func(o *InitialOption) {
		o.LogLevel = level
	}
}
func WithLogDevelopment(development bool) Option {
	return func(o *InitialOption) {
		o.LogDevelopment = development
	}
}

func WithCustomNodes(nodes []bcore.INode) Option {
	return func(o *InitialOption) {
		o.CustomNodeClass = nodes
	}
}
