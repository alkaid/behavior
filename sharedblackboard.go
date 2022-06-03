package behavior

import "github.com/alkaid/behavior/thread"

// TODO 全局 nodeTypeMap

const defaultNanoIDLen = 8

var sharedBlackboard = make(map[string]*Blackboard) // 全局共享黑板

// SharedBlackboard 获取全局共享黑板单例,若不存在会创建一个
//  @param key 指定key
//  @return *Blackboard
func SharedBlackboard(key string) *Blackboard {
	v, ok := sharedBlackboard[key]
	if !ok {
		v = NewBlackboard(thread.MainThreadID, nil)
		sharedBlackboard[key] = v
	}
	return v
}

// DefaultSharedBlackboard 获取全局默认共享黑板单例
//  @return *Blackboard
func DefaultSharedBlackboard() *Blackboard {
	return SharedBlackboard("default")
}
