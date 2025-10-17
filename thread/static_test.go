package thread

import (
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
)

func help(t *testing.T) {
	err := InitPool(nil)
	if err != nil {
		t.Error(err)
	}
}
func TestGo(t *testing.T) {
	type args struct {
		task func()
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Go(tt.args.task)
		})
	}
}

func Test_ThreadSafe(t *testing.T) {
	help(t)
	var wg sync.WaitGroup
	wg.Add(1)
	GoByID(1, func() {
		for i := 0; i < 5; i++ {
			t.Log("t1.1:", i)
			if i == 1 {
				wg.Add(1)
				GoByID(1, func() {
					for j := 0; j < 5; j++ {
						t.Log("t1.2:", j)
					}
					wg.Done()
				})
			}
		}
		wg.Done()
	})
	p, err := ants.NewPoolWithID(ants.DefaultAntsPoolSize, ants.WithExpiryDuration(time.Hour))
	if err != nil {
		t.Error(err)
	}
	wg.Add(1)
	p.Submit(1, func() {
		for i := 0; i < 5; i++ {
			t.Log("t4.1:", i)
		}
		wg.Done()
	})
	wg.Add(1)
	GoByID(2, func() {
		for i := 0; i < 5; i++ {
			t.Log("t2.1:", i)
		}
		wg.Done()
	})
	wg.Add(1)
	GoByID(3, func() {
		for i := 0; i < 5; i++ {
			t.Log("t3.1:", i)
		}
		wg.Done()
	})
	wg.Wait()
}
func TestGoByID(t *testing.T) {
	help(t)
	var wg sync.WaitGroup
	wg.Add(1)
	GoByID(3, func() {
		for i := 0; i < 5; i++ {
			t.Log("t3.1:", i)
		}
		wg.Done()
	})
	wg.Add(1)
	GoByID(3, func() {
		for i := 0; i < 5; i++ {
			t.Log("t3.1:", i)
		}
		wg.Done()
	})
	wg.Wait()
}

// Test_HugeAmount 测试大量协程情况下能否正常跑完
//
//	@param t
func Test_HugeAmount(t *testing.T) {
	help(t)
	var wg sync.WaitGroup
	var calc = func() {
		// 模拟一个耗时操作
		time.Sleep(time.Microsecond * time.Duration(rand.IntN(1000)))
		wg.Done()
	}
	for k := 0; k < 1000; k++ {
		wg.Add(1)
		GoByID(k, func() {
			wg.Add(1)
			GoByID(k+1, func() {
				for i := 0; i < 5; i++ {
					wg.Add(1)
					GoByID(i, calc)
				}
				calc()
			})
			wg.Add(1)
			GoByID(k+2, func() {
				for i := 0; i < 5; i++ {
					wg.Add(1)
					GoByID(i, calc)
				}
				calc()
			})
			wg.Add(1)
			GoByID(k+3, func() {
				for i := 0; i < 5; i++ {
					wg.Add(1)
					GoByID(i, calc)
				}
				calc()
			})
			calc()
		})
	}
	wg.Wait()
}
