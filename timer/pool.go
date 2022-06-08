package timer

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/alkaid/timingwheel"
)

// TimeWheelPool 时间轮池子
type TimeWheelPool struct {
	pool []*timingwheel.TimingWheel
	size int64
	incr int64 // not need for high accuracy
}

// NewTimeWheelPool
//  @param size 池子容量
//  @param interval 时间轮帧间隔
//  @param numSlots 时间槽数量 时间轮第一层总时长=interval*numSlots
//  @return *TimeWheelPool
//
func NewTimeWheelPool(size int, interval time.Duration, numSlots int) *TimeWheelPool {
	twp := &TimeWheelPool{
		pool: make([]*timingwheel.TimingWheel, size),
		size: int64(size),
	}
	for index := 0; index < size; index++ {
		tw := timingwheel.NewTimingWheel(interval, int64(numSlots))
		twp.pool[index] = tw
	}
	return twp
}

// Get 顺序获取下一个时间轮
//  @receiver tp
//  @return *TimeWheel
//
func (tp *TimeWheelPool) Get() *timingwheel.TimingWheel {
	incr := atomic.AddInt64(&tp.incr, 1)
	idx := incr % tp.size
	return tp.pool[idx]
}

// GetRandom 随机获取一个时间轮
//  @receiver tp
//  @return *TimeWheel
//
func (tp *TimeWheelPool) GetRandom() *timingwheel.TimingWheel {
	return tp.pool[rand.Intn(int(tp.size))]
}

// Start 启动
//  @receiver tp
//
func (tp *TimeWheelPool) Start() {
	for _, tw := range tp.pool {
		tw.Start()
	}
}

// Stop 停止
//  @receiver tp
//
func (tp *TimeWheelPool) Stop() {
	for _, tw := range tp.pool {
		tw.Stop()
	}
}
