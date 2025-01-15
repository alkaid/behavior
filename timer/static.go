package timer

import (
	"math/rand"
	"time"

	"github.com/alkaid/timingwheel"
)

var pool *TimeWheelPool // 全局时间轮单例,必须调用 InitPool 后方可使用
const half = 0.5

// InitPool 初始化 pool
//
//	@param size 池子容量
//	@param interval 时间轮帧间隔
//	@param numSlots 时间槽数量 时间轮第一层总时长=interval*numSlots
func InitPool(size int, interval time.Duration, numSlots int) {
	if pool != nil {
		return
	}
	pool = NewTimeWheelPool(size, interval, numSlots)
	pool.Start()
}

// TimeWheelInstance 从 pool 里获取一个时间轮
//
//	@return *TimeWheel
func TimeWheelInstance() *timingwheel.TimingWheel {
	return pool.Get()
}

// Cron wrap timingwheel.TimingWheel .Cron
//
//	@param interval 间隔
//	@param randomDeviation 随机离差范围,interval=interval+randomDeviation*[-0.5,0.5)
//	@param task
//	@param opts
func Cron(interval time.Duration, randomDeviation time.Duration, task func(), opts ...timingwheel.Option) *timingwheel.Timer {
	interval = interval - time.Duration(half*float32(randomDeviation)) + time.Duration(rand.Float32()*float32(randomDeviation))
	return TimeWheelInstance().Cron(interval, task, opts...)
}

// After wrap timingwheel.TimingWheel .AfterFunc
//
//	@param interval 间隔
//	@param randomDeviation 随机离差范围 interval = interval + randomDeviation*[-0.5,0.5)
//	@param task
//	@param opts
func After(interval time.Duration, randomDeviation time.Duration, task func(), opts ...timingwheel.Option) *timingwheel.Timer {
	interval = interval - time.Duration(half*float32(randomDeviation)) + time.Duration(rand.Float32()*float32(randomDeviation))
	return TimeWheelInstance().AfterFunc(interval, task, opts...)
}
