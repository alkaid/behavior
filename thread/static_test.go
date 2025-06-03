package thread

import (
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
	wg.Wait()
}

func TestGoMain(t *testing.T) {
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
			GoMain(tt.args.task)
		})
	}
}

func TestInitPool(t *testing.T) {
	type args struct {
		p *ants.PoolWithID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InitPool(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("InitPool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReleaseTableGoPool(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReleaseTableGoPool()
		})
	}
}

func TestWaitByID(t *testing.T) {
	type args struct {
		goID int
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
			WaitByID(tt.args.goID, tt.args.task)
		})
	}
}

func TestWaitMain(t *testing.T) {
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
			WaitMain(tt.args.task)
		})
	}
}
