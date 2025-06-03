package thread

import (
	"math"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/alkaid/behavior/logger"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

const MainThreadID = math.MaxInt // 主线程ID
const DefaultTaskBuffer = 100

var pool *ants.PoolWithID // 线程池,主要用于分离session线程

func PoolInstance() *ants.PoolWithID {
	return pool
}

// InitPool 用默认参数初始化全局线程池
func InitPool(p *ants.PoolWithID) error {
	if pool != nil {
		logger.Log.Warn("init goroutine pool duplicate")
		return nil
	}
	if p != nil {
		pool = p
		return nil
	}
	p, err := ants.NewPoolWithID(
		ants.DefaultAntsPoolSize,
		ants.WithTaskBuffer(DefaultTaskBuffer),
		ants.WithExpiryDuration(time.Hour),
		ants.WithDisablePurgeRunning(false))
	if err != nil {
		return errors.WithStack(err)
	}
	pool = p
	return nil
}

func ReleaseTableGoPool() {
	if pool != nil {
		pool.Release()
	}
}

// GoByID 根据指定的goroutineID派发线程
//
//	@receiver h
//	@param goID 若>0,派发到指定线程,否则随机派发
//	@param task
func GoByID[T int | int32 | int64](goID T, task func()) {
	if goID > 0 {
		err := pool.Submit(int(goID), task)
		if err != nil {
			logger.Log.Error("submit goroutine with id error", zap.Error(err), zap.Int("goID", int(goID)))
			return
		}
	} else {
		Go(task)
	}
}

func WaitByID[T int | int32 | int64](goID T, task func()) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	GoByID(goID, func() {
		task()
		wg.Done()
	})
	wg.Wait()
}

// GoMain 派发到主线程
//
//	@param task
func GoMain(task func()) {
	GoByID(MainThreadID, task)
}

// WaitMain 派发到主线程并等待执行完成
//
//	@param task
func WaitMain(task func()) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	GoMain(func() {
		task()
		wg.Done()
	})
	wg.Wait()
}

// Go 从默认线程池获取一个goroutine并派发任务
//
//	@param task
func Go(task func()) {
	err := ants.Submit(task)
	if err != nil {
		logger.Log.Error("submit goroutine error", zap.Error(err))
		return
	}
}
