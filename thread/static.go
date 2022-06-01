package thread

import (
	"hash/crc32"
	"strconv"
	"time"

	"github.com/alkaid/behavior/logger"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

var pool *ants.Pool // 线程池,主要用于分离session线程
const taskBuffer = ants.DefaultStatefulTaskBuffer * 10

// InitDefaultPool 用默认参数初始化全局线程池
func InitDefaultPool() {
	if pool != nil {
		logger.Log.Warn("init goroutine pool duplicate")
		return
	}
	p, err := ants.NewPool(ants.DefaultAntsPoolSize, ants.WithTaskBuffer(taskBuffer), ants.WithExpiryDuration(time.Hour))
	if err != nil {
		logger.Log.Fatal("create ants pool error", zap.Error(err))
	}
	pool = p
}

// InitWithPool 用指定线程池初始化全局线程池
//  @param p
//
func InitWithPool(p *ants.Pool) {
	if pool != nil {
		logger.Log.Warn("init goroutine pool duplicate")
		return
	}
	pool = p
}

// GoByID 根据指定的goroutineID派发线程
//  @receiver h
//  @param goID 若>0,派发到指定线程,否则随机派发
//  @param task
func GoByID(goID int, task func()) {
	if goID > 0 {
		err := pool.SubmitWithID(goID, task)
		if err != nil {
			logger.Log.Error("submit task with id error", zap.Error(err), zap.Int("goID", goID))
			return
		}
	} else {
		Go(task)
	}
}

// GoByUID 根据uid派发任务线程
//  @param uid
//  @param task
func GoByUID(uid string, task func()) {
	if uid == "" {
		logger.Log.Error("uid can not be empty")
	}
	goID, err := strconv.Atoi(uid)
	if err != nil {
		logger.Log.Warn("can't atoi uid", zap.String("uid", uid), zap.Error(err))
		goID = int(crc32.ChecksumIEEE([]byte(uid)))
	}
	GoByID(goID, task)
}

// Go 从默认线程池获取一个goroutine并派发任务
//  @param task
func Go(task func()) {
	err := ants.Submit(task)
	if err != nil {
		logger.Log.Error("submit task error", zap.Error(err))
		return
	}
}
