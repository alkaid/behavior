package behavior

import (
	"math/rand"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/alkaid/behavior/logger"
	"go.uber.org/zap"

	"github.com/alkaid/behavior/bcore"
)

func help() {
	InitSystem(WithLogDevelopment(true), WithLogLevel(zapcore.DebugLevel))
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

func TestRunTree_ThreadSafe(t *testing.T) {
	help()
	// 	content := `
	// {"root":"3bNAvYqyRBvbzJKHTan2FME42A1","nodes":{"b80d6ydKHFHN4mYTCfFeFDDB697B":{"id":"b80d6ydKHFHN4mYTCfFeFDDB697B","name":"Selector","title":"打牌场景","category":"composite","children":["30abbPs1ktNBKh96yn8dwj2F64DA","05f6djYn2pOQ5r0h9lEccYAA501E"],"properties":{},"delegator":{"target":"","method":"","script":""}},"30abbPs1ktNBKh96yn8dwj2F64DA":{"id":"30abbPs1ktNBKh96yn8dwj2F64DA","name":"BBCondition","title":"是否自己操作","category":"decorator","children":["becf7thI19O2K973s5Yb7xn8CD25"],"properties":{"operator":2,"key":"gameDdz.operate.bool","value":true,"abortMode":3},"delegator":{"target":"","method":"","script":""}},"becf7thI19O2K973s5Yb7xn8CD25":{"id":"becf7thI19O2K973s5Yb7xn8CD25","name":"Sequence","title":"出牌操作","category":"composite","children":["05f6djYn2pOQ5r0h9lEccYA563AC","ae1d56a8BhFrJII8Z+xvZOI1FDF7","ae1d56a8BhFrJII8Z+xvZOI16A5E","bfc337GZiFHa6LmtDp4VmsA877EF"],"properties":{},"delegator":{"target":"","method":"","script":""}},"05f6djYn2pOQ5r0h9lEccYA563AC":{"id":"05f6djYn2pOQ5r0h9lEccYA563AC","name":"Wait","title":"出牌任务等待时间","category":"task","children":[],"properties":{"waitTime":"3s","randomDeviation":"1s","forever":false},"delegator":{"target":"","method":"","script":""}},"ae1d56a8BhFrJII8Z+xvZOI1FDF7":{"id":"ae1d56a8BhFrJII8Z+xvZOI1FDF7","name":"Action","title":"出牌","category":"task","children":[],"properties":{},"delegator":{"target":"GameDdz","method":"ReqOutCard","script":""}},"ae1d56a8BhFrJII8Z+xvZOI16A5E":{"id":"ae1d56a8BhFrJII8Z+xvZOI16A5E","name":"Action","title":"清理收到出牌提示状态","category":"task","children":[],"properties":{},"delegator":{"target":"GameDdz","method":"ModifyOperateTipsState","script":""}},"bfc337GZiFHa6LmtDp4VmsA877EF":{"id":"bfc337GZiFHa6LmtDp4VmsA877EF","name":"Succeeded","title":"Succeeded","category":"decorator","children":["05f6djYn2pOQ5r0h9lEccYA802F5"],"properties":{},"delegator":{"target":"","method":"","script":""}},"05f6djYn2pOQ5r0h9lEccYA802F5":{"id":"05f6djYn2pOQ5r0h9lEccYA802F5","name":"Wait","title":"轮到自己打牌等待","category":"task","children":[],"properties":{"waitTime":"","randomDeviation":"","forever":true},"delegator":{"target":"","method":"","script":""}},"05f6djYn2pOQ5r0h9lEccYAA501E":{"id":"05f6djYn2pOQ5r0h9lEccYAA501E","name":"Wait","title":"打牌等待","category":"task","children":[],"properties":{"waitTime":"","randomDeviation":"","forever":true},"delegator":{"target":"","method":"","script":""}},"3bNAvYqyRBvbzJKHTan2FME42A1":{"id":"3bNAvYqyRBvbzJKHTan2FME42A1","name":"Root","category":"decorator","title":"Root","properties":{"once":false,"interval":""},"delegator":{},"children":["b80d6ydKHFHN4mYTCfFeFDDB697B"]}},"tag":"Play"}
	// `
	// 	content := `
	// {"root":"3eQiXf9jhDWLO8fQk4nf6820C28","nodes":{"b80d6ydKHFHN4mYTCfFeFDD1B4BA":{"id":"b80d6ydKHFHN4mYTCfFeFDD1B4BA","name":"Selector","title":"打牌场景","category":"composite","children":["30abbPs1ktNBKh96yn8dwj27C6C7","05f6djYn2pOQ5r0h9lEccYA7D876"],"properties":{},"delegator":{"target":"","method":"","script":""}},"30abbPs1ktNBKh96yn8dwj27C6C7":{"id":"30abbPs1ktNBKh96yn8dwj27C6C7","name":"BBCondition","title":"是否自己操作","category":"decorator","children":["becf7thI19O2K973s5Yb7xn07E1D"],"properties":{"operator":2,"key":"gameDdz.operate.bool","value":true,"abortMode":3},"delegator":{"target":"","method":"","script":""}},"becf7thI19O2K973s5Yb7xn07E1D":{"id":"becf7thI19O2K973s5Yb7xn07E1D","name":"Sequence","title":"出牌操作","category":"composite","children":["05f6djYn2pOQ5r0h9lEccYA6EF80","ae1d56a8BhFrJII8Z+xvZOI2E0FE","ae1d56a8BhFrJII8Z+xvZOIFBCD6"],"properties":{},"delegator":{"target":"","method":"","script":""}},"05f6djYn2pOQ5r0h9lEccYA6EF80":{"id":"05f6djYn2pOQ5r0h9lEccYA6EF80","name":"Wait","title":"出牌任务等待时间","category":"task","children":[],"properties":{"waitTime":"2s","randomDeviation":"2s","forever":false},"delegator":{"target":"","method":"","script":""}},"ae1d56a8BhFrJII8Z+xvZOI2E0FE":{"id":"ae1d56a8BhFrJII8Z+xvZOI2E0FE","name":"Action","title":"出牌","category":"task","children":[],"properties":{},"delegator":{"target":"GameDdz","method":"ReqOutCard","script":""}},"ae1d56a8BhFrJII8Z+xvZOIFBCD6":{"id":"ae1d56a8BhFrJII8Z+xvZOIFBCD6","name":"Action","title":"清理收到出牌提示状态","category":"task","children":[],"properties":{},"delegator":{"target":"GameDdz","method":"ModifyOperateTipsState","script":""}},"05f6djYn2pOQ5r0h9lEccYA7D876":{"id":"05f6djYn2pOQ5r0h9lEccYA7D876","name":"Wait","title":"打牌等待","category":"task","children":[],"properties":{"waitTime":"","randomDeviation":"","forever":true},"delegator":{"target":"","method":"","script":""}},"3eQiXf9jhDWLO8fQk4nf6820C28":{"id":"3eQiXf9jhDWLO8fQk4nf6820C28","name":"Root","category":"decorator","title":"Root","properties":{"once":false,"interval":""},"delegator":{},"children":["b80d6ydKHFHN4mYTCfFeFDD1B4BA"]}},"tag":"Play"}
	// `
	content := `
{"root":"3eQiXf9jhDWLO8fQk4nf68EDF39","nodes":{"c5c0cpdvsxCrZHDjcjTOnU273B66":{"id":"c5c0cpdvsxCrZHDjcjTOnU273B66","name":"Repeater","title":"Repeater","category":"decorator","children":["b80d6ydKHFHN4mYTCfFeFDD64C2D"],"properties":{"times":0},"delegator":{"target":"","method":"","script":""}},"b80d6ydKHFHN4mYTCfFeFDD64C2D":{"id":"b80d6ydKHFHN4mYTCfFeFDD64C2D","name":"Selector","title":"打牌场景","category":"composite","children":["30abbPs1ktNBKh96yn8dwj256DE4","05f6djYn2pOQ5r0h9lEccYA0B081"],"properties":{},"delegator":{"target":"","method":"","script":""}},"30abbPs1ktNBKh96yn8dwj256DE4":{"id":"30abbPs1ktNBKh96yn8dwj256DE4","name":"BBCondition","title":"是否自己操作","category":"decorator","children":["becf7thI19O2K973s5Yb7xn7184D"],"properties":{"operator":2,"key":"gameDdz.operate.bool","value":true,"abortMode":3},"delegator":{"target":"","method":"","script":""}},"becf7thI19O2K973s5Yb7xn7184D":{"id":"becf7thI19O2K973s5Yb7xn7184D","name":"Sequence","title":"出牌操作","category":"composite","children":["05f6djYn2pOQ5r0h9lEccYA7C352","ae1d56a8BhFrJII8Z+xvZOI31538","ae1d56a8BhFrJII8Z+xvZOIBAF91"],"properties":{},"delegator":{"target":"","method":"","script":""}},"05f6djYn2pOQ5r0h9lEccYA7C352":{"id":"05f6djYn2pOQ5r0h9lEccYA7C352","name":"Wait","title":"出牌任务等待时间","category":"task","children":[],"properties":{"waitTime":"2s","randomDeviation":"2s","forever":false},"delegator":{"target":"","method":"","script":""}},"ae1d56a8BhFrJII8Z+xvZOI31538":{"id":"ae1d56a8BhFrJII8Z+xvZOI31538","name":"Action","title":"出牌","category":"task","children":[],"properties":{},"delegator":{"target":"GameDdz","method":"ReqOutCard","script":""}},"ae1d56a8BhFrJII8Z+xvZOIBAF91":{"id":"ae1d56a8BhFrJII8Z+xvZOIBAF91","name":"Action","title":"清理收到出牌提示状态","category":"task","children":[],"properties":{},"delegator":{"target":"GameDdz","method":"ModifyOperateTipsState","script":""}},"05f6djYn2pOQ5r0h9lEccYA0B081":{"id":"05f6djYn2pOQ5r0h9lEccYA0B081","name":"Wait","title":"打牌等待","category":"task","children":[],"properties":{"waitTime":"","randomDeviation":"","forever":true},"delegator":{"target":"","method":"","script":""}},"3eQiXf9jhDWLO8fQk4nf68EDF39":{"id":"3eQiXf9jhDWLO8fQk4nf68EDF39","name":"Root","category":"decorator","title":"Root","properties":{"once":false,"interval":""},"delegator":{},"children":["c5c0cpdvsxCrZHDjcjTOnU273B66"]}},"tag":"Play"}
`
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"thread", content, false},
	}
	fch := make(chan *bcore.FinishEvent, 10)
	RegisterDelegatorType(NameGameDdz, &GameDdz{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GlobalTreeRegistry().LoadFromJson([]byte(tt.content)); (err != nil) != tt.wantErr {
				logger.Log.Error("", zap.Error(err))
				t.Errorf("LoadFromPaths() error = %v, wantErr %v", err, tt.wantErr)
			}
			for j := 1; j < 4; j++ {
				ddz := &GameDdz{}
				brain := NewBrain(bcore.NewBlackboard(j, nil), map[string]any{NameGameDdz: ddz}, fch)
				ddz.brain = brain
				brain.Run("Play", false)

				go func(brain bcore.IBrain) {
					for i := 0; i < 1000; i++ {
						if i%100 == 0 {
							logger.Sugar.Info(i)
						}
						delay := time.Duration(rand.Float32() * float32(time.Second) * 5)
						time.Sleep(delay)
						brain.Blackboard().Set(BBKeyGameDdzOperateTipsBool, true)
					}
				}(brain)
			}
		})
	}
	<-fch
}

const NameGameDdz = "GameDdz"
const (
	BBKeyGameDdzOperateTipsBool = "gameDdz.operate.bool" // 是否自己收到操作提示

)

type GameDdz struct {
	brain bcore.IBrain
}

func (g *GameDdz) ReqOutCard() error {
	return nil
}

func (g *GameDdz) ModifyOperateTipsState() error {
	g.brain.Blackboard().Del(BBKeyGameDdzOperateTipsBool)
	return nil
}
