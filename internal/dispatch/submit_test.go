package dispatch

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alkaid/ants/v2"
)

func newTestPool(t *testing.T) *ants.PoolWithID {
	t.Helper()
	p, err := ants.NewPoolWithID(1, ants.WithTaskBuffer(1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Release)
	return p
}

func waitIdle(t *testing.T) {
	t.Helper()
	deadline := time.After(time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		mu.Lock()
		pending := len(queues)
		mu.Unlock()
		if pending == 0 {
			return
		}
		select {
		case <-ticker.C:
		case <-deadline:
			t.Fatal("idle dispatch queues were retained")
		}
	}
}

// TestSubmitOrderAndPoolIsolation 覆盖满队列后的 FIFO 和同 ID 在不同池之间的隔离。
func TestSubmitOrderAndPoolIsolation(t *testing.T) {
	p, other := newTestPool(t), newTestPool(t)
	blocked, unblock := make(chan struct{}), make(chan struct{})
	var release sync.Once
	t.Cleanup(func() { release.Do(func() { close(unblock) }) })
	if err := p.Submit(1, func() { close(blocked); <-unblock }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}
	for range 2 {
		if err := p.Submit(1, func() {}); err != nil {
			t.Fatal(err)
		}
	}
	completed := make(chan int, 100)
	failed := make(chan error, 101)
	for i := range 100 {
		Submit(p, 1, func() { completed <- i }, func(err error) { failed <- err })
	}
	independent := make(chan struct{})
	Submit(other, 1, func() { close(independent) }, func(err error) { failed <- err })
	select {
	case <-independent:
	case err := <-failed:
		t.Fatal(err)
	case <-time.After(time.Second):
		t.Fatal("one pool blocked another pool with the same ID")
	}
	release.Do(func() { close(unblock) })
	deadline := time.After(time.Second)
	for want := range 100 {
		select {
		case got := <-completed:
			if got != want {
				t.Fatalf("submission order: got %d, want %d", got, want)
			}
		case err := <-failed:
			t.Fatal(err)
		case <-deadline:
			t.Fatal("queued submissions did not drain")
		}
	}
	waitIdle(t)
}

// TestSubmitClosedPool 确认提交失败会逐项报告，并释放已关闭池的派发状态。
func TestSubmitClosedPool(t *testing.T) {
	p := newTestPool(t)
	p.Release()
	failed := make(chan error, 3)
	for range 3 {
		Submit(p, 1, func() { failed <- errors.New("closed pool ran a callback") }, func(err error) { failed <- err })
	}
	for range 3 {
		select {
		case err := <-failed:
			if !errors.Is(err, ants.ErrPoolClosed) {
				t.Fatalf("unexpected submit result: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("failed submission was not reported")
		}
	}
	waitIdle(t)
}

// TestClosedQueueIsRejectedBeforeReboot 确认已观察到关闭的队列不会继续投递到新一轮池。
func TestClosedQueueIsRejectedBeforeReboot(t *testing.T) {
	p := newTestPool(t)
	started, unblock := make(chan struct{}), make(chan struct{})
	resumeError := make(chan struct{})
	var releaseWorker, releaseError sync.Once
	t.Cleanup(func() {
		releaseWorker.Do(func() { close(unblock) })
		releaseError.Do(func() { close(resumeError) })
	})
	if err := p.Submit(1, func() { close(started); <-unblock }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}
	for range 2 {
		if err := p.Submit(1, func() {}); err != nil {
			t.Fatal(err)
		}
	}
	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	Submit(p, 1, func() { firstResult <- nil }, func(err error) {
		firstResult <- err
		<-resumeError
	})
	Submit(p, 1, func() { secondResult <- nil }, func(err error) { secondResult <- err })
	p.Release()
	select {
	case err := <-firstResult:
		if !errors.Is(err, ants.ErrPoolClosed) {
			t.Fatalf("first submission: got %v, want ErrPoolClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("blocked submission was not rejected")
	}
	// 在首个错误回调暂停时重启，证明其余旧任务必须在报告错误前从队列摘除。
	releaseWorker.Do(func() { close(unblock) })
	rebooted := make(chan struct{})
	go func() { p.Reboot(); close(rebooted) }()
	select {
	case <-rebooted:
	case <-time.After(time.Second):
		t.Fatal("pool did not reboot")
	}
	releaseError.Do(func() { close(resumeError) })
	select {
	case err := <-secondResult:
		if !errors.Is(err, ants.ErrPoolClosed) {
			t.Fatalf("old pending submission crossed reboot: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("pending submission was not rejected")
	}
	// 旧 drain 退出后不能删除后续新提交建立的队列。
	Submit(p, 1, func() { secondResult <- nil }, func(err error) { secondResult <- err })
	select {
	case err := <-secondResult:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("new submission did not run")
	}
	waitIdle(t)
}
