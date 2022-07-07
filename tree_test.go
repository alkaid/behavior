package behavior

import (
	"testing"

	"github.com/alkaid/behavior/logger"
	"go.uber.org/zap"

	"github.com/alkaid/behavior/bcore"
)

func help() {
	InitSystem(WithLogDevelopment(true))
}

func TestTreeRegistry_LoadFromPaths(t *testing.T) {
	help()
	tests := []struct {
		name    string
		paths   []string
		wantErr bool
	}{
		{"test1", []string{"/home/alkaid/bt1.json"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTreeRegistry()
			if err := r.LoadFromPaths(tt.paths); (err != nil) != tt.wantErr {
				logger.Log.Error("", zap.Error(err))
				t.Errorf("LoadFromPaths() error = %v, wantErr %v", err, tt.wantErr)
			}
			for _, tree := range r.TreesByID {
				bcore.Print(tree.Root, NewBrain(bcore.NewBlackboard(1, nil), nil, nil))
			}
		})
	}
}
