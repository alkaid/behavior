package bcore

import (
	"time"

	"github.com/alkaid/behavior/script"
)

const DefaultInterval = time.Millisecond * 50 // 行为树默认更新间隔

const (
	NodeStateInactive NodeState = iota // 非活跃
	NodeStateActive                    // 活跃
	NodeStateAborting                  // 正在中断
)

// Result 执行结果
type Result int

const (
	ResultFailed     Result = iota // 失败
	ResultSucceeded                // 成功
	ResultInProgress               // 进行中
)

// AbortMode 中断模式
type AbortMode int

const (
	AbortModeNone          AbortMode = iota // 不中止执行
	AbortModeSelf                           // 当条件不满足时，中断修饰的节点所在分支，执行比修饰的节点优先级更低的分支
	AbortModeLowerPriority                  // 当条件满足时，中断比修饰的节点优先级更低的分支，执行所在分支
	AbortModeBoth                           // 两种情况都中断: AbortModeSelf && AbortModeLowerPriority
)

// FinishMode 平行节点的完成模式
type FinishMode int

const (
	FinishModeOne FinishMode = iota // 只要有一个子节点完成就完成
	FinishModeAll                   // 全部子节点完成才完成
)

// EventType 行为树回调给委托的事件类型
type EventType int

const (
	EventTypeOnStart  EventType = iota // 节点启动时
	EventTypeOnUpdate                  // 节点更新时(瞬时节点无效)
	EventTypeOnAbort                   // 节点被打断时(瞬时节点无效)
)

// Operator 运算符
type Operator int

const (
	OperatorIsSet      Operator = iota // 是否设存在KEY
	OperatorIsNotSet                   // 是否不存在key
	OperatorIsEqual                    // 是否相等
	OperatorIsNotEqual                 // 是否不等
	OperatorIsGt                       // 是否大于
	OperatorIsGte                      // 是否大于等于
	OperatorIsLt                       // 是否小于
	OperatorIsLte                      // 是否小于等于
)

// DynamicBehaviorMode 动态子树中断模式,描述的是装饰器动态挂载子树时若旧子树正在运行应该如何处理
type DynamicBehaviorMode int

const (
	DynamicBehaviorModeRestart  DynamicBehaviorMode = iota // 动态挂载子树时,若正在运行则重启装饰器
	DynamicBehaviorModeContinue                            // 动态挂载子树时，不做任何行为,即不中断正在运行的子树
	DynamicBehaviorModeAbort                               // 态挂载子树时,若正在运行则中断装饰器
)

// 节点类别
const (
	CategoryComposite = "composite" // 组合节点
	CategoryDecorator = "decorator" // 组合节点
	CategoryTask      = "task"      // 组合节点
)

// 节点名称
const (
	NodeNameRoot           = "Root"
	NodeNameSubtree        = "Subtree"
	NodeNameDynamicSubtree = "DynamicSubtree"
)

func init() {
	// 注册常量到脚本执行环境
	script.RegisterApi(map[string]any{
		"ResultFailed":      ResultFailed,
		"ResultSucceeded":   ResultSucceeded,
		"ResultInProgress":  ResultInProgress,
		"EventTypeOnStart":  EventTypeOnStart,
		"EventTypeOnUpdate": EventTypeOnUpdate,
		"EventTypeOnAbort":  EventTypeOnAbort,
	})
}
