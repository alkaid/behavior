package bcore

import (
	"errors"
	"time"

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
	Ext                 Memory     // 扩展数据,给框架之外的自定义节点使用
	State               NodeState  // 节点状态
	MountParent         IContainer // 动态挂载的父节点,仅 Root 有效
	DynamicChild        INode      // 动态挂载的子节点
	RequestDynamicChild INode      // 请求挂载的子节点,延迟到旧的分支执行完成后才真正挂载到 DynamicChild
	Observing           bool       // 是否监听中,仅 ObservingDecorator 及其派生类有效
	// 根据节点类型意义不同:
	//  1.非随机组合节点:当前运行中的子节点索引;
	//  2.随机组合节点:完成了几个子节点;
	//  3.循环装饰器:当前为第几次循环
	CurrIndex        int
	ChildrenOrder    []int              // 孩子节点排序索引
	Parallel         *ParallelMemory    // 并发节点的数据
	CronTask         *timingwheel.Timer // 定时任务
	DefaultObserver  Observer           // 默认监听函数
	Cooling          bool               // 是否cd中
	LimitReached     bool               // 是否达到限制
	DecoratedDone    bool               // 被装饰节点是否完成
	DecoratedSuccess bool               // 被装饰节点是否成功
	Elapsed          time.Duration      // 启动后流逝的时间
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
