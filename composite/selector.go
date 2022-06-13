package composite

import (
	"github.com/alkaid/behavior/bcore"
)

// Selector 选择器.节点按从左到右的顺序执行其子节点。当其中一个子节点执行成功时，选择器节点将停止执行。如果选择器的一个子节点成功运行，则选择器运行成功。如果选择器的所有子节点运行失败，则选择器运行失败。
type Selector struct {
	NonParallel
}

// SuccessMode @implement INonParallelWorker.SuccessMode
//  @receiver s
//  @param brain
//  @return behavior.FinishMode
func (s *Selector) SuccessMode() bcore.FinishMode {
	return bcore.FinishModeOne
}
