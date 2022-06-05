package behavior

import (
	"fmt"

	"github.com/alkaid/behavior/logger"
)

// Sequence 序列节点,节点按从左到右的顺序执行其子节点。当其中一个子节点失败时，序列节点也将停止执行。如果有子节点失败，那么序列就会失败。如果该序列的所有子节点运行都成功执行，则序列节点成功。
type Sequence struct {
	Composite
}

func (s *Sequence) currentChild(brain IBrain) int {
	return brain.Blackboard().(IBlackboardInternal).NodeData(s.id).currentChild
}
func (s *Sequence) setCurrentChild(brain IBrain, currChild int) {
	brain.Blackboard().(IBlackboardInternal).NodeData(s.id).currentChild = currChild
}

// OnStart
//  @override Node.OnStart
//  @receiver n
//  @param brain
func (s *Sequence) OnStart(brain IBrain) {
	s.Composite.OnStart(brain)
	for _, child := range s.children {
		if !child.IsInactive(brain) {
			logger.Log.Fatal("child must be inactive")
		}
	}
	s.setCurrentChild(brain, -1)
	s.processChildren(brain)
}

// OnAbort
//  @override Node.OnAbort
//  @receiver n
//  @param brain
func (s *Sequence) OnAbort(brain IBrain) {
	// 向下传播给当前活跃子节点
	s.children[s.currentChild(brain)].Abort(brain)
}

// OnChildFinished
//  @override Container.OnChildFinished
//  @receiver r
//  @param brain
//  @param child
//  @param succeeded
func (s *Sequence) OnChildFinished(brain IBrain, child INode, succeeded bool) {
	s.Composite.OnChildFinished(brain, child, succeeded)
	// 子节点不成功,则执行返回错误
	if !succeeded {
		s.Finish(brain, succeeded)
		return
	}
	// 成功则处理下一个
	s.processChildren(brain)
}

func (s *Sequence) processChildren(brain IBrain) {
	currChild := s.currentChild(brain) + 1
	s.setCurrentChild(brain, currChild)
	// 说明子节点全部成功,则自己也返回成功
	if currChild >= len(s.children) {
		s.Finish(brain, true)
		return
	}
	// 如果是被装饰器打断的,直接返回错误
	if s.IsAborting(brain) {
		s.Finish(brain, false)
		return
	}
	s.children[currChild].Start(brain)
}

// AbortLowerPriorityChildrenForChild
//  @implement IComposite.AbortLowerPriorityChildrenForChild
//  @receiver c
//  @param child
func (s *Sequence) AbortLowerPriorityChildrenForChild(brain IBrain, child INode) {
	idx := -1
	for i, currChild := range s.children {
		// 找到发起中断请求的子节点的索引
		if child.Id() == currChild.Id() {
			idx = i
		}
		// 找到正在运行的第一个右侧节点,中断掉
		if idx > 0 && currChild.IsActive(brain) {
			s.setCurrentChild(brain, idx-1)
			currChild.Abort(brain)
			break
		}
	}
}
func (s *Sequence) OnString(brain IBrain) string {
	return fmt.Sprintf("%s[%d]", s.Composite.OnString(brain), s.currentChild(brain))
}
