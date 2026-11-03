// Package dispatch 将框架的异步投递与 ants 的阻塞准入分开。
package dispatch

import (
	"errors"
	"sync"

	"github.com/alkaid/ants/v2"
)

type queueKey struct {
	pool *ants.PoolWithID
	id   int
}

type submission struct {
	task    func()
	onError func(error)
}

var (
	mu     sync.Mutex
	queues = make(map[queueKey][]submission)
)

// Submit 按同一池、同一 ID 的调用顺序异步投递，实际回调仍在 ants worker 执行。
// 每个有待投递任务的 ID 只有一个派发 goroutine，清空后退出；onError 必须非 nil。
// 不在 AI 线程等待队列容量，否则黑板通知或嵌套 Brain.Go 会等待自身释放容量。
func Submit(pool *ants.PoolWithID, id int, task func(), onError func(error)) {
	key := queueKey{pool: pool, id: id}
	mu.Lock()
	pending, running := queues[key]
	queues[key] = append(pending, submission{task: task, onError: onError})
	mu.Unlock()
	if !running {
		go drain(key)
	}
}

func drain(key queueKey) {
	for {
		mu.Lock()
		pending := queues[key]
		if len(pending) == 0 {
			delete(queues, key)
			mu.Unlock()
			return
		}
		next := pending[0]
		pending[0] = submission{} // 释放已投递闭包，避免切片底层数组继续持有 AI。
		queues[key] = pending[1:]
		mu.Unlock()

		// Submit 等待时仍保留队列所有权，后续调用只追加，不能另起派发者越序。
		if err := key.pool.Submit(key.id, next.task); err != nil {
			if errors.Is(err, ants.ErrPoolClosed) {
				// 先摘除整条旧队列，再报告失败；错误回调期间池可能已经重启。
				// 后续 Submit 可建立新队列，当前 drain 不再访问该 key。
				mu.Lock()
				pending = queues[key]
				delete(queues, key)
				mu.Unlock()
				next.onError(err)
				for index, rejected := range pending {
					pending[index] = submission{}
					rejected.onError(err)
				}
				return
			}
			next.onError(err)
		}
	}
}
