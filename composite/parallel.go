package composite

import (
	"github.com/alkaid/behavior/bcore"
	"go.uber.org/zap"
)

type IParallelProperties interface {
	GetSuccessPolicy() bcore.FinishMode
	GetFailurePolicy() bcore.FinishMode
}

type ParallelProperties struct {
	bcore.BaseProperties
	SuccessPolicy bcore.FinishMode `json:"successPolicy"` // 成功策略
	FailurePolicy bcore.FinishMode `json:"failurePolicy"` // 失败策略
}

// Parallel 并行组合基类,节点按从左到右的顺序根据结束模式决定完成时机
// 	并行执行所有子节点，根据成功原则和失败原则，决定节点停用时机
// 	| 成功原则 | 失败原则 |
// 	| ParallelProperties.SuccessPolicy | ParallelProperties.FailurePolicy |
//
// 	| behavior.FinishModeOne | behavior.FinishModeOne |
// 	第一个子节点停用返回true(false)后，关闭所有子节点，停用当前节点返回true(false)
//
// 	| behavior.FinishModeOne | behavior.FinishModeAll |
// 	当有一个子节点停用返回true后，关闭所有子节点，停用当前节点返回true，否则返回false
//
// 	| behavior.FinishModeAll | behavior.FinishModeOne |
// 	所有子节点停用返回true后，当前节点停用返回true，否则返回false
//
// 	| behavior.FinishModeAll | behavior.FinishModeAll |
// 	所有子节点停用返回true后，当前节点停用返回true，否则返回false
type Parallel struct {
	bcore.Composite
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (p *Parallel) PropertiesClassProvider() any {
	return &ParallelProperties{}
}

func (p *Parallel) PMemory(brain bcore.IBrain) *bcore.ParallelMemory {
	return brain.Blackboard().(bcore.IBlackboardInternal).NodeMemory(p.ID()).Parallel
}

// OnStart
//  @override Node.OnStart
//  @receiver n
//  @param brain
func (p *Parallel) OnStart(brain bcore.IBrain) {
	p.Composite.OnStart(brain)
	// 初始化状态
	brain.Blackboard().(bcore.IBlackboardInternal).NodeMemory(p.ID()).Parallel = &bcore.ParallelMemory{
		ChildrenSucceeded: map[string]bool{},
	}
	for _, child := range p.Children() {
		if !child.IsInactive(brain) {
			p.Log().Fatal("child must be inactive", zap.String("child", child.String(brain)))
		}
	}
	for _, child := range p.Children() {
		p.PMemory(brain).RunningCount++
		child.Start(brain)
	}
}

// OnAbort
//  @override Node.OnAbort
//  @receiver n
//  @param brain
func (p *Parallel) OnAbort(brain bcore.IBrain) {
	memory := p.PMemory(brain)
	allChildrenStarted := len(p.Children()) == memory.RunningCount+memory.SucceededCount+memory.FailedCount
	if !allChildrenStarted {
		p.Log().Fatal("parallel status error", zap.Int("runningCount", memory.RunningCount), zap.Int("succeededCount", memory.SucceededCount), zap.Int("failedCount", memory.FailedCount))
	}
	// 向下传播给当前活跃子节点
	for _, child := range p.Children() {
		if child.IsActive(brain) {
			child.Abort(brain)
		}
	}
}

// OnChildFinished
//  @override Container.OnChildFinished
//  @receiver r
//  @param brain
//  @param child
//  @param succeeded
// nolint
func (p *Parallel) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	p.Composite.OnChildFinished(brain, child, succeeded)
	memory := p.PMemory(brain)
	memory.RunningCount--
	if succeeded {
		memory.SucceededCount++
	} else {
		memory.FailedCount++
	}
	memory.ChildrenSucceeded[child.ID()] = succeeded
	allChildrenStarted := len(p.Children()) == memory.RunningCount+memory.SucceededCount+memory.FailedCount
	if !allChildrenStarted {
		return
	}
	// 执行策略 逻辑见最前面 Parallel 的说明
	if memory.RunningCount == 0 {
		if !memory.ChildrenAborted {
			if p.Properties().(IParallelProperties).GetFailurePolicy() == bcore.FinishModeOne && memory.FailedCount > 0 {
				memory.Succeeded = false
			} else if p.Properties().(IParallelProperties).GetSuccessPolicy() == bcore.FinishModeOne && memory.SucceededCount > 0 {
				memory.Succeeded = true
			} else if p.Properties().(IParallelProperties).GetSuccessPolicy() == bcore.FinishModeAll && memory.SucceededCount == len(p.Children()) {
				memory.Succeeded = true
			} else {
				memory.Succeeded = false
			}
		}
		p.Finish(brain, memory.Succeeded)
	} else if !memory.ChildrenAborted {
		if memory.SucceededCount == len(p.Children()) {
			p.Log().Fatal("succeeded count error")
		}
		if memory.FailedCount == len(p.Children()) {
			p.Log().Fatal("failed count error")
		}
		if p.Properties().(IParallelProperties).GetFailurePolicy() == bcore.FinishModeOne && memory.FailedCount > 0 {
			memory.Succeeded = false
			memory.ChildrenAborted = true
		} else if p.Properties().(IParallelProperties).GetSuccessPolicy() == bcore.FinishModeOne && memory.SucceededCount > 0 {
			memory.Succeeded = true
			memory.ChildrenAborted = true
		}
		if memory.ChildrenAborted {
			for _, node := range p.Children() {
				if node.IsActive(brain) {
					node.Abort(brain)
				}
			}
		}
	}
}

// AbortLowerPriorityChildrenForChild
//  @implement IComposite.AbortLowerPriorityChildrenForChild
//  @receiver c
//  @param childAbortBy
func (p *Parallel) AbortLowerPriorityChildrenForChild(brain bcore.IBrain, childAbortBy bcore.INode) {
	if !childAbortBy.IsActive(brain) {
		p.Log().Fatal("child must be active")
	}
	memory := p.PMemory(brain)
	if memory.ChildrenSucceeded[childAbortBy.ID()] {
		memory.SucceededCount--
	} else {
		memory.FailedCount--
	}
	memory.RunningCount++
	childAbortBy.Start(brain)
}
