package composite

import "github.com/alkaid/behavior"

// RandomSelector 随机选择器.节点按从左到右的随机执行其子节点。当其中一个子节点执行成功时，选择器节点将停止执行。如果选择器的一个子节点成功运行，则选择器运行成功。如果选择器的所有子节点运行失败，则选择器运行失败。
type RandomSelector struct {
	Selector
	randomWorker *RandomWorker
}

// InitNodeWorker
//  @override Node.InitNodeWorker
//  @receiver c
//  @param worker
func (r *RandomSelector) InitNodeWorker(worker behavior.INodeWorker) {
	r.Selector.InitNodeWorker(worker)
	r.randomWorker = NewRandomWorker(r)
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (r *RandomSelector) PropertiesClassProvider() any {
	return r.randomWorker.PropertiesClassProvider()
}

// OnOrder
//  @implement INonParallelWorker.OnOrder
//  @receiver r
//  @param brain
//  @param originChildrenOrder
//  @return orders
//  @return needOrder
func (r *RandomSelector) OnOrder(brain behavior.IBrain, originChildrenOrder []int) (orders []int, needOrder bool) {
	return r.randomWorker.OnOrder(brain, originChildrenOrder)
}
