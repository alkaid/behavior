package composite

import "github.com/alkaid/behavior"

// RandomSequence 随机序列,节点按从左到右的顺序随机执行其子节点。当其中一个子节点失败时，序列节点也将停止执行。如果有子节点失败，那么序列就会失败。如果该序列的所有子节点运行都成功执行，则序列节点成功。
type RandomSequence struct {
	Sequence
	randomWorker *RandomWorker
}

// InitNodeWorker
//  @override Node.InitNodeWorker
//  @receiver c
//  @param worker
func (r *RandomSequence) InitNodeWorker(worker behavior.INodeWorker) {
	r.Sequence.InitNodeWorker(worker)
	r.randomWorker = NewRandomWorker(r)
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (r *RandomSequence) PropertiesClassProvider() any {
	return r.randomWorker.PropertiesClassProvider()
}

// OnOrder
//  @implement INonParallelWorker.OnOrder
//  @receiver r
//  @param brain
//  @param originChildrenOrder
//  @return orders
//  @return needOrder
func (r *RandomSequence) OnOrder(brain behavior.IBrain, originChildrenOrder []int) (orders []int, needOrder bool) {
	return r.randomWorker.OnOrder(brain, originChildrenOrder)
}
