package timer

import (
	"math/rand"
	"time"

	"github.com/alkaid/timingwheel"
)

var poolInstance *TimeWheelPool // 全局时间轮单例,必须调用 InitPoolInstance 后方可使用
const minRandomDeviationRate = 0.5

// InitPoolInstance 初始化 poolInstance
//  @param size 池子容量
//  @param interval 时间轮帧间隔
//  @param numSlots 时间槽数量 时间轮第一层总时长=interval*numSlots
func InitPoolInstance(size int, interval time.Duration, numSlots int) {
	if poolInstance != nil {
		return
	}
	poolInstance = NewTimeWheelPool(size, interval, numSlots)
}

// TimeWheelInstance 从 pool 里获取一个时间轮
//  @return *TimeWheel
func TimeWheelInstance() *timingwheel.TimingWheel {
	return poolInstance.Get()
}

// Cron wrap timingwheel.TimingWheel .Cron
//  @param interval 间隔
//  @param randomDeviation 随机方差范围,会乘以[0,1)的一个随机数累加到interval上
//  @param task
//  @param opts
func Cron(interval time.Duration, randomDeviation time.Duration, task func(), opts ...timingwheel.Option) *timingwheel.Timer {
	interval = interval - time.Duration(minRandomDeviationRate*float32(randomDeviation)) + time.Duration(rand.Float32()*float32(randomDeviation))
	return TimeWheelInstance().Cron(interval, task, opts...)
}

// After wrap timingwheel.TimingWheel .AfterFunc
//  @param interval 间隔
//  @param randomDeviation 随机方差范围,会乘以[0,1)的一个随机数累加到interval上
//  @param task
//  @param opts
func After(interval time.Duration, randomDeviation time.Duration, task func(), opts ...timingwheel.Option) *timingwheel.Timer {
	interval = interval - time.Duration(minRandomDeviationRate*float32(randomDeviation)) + time.Duration(rand.Float32()*float32(randomDeviation))
	return TimeWheelInstance().AfterFunc(interval, task, opts...)
}
