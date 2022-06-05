package behavior

import (
	"errors"

	"github.com/alkaid/behavior/logger"
	"go.uber.org/zap"
)

var ErrConvertGenericType = errors.New("convert generic type error")

// Memory 为 Blackboard 提供内部使用的数据结构
type Memory = map[string]any

// ConvertAnyValue [T any] 转换 any 类型的值为传入的范型,一般配合map使用
//  @param v 值
//  @param ok 值是否有效
//  @return T
//  @return bool
//
func ConvertAnyValue[T any](v any, ok bool) (T, bool) {
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

// MapValue [T any] 从 Memory 中获取值
//  @param m
//  @param key
//  @return T
//  @return bool
func MapValue[T any](m Memory, key string) (T, bool) {
	v, ok := m[key]
	return ConvertAnyValue[T](v, ok)
}

// NodeData 节点数据
type NodeData struct {
	State        NodeState  // 节点状态
	MountParent  IContainer // 父节点,仅 Root 有效
	Observing    bool       // 是否监听中,仅 ObservingDecorator 及其派生类有效
	currentChild int        // 当前运行中的子节点索引
}
