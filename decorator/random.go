package decorator

import (
	"math/rand"

	"github.com/alkaid/behavior/bcore"
)

type IRandomProperties interface {
	GetProbability() float64
}

// RandomProperties 随机节点属性
type RandomProperties struct {
	Probability float64 `json:"probability"` // 概率,必须0<=probability<=1
}

func (r *RandomProperties) GetProbability() float64 {
	if r.Probability < 0 {
		return 0
	}
	if r.Probability > 1 {
		return 1
	}
	return r.Probability
}

// Random 随机装饰器
//
//	将根据配置的概率决定是否执行子节点。
type Random struct {
	bcore.Decorator
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (r *Random) PropertiesClassProvider() any {
	return &RandomProperties{}
}

// OnStart
//
//	@override Node.OnStart
//	@receiver n
//	@param brain
func (r *Random) OnStart(brain bcore.IBrain) {
	r.Decorator.OnStart(brain)
	if rand.Float64() <= r.Properties().(IRandomProperties).GetProbability() {
		r.Decorated(brain).Start(brain)
	} else {
		r.Finish(brain, false)
	}
}

// OnAbort
//
//	@override Node.OnAbort
//	@receiver n
//	@param brain
func (r *Random) OnAbort(brain bcore.IBrain) {
	r.Decorator.OnAbort(brain)
	r.Decorated(brain).SetUpstream(brain, r)
	r.Decorated(brain).Abort(brain)
}

// OnChildFinished
//
//	@override bcore.Decorator .OnChildFinished
//	@receiver s
//	@param brain
//	@param child
//	@param succeeded
func (r *Random) OnChildFinished(brain bcore.IBrain, child bcore.INode, succeeded bool) {
	r.Decorator.OnChildFinished(brain, child, succeeded)
	r.Finish(brain, succeeded)
}
