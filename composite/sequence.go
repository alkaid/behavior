package composite

import "github.com/alkaid/behavior"

// Sequence 序列,节点按从左到右的顺序执行其子节点。当其中一个子节点失败时，序列节点也将停止执行。如果有子节点失败，那么序列就会失败。如果该序列的所有子节点运行都成功执行，则序列节点成功。
type Sequence struct {
	NonParallel
}

// SuccessMode @implement INonParallelWorker.SuccessMode
//  @receiver s
//  @param brain
//  @return behavior.FinishMode
func (s *Sequence) SuccessMode(brain behavior.IBrain) behavior.FinishMode {
	return behavior.FinishModeAll
}
