package behavior

import (
	"github.com/alkaid/behavior/bcore"
	"github.com/alkaid/behavior/thread"
)

// TODO 全局 nodeTypeMap

var sharedBlackboard = make(map[string]*bcore.Blackboard) // 全局共享黑板

// SharedBlackboard 获取全局共享黑板单例,若不存在会创建一个
//  @param key 指定key
//  @return *Blackboard
func SharedBlackboard(key string) *bcore.Blackboard {
	v, ok := sharedBlackboard[key]
	if !ok {
		v = bcore.NewBlackboard(thread.MainThreadID, nil)
		sharedBlackboard[key] = v
	}
	return v
}

// DefaultSharedBlackboard 获取全局默认共享黑板单例
//  @return *Blackboard
func DefaultSharedBlackboard() *bcore.Blackboard {
	return SharedBlackboard("default")
}
