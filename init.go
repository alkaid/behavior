package behavior

import (
	"math/rand"
	"time"

	"github.com/alkaid/behavior/script"
	"go.uber.org/zap"

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
//
//	要使用该库必须先初始化
//	@param option
//	@panic 若发生异常将主动panic
func InitSystem(opts ...Option) {
	rand.Seed(time.Now().UnixNano())
	option := &InitialOption{
		ThreadPool:        nil,
		TimerPoolSize:     1,
		TimerInterval:     10 * time.Millisecond,
		TimerNumSlots:     100,
		LogLevel:          zapcore.ErrorLevel,
		LogDevelopment:    false,
		ScriptPoolMinSize: 10,
		ScriptPoolMaxSize: 1000,
	}
	for _, opt := range opts {
		opt(option)
	}
	logger.SetDevelopment(option.LogDevelopment)
	logger.SetLevel(option.LogLevel)
	err := thread.InitPool(option.ThreadPool)
	if err != nil {
		logger.Log.Fatal("init behavior system error", zap.Error(err))
		return
	}
	timer.InitPool(option.TimerPoolSize, option.TimerInterval, option.TimerNumSlots)
	err = script.InitPool(option.ScriptPoolMinSize, option.ScriptPoolMaxSize, option.ScriptPoolApiLib)
	if err != nil {
		logger.Log.Fatal("init behavior system error", zap.Error(err))
		return
	}
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
	ThreadPool        *ants.PoolWithID // 线程池 为空则使用默认
	TimerPoolSize     int              // 时间轮池子容量 为0则使用默认
	TimerInterval     time.Duration    // 时间轮帧间隔 为0则使用默认
	TimerNumSlots     int              // 时间槽数量 时间轮第一层总时长=interval*numSlots 为0则使用默认
	LogLevel          zapcore.Level    // 日志级别
	LogDevelopment    bool             // 日志模式是否开发模式
	CustomNodeClass   []bcore.INode    // 用于注册自定义节点类
	ScriptPoolMinSize int              // 脚本引擎池子最小容量
	ScriptPoolMaxSize int              // 脚本引擎池子最大容量
	ScriptPoolApiLib  map[string]any   // 需注入到脚本引擎池的api库,最好仅注入一些公共的无状态函数或参数,避免状态副作用
}

type Option func(option *InitialOption)

func WithThreadPool(pool *ants.PoolWithID) Option {
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

func WithScriptPoolMinSize(size int) Option {
	return func(o *InitialOption) {
		o.ScriptPoolMinSize = size
	}
}
func WithScriptPoolMaxSize(size int) Option {
	return func(o *InitialOption) {
		o.ScriptPoolMaxSize = size
	}
}
func WithScriptPoolApiLib(api map[string]any) Option {
	return func(o *InitialOption) {
		o.ScriptPoolApiLib = api
	}
}
