package timer

import (
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/alkaid/timingwheel"
)

func helper() {
	rand.Seed(time.Now().UnixNano())
	InitPool(1, 10*time.Millisecond, 100)
}

func TestAfter(t *testing.T) {
	helper()
	var wg sync.WaitGroup
	type args struct {
		interval        time.Duration
		randomDeviation time.Duration
		task            func()
	}
	tests := []struct {
		name string
		args args
	}{
		{"t1", args{
			interval:        time.Second * 3,
			randomDeviation: time.Second * 6,
			task: func() {
				t.Log("t1 done")
				wg.Done()
			},
		}},
	}
	for _, tt := range tests {
		wg.Add(1)
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 10; i++ {
				st := time.Now()
				After(tt.args.interval, tt.args.randomDeviation, func() {
					t.Log(time.Since(st))
				}, timingwheel.WithGoID(1))
			}
		})
	}
	wg.Wait()
}

func TestCron(t *testing.T) {
	type args struct {
		interval        time.Duration
		randomDeviation time.Duration
		task            func()
		opts            []timingwheel.Option
	}
	tests := []struct {
		name string
		args args
		want *timingwheel.Timer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Cron(tt.args.interval, tt.args.randomDeviation, tt.args.task, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cron() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitPool(t *testing.T) {
	type args struct {
		size     int
		interval time.Duration
		numSlots int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitPool(tt.args.size, tt.args.interval, tt.args.numSlots)
		})
	}
}

func TestTimeWheelInstance(t *testing.T) {
	tests := []struct {
		name string
		want *timingwheel.TimingWheel
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeWheelInstance(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TimeWheelInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}
