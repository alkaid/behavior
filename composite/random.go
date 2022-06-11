package composite

import (
	"github.com/alkaid/behavior/bcore"
	"github.com/alkaid/behavior/wrand"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type IRandomCompositeProperties interface {
	GetWeight() []int
}

// RandomCompositeProperties 随机组合节点属性
type RandomCompositeProperties struct {
	Weight []int `json:"weight"`
}

func (r *RandomCompositeProperties) GetWeight() []int {
	return r.Weight
}

// RandomWorker 组合节点随机排序的委托
type RandomWorker struct {
	node bcore.IComposite
}

func NewRandomWorker(node bcore.IComposite) *RandomWorker {
	return &RandomWorker{node: node}
}

var _ IChildrenOrder = (*RandomWorker)(nil)

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (r *RandomWorker) PropertiesClassProvider() any {
	return &RandomCompositeProperties{}
}

// OnOrder
//  @implement INonParallelWorker.OnOrder
//  @receiver r
//  @param brain
//  @param originChildrenOrder
//  @return orders
//  @return needOrder
func (r *RandomWorker) OnOrder(brain bcore.IBrain, originChildrenOrder []int) (orders []int, needOrder bool) {
	// 根据权重属性排序,若没有配置,则随机
	weights := r.node.Properties().(IRandomCompositeProperties).GetWeight()
	if len(weights) == 0 {
		return lo.Shuffle(originChildrenOrder), true
	}
	// 不允许<=0,修正
	for i := 0; i < len(weights); i++ {
		if weights[i] <= 0 {
			weights[i] = 1
		}
	}
	var realWeights []int
	if len(weights) == len(originChildrenOrder) {
		realWeights = weights
	} else if len(weights) > len(originChildrenOrder) {
		realWeights = lo.DropRight(weights, len(weights)-len(originChildrenOrder))
	} else {
		// 填充最小值
		min := lo.Min(weights)
		right := make([]int, len(originChildrenOrder)-len(weights))
		for i := 0; i < len(right); i++ {
			right[i] = min
		}
		realWeights = weights
		realWeights = append(realWeights, right...)
	}
	shuffled, err := wrand.ShuffleWithWeights(realWeights)
	if err != nil {
		r.node.Log().Fatal("order children error", zap.Error(err))
	}
	return shuffled, true
}
