package timer

import (
	"sync"
	"time"

	"github.com/alkaid/ants/v2"
	"github.com/alkaid/behavior/internal/dispatch"
	"github.com/alkaid/behavior/logger"
	"github.com/zeromicro/go-zero/core/collection"
	"go.uber.org/zap"
)

// Options 指定定时回调的执行线程；未指定 Pool 时使用 ants 默认池。
type Options struct {
	Pool *ants.PoolWithID
	GoID int
}

// Option 配置定时回调的线程归属。
type Option func(*Options)

// WithPool 让定时回调和行为树共享同一个串行任务池。
func WithPool(pool *ants.PoolWithID) Option {
	return func(options *Options) { options.Pool = pool }
}

// WithGoID 指定回调在任务池中的线程 ID，与黑板的 ThreadID 保持一致。
func WithGoID(id int) Option {
	return func(options *Options) { options.GoID = id }
}

// Timer 在 go-zero 单次调度上维护循环和取消状态。
// 回调进入目标线程后再次检查状态，避免已取消的排队任务继续访问行为树。
type Timer struct {
	wheel    *collection.TimingWheel
	options  Options
	interval time.Duration
	repeat   bool
	mu       sync.Mutex
	next     time.Time
	active   bool
	task     func()
}

func newTimer(wheel *collection.TimingWheel, interval time.Duration, repeat bool, task func(), opts ...Option) *Timer {
	// go-zero 不接收非正延时；交给它的最小正值会被提升到一个 tick，避免空转。
	if interval <= 0 {
		interval = time.Nanosecond
	}
	t := &Timer{
		wheel: wheel, interval: interval, repeat: repeat,
		next: time.Now().Add(interval), active: true, task: task,
		options: Options{GoID: -1},
	}
	for _, opt := range opts {
		opt(&t.options)
	}
	t.mu.Lock()
	t.schedule()
	t.mu.Unlock()
	return t
}

// schedule 必须持有 mu，与 Stop 串行处理，取消后不能再插入下一次循环。
func (t *Timer) schedule() {
	delay := time.Until(t.next)
	if delay <= 0 {
		delay = time.Nanosecond
	}
	if err := t.wheel.SetTimer(t, t, delay); err != nil {
		t.active = false
		t.task = nil
		logger.Log.Error("schedule behavior timer failed", zap.Error(err))
	}
}

func (t *Timer) dispatch() {
	if t.options.Pool != nil {
		// 与 Brain.Go 共用有序派发，且一个满 ID 不会阻塞 go-zero 同批的其他任务。
		dispatch.Submit(t.options.Pool, t.options.GoID, t.run, t.submitFailed)
	} else {
		// 默认池也可能被调用方限容，准入等待不能留在时间轮的 execute 中。
		go func() {
			if err := ants.Submit(t.run); err != nil {
				t.submitFailed(err)
			}
		}()
	}
}

func (t *Timer) submitFailed(err error) {
	t.Stop()
	logger.Log.Error("submit behavior timer failed", zap.Error(err))
}

func (t *Timer) run() {
	t.mu.Lock()
	if !t.active {
		t.mu.Unlock()
		return
	}
	task := t.task
	if t.repeat {
		// 沿用原定时器的固定节奏，在业务回调前安排下一次，而非按完成时间累计漂移。
		t.next = t.next.Add(t.interval)
		t.schedule()
	} else {
		t.active = false
		t.task = nil
	}
	t.mu.Unlock()
	task()
}

// Stop 取消尚未开始的回调及后续循环，返回是否取消了活跃任务。
// 不等待正在执行的回调，回调内部也可以安全调用 Stop。
func (t *Timer) Stop() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return false
	}
	t.active = false
	// go-zero 延迟清理已删除的槽位；立即释放闭包，避免继续持有已销毁的 AI。
	t.task = nil
	if err := t.wheel.RemoveTimer(t); err != nil {
		logger.Log.Error("remove behavior timer failed", zap.Error(err))
	}
	return true
}
