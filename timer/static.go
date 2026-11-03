package timer

import (
	"math/rand/v2"
	"time"

	"github.com/zeromicro/go-zero/core/collection"
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
}

// TimeWheelInstance 从 pool 里获取一个时间轮
//
//	@return *TimeWheel
func TimeWheelInstance() *collection.TimingWheel {
	return pool.Get()
}

// Cron 创建循环任务；随机离差只在创建时采样一次，后续沿用相同间隔。
//
//	@param interval 间隔
//	@param randomDeviation 随机离差范围,interval=interval+randomDeviation*[-0.5,0.5)
//	@param task
//	@param opts
func Cron(interval time.Duration, randomDeviation time.Duration, task func(), opts ...Option) *Timer {
	interval = interval - time.Duration(half*float32(randomDeviation)) + time.Duration(rand.Float32()*float32(randomDeviation))
	return newTimer(TimeWheelInstance(), interval, true, task, opts...)
}

// After 创建单次任务；非正延时在下一时间轮 tick 异步触发。
//
//	@param interval 间隔
//	@param randomDeviation 随机离差范围 interval = interval + randomDeviation*[-0.5,0.5)
//	@param task
//	@param opts
func After(interval time.Duration, randomDeviation time.Duration, task func(), opts ...Option) *Timer {
	interval = interval - time.Duration(half*float32(randomDeviation)) + time.Duration(rand.Float32()*float32(randomDeviation))
	return newTimer(TimeWheelInstance(), interval, false, task, opts...)
}
