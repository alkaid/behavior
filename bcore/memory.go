package bcore

import (
	"errors"

	"github.com/alkaid/timingwheel"

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

// NodeMemory 节点数据
type NodeMemory struct {
	Ext           Memory             // 扩展数据,给框架之外的自定义节点使用
	State         NodeState          // 节点状态
	MountParent   IContainer         // 父节点,仅 Root 有效
	Observing     bool               // 是否监听中,仅 ObservingDecorator 及其派生类有效
	CurrentChild  int                // 当前运行中的子节点索引,若组合节点支持子节点随机,则该字段的意义为组合节点当前step index
	ChildrenOrder []int              // 孩子节点排序索引
	Parallel      *ParallelMemory    // 并发节点的数据
	CronTask      *timingwheel.Timer // 定时任务
}

func NewNodeMemory() *NodeMemory {
	return &NodeMemory{Ext: map[string]any{}}
}

// ParallelMemory 并发节点的数据
type ParallelMemory struct {
	RunningCount      int             // 运行中的子节点数量
	SucceededCount    int             // 成功的子节点数量
	FailedCount       int             // 失败的子节点数量
	ChildrenSucceeded map[string]bool // 子节点是否成功
	Succeeded         bool            // 自己是否成功
	ChildrenAborted   bool            // 是否中断子节点
}
