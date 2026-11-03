package timer

import (
	"math/rand/v2"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/collection"
)

// TimeWheelPool 时间轮池子
type TimeWheelPool struct {
	pool []*collection.TimingWheel
	size int64
	incr int64 // not need for high accuracy
}

// NewTimeWheelPool
//
//	@param size 池子容量
//	@param interval 时间轮帧间隔
//	@param numSlots 时间槽数量 时间轮第一层总时长=interval*numSlots
//	@return *TimeWheelPool
func NewTimeWheelPool(size int, interval time.Duration, numSlots int) *TimeWheelPool {
	if size <= 0 {
		panic("timer pool size must be positive")
	}
	twp := &TimeWheelPool{
		pool: make([]*collection.TimingWheel, size),
		size: int64(size),
	}
	for index := 0; index < size; index++ {
		// go-zero 在构造时启动时间轮；到期任务仍交回指定的 ants 线程。
		tw, err := collection.NewTimingWheel(interval, numSlots, func(_, value any) {
			value.(*Timer).dispatch()
		})
		if err != nil {
			panic(err)
		}
		twp.pool[index] = tw
	}
	return twp
}

// Get 顺序获取下一个时间轮
//
//	@receiver tp
//	@return *TimeWheel
func (tp *TimeWheelPool) Get() *collection.TimingWheel {
	incr := atomic.AddInt64(&tp.incr, 1)
	idx := incr % tp.size
	return tp.pool[idx]
}

// GetRandom 随机获取一个时间轮
//
//	@receiver tp
//	@return *TimeWheel
func (tp *TimeWheelPool) GetRandom() *collection.TimingWheel {
	return tp.pool[rand.IntN(int(tp.size))]
}

// Stop 停止
//
//	@receiver tp
func (tp *TimeWheelPool) Stop() {
	for _, tw := range tp.pool {
		tw.Stop()
	}
}
