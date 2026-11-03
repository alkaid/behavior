package behavior

import (
	"sync"
	"testing"
	"time"

	"github.com/alkaid/behavior/bcore"
	"github.com/alkaid/behavior/thread"
	"github.com/alkaid/behavior/timer"
)

// TestBrainTimersShareThread 验证定时回调和普通 Brain.Go 共用黑板所属的串行线程。
func TestBrainTimersShareThread(t *testing.T) {
	if err := thread.InitPool(nil); err != nil {
		t.Fatal(err)
	}
	timer.InitPool(1, time.Millisecond, 16)
	brain := NewBrain(bcore.NewBlackboard(901, nil), nil, nil)
	done := make(chan int, 2)
	var value int // 故意不加锁；共享线程失效时 race 检测应报告并发访问。
	blocked, unblock := make(chan struct{}), make(chan struct{})
	var release sync.Once
	t.Cleanup(func() { release.Do(func() { close(unblock) }) })
	brain.Go(func() { close(blocked); <-unblock; value = 10 })
	select {
	case <-blocked:
	case <-time.After(3 * time.Second):
		t.Fatal("brain thread did not start")
	}
	for i := 0; i < 10; i++ {
		brain.Go(func() { value++ })
	}
	after := brain.After(3*time.Millisecond, 0, func() { value++; done <- value })
	t.Cleanup(func() { after.Stop() })
	select {
	case <-done:
		t.Fatal("timer bypassed the blocked brain thread")
	case <-time.After(20 * time.Millisecond):
	}
	// 一个 AI 阻塞时，另一黑板 ID 的定时器仍应独立执行。
	otherDone := make(chan struct{})
	other := NewBrain(bcore.NewBlackboard(902, nil), nil, nil)
	otherTimer := other.After(time.Millisecond, 0, func() { close(otherDone) })
	t.Cleanup(func() { otherTimer.Stop() })
	select {
	case <-otherDone:
	case <-time.After(3 * time.Second):
		t.Fatal("another brain was blocked by the first brain")
	}
	release.Do(func() { close(unblock) })
	select {
	case got := <-done:
		if got != 21 {
			t.Fatalf("timer observed %d, want 21 after queued brain tasks", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("brain timer did not finish")
	}
	ready := make(chan struct{})
	var cron *timer.Timer
	cron = brain.Cron(time.Millisecond, 0, func() {
		<-ready
		value++
		cron.Stop()
		done <- value
	})
	t.Cleanup(func() { cron.Stop() })
	close(ready)
	select {
	case got := <-done:
		if got != 22 {
			t.Fatalf("cron observed %d, want 22", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("brain cron did not finish")
	}
}
