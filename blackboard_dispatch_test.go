package behavior

import (
	"testing"
	"time"

	"github.com/alkaid/behavior/bcore"
	"github.com/alkaid/behavior/thread"
)

// TestBlackboardBatchInCallback 证明黑板通知不会因回调写入次数超过 ants 队列容量而锁住自身。
func TestBlackboardBatchInCallback(t *testing.T) {
	if err := thread.InitPool(nil); err != nil {
		t.Fatal(err)
	}
	board := bcore.NewBlackboard(903, nil)
	completed := make(chan int, 1)
	const updates = 300
	thread.GoByID(903, func() {
		board.Start()
		observer := &countObserver{}
		board.AddObserver("counter", observer)
		for i := range updates {
			board.Set("counter", i)
		}
		// 屏障必须排在所有通知后面；也检查通知没有被异步回退丢弃或越序。
		thread.GoByID(903, func() {
			board.Stop()
			completed <- observer.count
		})
	})
	select {
	case got := <-completed:
		if got != updates {
			t.Fatalf("delivered %d notifications, want %d", got, updates)
		}
	case <-time.After(time.Second):
		t.Fatal("blackboard updates blocked on their own notification queue")
	}
}

// countObserver 只在黑板所属线程使用，屏障负责将结果交回测试线程。
type countObserver struct{ count int }

func (o *countObserver) Fire(bcore.OpType, string, any, any) { o.count++ }
