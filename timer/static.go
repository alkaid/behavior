package timer

import (
	"time"

	"github.com/alkaid/timingwheel"
)

var poolInstance *TimeWheelPool // 全局时间轮单例,必须调用 InitPoolInstance 后方可使用

// InitPoolInstance 初始化 poolInstance
//  @param size
//  @param tick
//  @param bucketsNum
//
func InitPoolInstance(size int, tick time.Duration, bucketsNum int) {
	if poolInstance != nil {
		return
	}
	poolInstance = NewTimeWheelPool(size, tick, bucketsNum)
}

// TimeWheelInstance 从 pool 里获取一个时间轮
//  @return *TimeWheel
//
func TimeWheelInstance() *timingwheel.TimingWheel {
	return poolInstance.Get()
}

func After() {

}
