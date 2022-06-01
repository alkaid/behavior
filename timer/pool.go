package timer

import (
	"crypto/rand"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/alkaid/timingwheel"

	"github.com/alkaid/behavior/logger"
	"go.uber.org/zap"
)

// TimeWheelPool 时间轮池子
type TimeWheelPool struct {
	pool []*timingwheel.TimingWheel
	size int64
	incr int64 // not need for high accuracy
}

// NewTimeWheelPool
//  @param size
//  @param tick
//  @param bucketsNum
//  @return *TimeWheelPool
//
func NewTimeWheelPool(size int, tick time.Duration, bucketsNum int) *TimeWheelPool {
	twp := &TimeWheelPool{
		pool: make([]*timingwheel.TimingWheel, size),
		size: int64(size),
	}
	for index := 0; index < bucketsNum; index++ {
		tw := timingwheel.NewTimingWheel(tick, int64(bucketsNum))
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
	idx, err := rand.Int(rand.Reader, big.NewInt(tp.size))
	if err != nil {
		logger.Log.Error("", zap.Error(err))
		idx = big.NewInt(0)
	}
	return tp.pool[idx.Int64()]
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
