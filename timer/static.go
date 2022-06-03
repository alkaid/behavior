package timer

import (
	"time"

	"github.com/alkaid/timingwheel"
)

var poolInstance *TimeWheelPool // 全局时间轮单例,必须调用 InitPoolInstance 后方可使用

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
