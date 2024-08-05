package behavior

import (
	"github.com/alkaid/behavior/bcore"
	"math"
)

// TODO 全局 nodeTypeMap

var sharedBlackboard = make(map[string]*bcore.Blackboard) // 全局共享黑板

// SharedBlackboardID TODO 后期改为配置
const SharedBlackboardID = math.MaxInt - 1 // 共享黑板ID

// SharedBlackboard 获取全局共享黑板单例,若不存在会创建一个
//
//	@param key 指定key
//	@return *Blackboard
func SharedBlackboard(key string) *bcore.Blackboard {
	v, ok := sharedBlackboard[key]
	if !ok {
		v = bcore.NewBlackboard(SharedBlackboardID, nil)
		sharedBlackboard[key] = v
	}
	return v
}

// DefaultSharedBlackboard 获取全局默认共享黑板单例
//
//	@return *Blackboard
func DefaultSharedBlackboard() *bcore.Blackboard {
	return SharedBlackboard("default")
}
