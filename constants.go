package behavior

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

// 节点类别
const (
	CategoryComposite = "composite" // 组合节点
)

// 节点名称
const (
	NodeNameRoot = "Root" // root
)
