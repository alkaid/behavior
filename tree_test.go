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

func TestRunTree_LoadFromStr(t *testing.T) {
	help()
	content := `
{"root":"3bNAvYqyRBvbzJKHTan2FM56BD0","nodes":{"becf7thI19O2K973s5Yb7xnA357D":{"id":"becf7thI19O2K973s5Yb7xnA357D","name":"Sequence","title":"Sequence","category":"composite","children":["05f6djYn2pOQ5r0h9lEccYABD65F","ae1d56a8BhFrJII8Z+xvZOI0CB50"],"properties":{},"delegator":{"target":"","method":"","script":""}},"05f6djYn2pOQ5r0h9lEccYABD65F":{"id":"05f6djYn2pOQ5r0h9lEccYABD65F","name":"Wait","title":"Wait","category":"task","children":[],"properties":{"waitTime":"2s","randomDeviation":"1s","forever":false},"delegator":{"target":"","method":"","script":""}},"ae1d56a8BhFrJII8Z+xvZOI0CB50":{"id":"ae1d56a8BhFrJII8Z+xvZOI0CB50","name":"Action","title":"Action","category":"task","children":[],"properties":{},"delegator":{"target":"","method":"","script":""}},"3bNAvYqyRBvbzJKHTan2FM56BD0":{"id":"3bNAvYqyRBvbzJKHTan2FM56BD0","name":"Root","category":"decorator","title":"Root","properties":{"once":true,"interval":""},"delegator":{},"children":["becf7thI19O2K973s5Yb7xnA357D"]}},"tag":"testwait"}
`
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"testWait", content, false},
	}
	fch := make(chan *bcore.FinishEvent, 10)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GlobalTreeRegistry().LoadFromJson([]byte(tt.content)); (err != nil) != tt.wantErr {
				logger.Log.Error("", zap.Error(err))
				t.Errorf("LoadFromPaths() error = %v, wantErr %v", err, tt.wantErr)
			}
			brain := NewBrain(bcore.NewBlackboard(1, nil), nil, fch)
			brain.Run("testwait", false)
		})
	}
	<-fch
}
