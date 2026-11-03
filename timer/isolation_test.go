package timer

import (
	"sync"
	"testing"
	"time"

	"github.com/alkaid/ants/v2"
	"github.com/zeromicro/go-zero/core/collection"
)

// manualTicker 把同槽到期固定在一个 tick，避免依赖两个真实定时器碰巧同时触发。
type manualTicker struct{ ticks chan time.Time }

func (t manualTicker) Chan() <-chan time.Time { return t.ticks }
func (t manualTicker) Stop()                  {}

// TestFullQueueDoesNotBlockOtherTimer 保证满队列只延迟所属 AI，不阻塞同批的其他 AI。
func TestFullQueueDoesNotBlockOtherTimer(t *testing.T) {
	p, err := ants.NewPoolWithID(2, ants.WithTaskBuffer(1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Release)
	blocked, unblock := make(chan struct{}), make(chan struct{})
	var release sync.Once
	t.Cleanup(func() { release.Do(func() { close(unblock) }) })
	if err = p.Submit(1, func() { close(blocked); <-unblock }); err != nil {
		t.Fatal(err)
	}
	await(t, blocked)
	// ants 的物理队列容量为 2 * TaskBuffer。
	for range 2 {
		if err = p.Submit(1, func() {}); err != nil {
			t.Fatal(err)
		}
	}
	ticker := manualTicker{ticks: make(chan time.Time)}
	wheel, err := collection.NewTimingWheelWithTicker(time.Millisecond, 8, func(_, value any) {
		value.(*Timer).dispatch()
	}, ticker)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(wheel.Stop)
	first := newTimer(wheel, time.Nanosecond, false, func() {}, WithPool(p), WithGoID(1))
	t.Cleanup(func() { first.Stop() })
	done := make(chan struct{})
	second := newTimer(wheel, time.Nanosecond, false, func() { close(done) }, WithPool(p), WithGoID(2))
	t.Cleanup(func() { second.Stop() })
	ticker.ticks <- time.Now()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("full ID queue blocked another timer in the same tick")
	}
}
