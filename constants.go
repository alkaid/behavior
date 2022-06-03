package behavior

const (
	NodeStateInactive  NodeState = iota // 非活跃
	NodeStateActive                     // 活跃
	NodeStateCanceling                  // 正在取消
)

const (
	NodeNameRoot = "Root"
)

type Result int // 执行结果

const (
	ResultFailed     Result = iota // 失败
	ResultSucceeded                // 成功
	ResultInProgress               // 进行中
)
