package blackboard

import (
	"errors"

	"github.com/alkaid/behavior/logger"
	"go.uber.org/zap"
)

var ErrConvertGenericType = errors.New("convert generic type error")

// Memory 为 Blackboard 提供内部使用的数据结构
type Memory = map[string]any

// get [T any] 从 Memory 中获取值
//  @param memory
//  @return T
func mapValue[T any](v any, ok bool) (T, bool) {
	var zero T
	if !ok {
		return zero, ok
	}
	result, ok := v.(T)
	if !ok {
		logger.Log.Error("", zap.Error(ErrConvertGenericType))
		return zero, ok
	}
	return result, true
}

func get[T any](m Memory, key string) (T, bool) {
	v, ok := m[key]
	return mapValue[T](v, ok)
}

// TreeMemory 作用域为<行为数域>的数据结构
type TreeMemory struct {
	memory Memory            // 树的数据
	nodes  map[string]Memory // 索引为节点ID,元素为对应节点的数据
}

func newTreeMemory() *TreeMemory {
	return &TreeMemory{
		memory: make(Memory),
		nodes:  make(map[string]Memory),
	}
}
