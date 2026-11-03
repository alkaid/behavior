package timer

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alkaid/ants/v2"
	"github.com/zeromicro/go-zero/core/collection"
)

func initTestPool(t *testing.T) {
	t.Helper()
	InitPool(2, time.Millisecond, 8)
	t.Cleanup(func() {
		pool.Stop()
		pool = nil
	})
}

func await[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(3 * time.Second):
		t.Fatal("timer task did not finish")
		var zero T
		return zero
	}
}

// TestAfter 覆盖非正延时、随机离差以及跨多圈延时，且单次任务只能执行一次。
func TestAfter(t *testing.T) {
	initTestPool(t)
	for _, delay := range []time.Duration{-time.Second, 0, time.Millisecond, 40 * time.Millisecond} {
		t.Run(delay.String(), func(t *testing.T) {
			done := make(chan struct{}, 2)
			timer := After(delay, time.Millisecond, func() { done <- struct{}{} })
			t.Cleanup(func() { timer.Stop() })
			await(t, done)
			if timer.Stop() {
				t.Fatal("completed one-shot timer must be inactive")
			}
			select {
			case <-done:
				t.Fatal("one-shot timer ran twice")
			case <-time.After(15 * time.Millisecond):
			}
		})
	}
}

// TestCron 允许在回调内停止循环，并保证停止后不再重排。
func TestCron(t *testing.T) {
	initTestPool(t)
	p, err := ants.NewPoolWithID(1, ants.WithTaskBuffer(8))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Release)
	for _, delay := range []time.Duration{0, 4 * time.Millisecond} {
		t.Run(delay.String(), func(t *testing.T) {
			ready := make(chan struct{})
			done := make(chan bool, 1)
			var calls atomic.Int32
			var timer *Timer
			timer = Cron(delay, 0, func() {
				<-ready
				if calls.Add(1) == 3 {
					done <- timer.Stop()
				}
			}, WithPool(p), WithGoID(1))
			t.Cleanup(func() { timer.Stop() })
			close(ready)
			if !await(t, done) {
				t.Fatal("callback must be able to stop its next iteration")
			}
			time.Sleep(20 * time.Millisecond)
			if got := calls.Load(); got != 3 {
				t.Fatalf("cron ran %d times after cancellation, want 3", got)
			}
		})
	}
}

// TestStopBeforeExpiry 验证取消结果和重复取消的语义。
func TestStopBeforeExpiry(t *testing.T) {
	initTestPool(t)
	done := make(chan struct{}, 1)
	timer := After(15*time.Millisecond, 0, func() { done <- struct{}{} })
	if !timer.Stop() || timer.Stop() {
		t.Fatal("only the first cancellation should report success")
	}
	select {
	case <-done:
		t.Fatal("canceled timer ran")
	case <-time.After(30 * time.Millisecond):
	}
}

// TestStopQueuedCallback 保证已投递但尚未进入 AI 线程的回调仍可取消。
func TestStopQueuedCallback(t *testing.T) {
	p, err := ants.NewPoolWithID(1, ants.WithTaskBuffer(8))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Release)
	blocked, unblock := make(chan struct{}), make(chan struct{})
	var release sync.Once
	t.Cleanup(func() { release.Do(func() { close(unblock) }) })
	if err = p.Submit(7, func() { close(blocked); <-unblock }); err != nil {
		t.Fatal(err)
	}
	await(t, blocked)
	enqueued := make(chan struct{})
	wheel, err := collection.NewTimingWheel(time.Millisecond, 8, func(_, value any) {
		value.(*Timer).dispatch()
		close(enqueued)
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(wheel.Stop)
	var calls atomic.Int32
	timer := newTimer(wheel, time.Millisecond, false, func() { calls.Add(1) }, WithPool(p), WithGoID(7))
	await(t, enqueued)
	if calls.Load() != 0 || !timer.Stop() {
		t.Fatal("timer must wait on the selected thread and remain cancellable")
	}
	flushed := make(chan struct{})
	if err = p.Submit(7, func() { close(flushed) }); err != nil {
		t.Fatal(err)
	}
	release.Do(func() { close(unblock) })
	await(t, flushed)
	if calls.Load() != 0 {
		t.Fatal("canceled queued callback ran")
	}
}

// TestConcurrentStopAndExpiry 反复交错到期、循环重排与取消，供 race 检测覆盖状态同步。
func TestConcurrentStopAndExpiry(t *testing.T) {
	initTestPool(t)
	var workers sync.WaitGroup
	for range 100 {
		timer := Cron(time.Millisecond, 0, func() {})
		workers.Add(1)
		go func() {
			defer workers.Done()
			time.Sleep(time.Millisecond)
			timer.Stop()
		}()
	}
	workers.Wait()
}

// TestClosedWheel 不应在调度失败后留下仍标记活跃的任务。
func TestClosedWheel(t *testing.T) {
	p := NewTimeWheelPool(1, time.Millisecond, 8)
	p.Stop()
	timer := newTimer(p.Get(), time.Second, false, func() { t.Error("closed wheel ran task") })
	if timer.Stop() {
		t.Fatal("failed scheduling must leave the timer inactive")
	}
}

// TestInitPool 检查池轮转与重复初始化，避免替换运行中的全局时间轮。
func TestInitPool(t *testing.T) {
	initTestPool(t)
	first, second := TimeWheelInstance(), TimeWheelInstance()
	if first == second || TimeWheelInstance() != first {
		t.Fatal("wheel selection must rotate through the configured pool")
	}
	previous := pool
	InitPool(3, time.Second, 16)
	if pool != previous {
		t.Fatal("reinitialization replaced the active pool")
	}
	for range 20 {
		if got := pool.GetRandom(); got != first && got != second {
			t.Fatal("random selection returned an unknown wheel")
		}
	}
}
