package composite

import (
	"fmt"

	"github.com/alkaid/behavior/bcore"

	"go.uber.org/zap"
)

// NonParallel 非并行组合基类,节点按从左到右的顺序执行其子节点
type NonParallel struct {
	bcore.Composite
	INonParallelWorker
	needOrder bool
}

// INonParallelWorker 继承 NonParallel 时必须实现的接口
type INonParallelWorker interface {
	IChildrenOrder
	// SuccessMode  成功模式
	//  @return behavior.FinishMode
	SuccessMode() bcore.FinishMode
}
type IChildrenOrder interface {
	// OnOrder 为孩子节点索引排序, NonParallel.OnStart 里回调
	//  @param originChildrenOrder
	//  @return orders 排序后的索引
	//  @return needOrder 是否需要排序
	OnOrder(brain bcore.IBrain, originChildrenOrder []int) (orders []int, needOrder bool)
}

// CurrIdx 当前处理进度index
//  @receiver n
//  @param brain
//  @return int
func (n *NonParallel) CurrIdx(brain bcore.IBrain) int {
	return brain.Blackboard().(bcore.IBlackboardInternal).NodeMemory(n.ID()).CurrentChild
}
func (n *NonParallel) SetCurrIdx(brain bcore.IBrain, currChild int) {
	brain.Blackboard().(bcore.IBlackboardInternal).NodeMemory(n.ID()).CurrentChild = currChild
}

// CurrChildIdx 当前处理的子节点index
//  @receiver n
//  @param brain
//  @return int
func (n *NonParallel) CurrChildIdx(brain bcore.IBrain) int {
	nodeData := brain.Blackboard().(bcore.IBlackboardInternal).NodeMemory(n.ID())
	// 无需重新排序的节点,使用单例里的索引
	if !n.needOrder {
		return nodeData.CurrentChild
	}
	// 使用保存到黑板的排序
	return nodeData.ChildrenOrder[nodeData.CurrentChild]
}

// CurrChild 当前处理的子节点
//  @receiver n
//  @param brain
//  @return behavior.INode
func (n *NonParallel) CurrChild(brain bcore.IBrain) bcore.INode {
	return n.Children()[n.CurrChildIdx(brain)]
}

// OnOrder
//  @implement INonParallelWorker.OnOrder
//  @receiver n
//  @param originChildrenOrder
//  @return orders
//  @return needOrder
func (n *NonParallel) OnOrder(brain bcore.IBrain, originChildrenOrder []int) (orders []int, needOrder bool) {
	return originChildrenOrder, false
}

// InitNodeWorker
//  @override Node.InitNodeWorker
//  @receiver c
//  @param worker
func (n *NonParallel) InitNodeWorker(worker bcore.INodeWorker) {
	n.Composite.InitNodeWorker(worker)
	// 强转,由框架本身保证实例化时传进来的worker是自己(自己实现了 INonParallelWorker 接口,故强转不会panic)
	n.INonParallelWorker = worker.(INonParallelWorker)
}

// OnStart
//  @override Node.OnStart
//  @receiver n
//  @param brain
func (n *NonParallel) OnStart(brain bcore.IBrain) {
	n.Composite.OnStart(brain)
	for _, child := range n.Children() {
		if !child.IsInactive(brain) {
			n.Log().Fatal("child must be inactive", zap.String("child", child.String(brain)))
		}
	}
	n.SetCurrIdx(brain, -1)
	childrenOrder := make([]int, len(n.Children()))
	for i := 0; i < len(n.Children()); i++ {
		childrenOrder[i] = i
	}
	// 若子类返回需要重新排序,需要记录排序索引到黑板
	childrenOrder, n.needOrder = n.INonParallelWorker.OnOrder(brain, childrenOrder)
	if n.needOrder {
		brain.Blackboard().(bcore.IBlackboardInternal).NodeMemory(n.ID()).ChildrenOrder = childrenOrder
	}
	n.processChildren(brain)
}

// OnAbort
//  @override Node.OnAbort
//  @receiver n
//  @param brain
func (n *NonParallel) OnAbort(brain bcore.IBrain) {
	// 向下传播给当前活跃子节点
	n.CurrChild(brain).Abort(brain)
}

// OnChildFinished
//  @override Container.OnChildFinished
//  @receiver r
//  @param brain
//  @param child
//  @param succeeded
func (n *NonParallel) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	n.Composite.OnChildFinished(brain, child, succeeded)
	// 只要一个成功 || 只要一个失败  就结束
	if (n.SuccessMode() == bcore.FinishModeOne && succeeded) || (n.SuccessMode() == bcore.FinishModeAll && !succeeded) {
		n.Finish(brain, succeeded)
		return
	}
	// 成功则处理下一个
	n.processChildren(brain)
}

func (n *NonParallel) processChildren(brain bcore.IBrain) {
	currChild := n.CurrIdx(brain) + 1
	n.SetCurrIdx(brain, currChild)
	// 说明子节点全部成功/失败,则自己也返回成功/失败
	if currChild >= len(n.Children()) {
		// 模式为只要一个成功就成功的情况下,走到这里所有子节点都执行了一遍都还没返回，说明失败了.反之亦然
		successWhenAllProcessed := n.SuccessMode() == bcore.FinishModeAll
		n.Finish(brain, successWhenAllProcessed)
		return
	}
	// 如果是被装饰器打断的,直接返回失败
	if n.IsAborting(brain) {
		n.Finish(brain, false)
		return
	}
	n.CurrChild(brain).Start(brain)
}

// AbortLowerPriorityChildrenForChild
//  @implement IComposite.AbortLowerPriorityChildrenForChild
//  @receiver c
//  @param childAbortBy
func (n *NonParallel) AbortLowerPriorityChildrenForChild(brain bcore.IBrain, childAbortBy bcore.INode) {
	idx := -1
	for i, currChild := range n.Children() {
		// 找到发起中断请求的子节点的索引
		if childAbortBy.ID() == currChild.ID() {
			idx = i
		}
		// 找到正在运行的第一个右侧节点,中断掉
		if idx > 0 && currChild.IsActive(brain) {
			n.SetCurrIdx(brain, idx-1)
			currChild.Abort(brain)
			break
		}
	}
}
func (n *NonParallel) OnString(brain bcore.IBrain) string {
	return fmt.Sprintf("%s[%d]", n.Composite.OnString(brain), n.CurrChildIdx(brain))
}
