package decorator

import (
	"fmt"

	"github.com/alkaid/behavior/bcore"

	"github.com/alkaid/behavior/util"

	"github.com/google/go-cmp/cmp"

	"go.uber.org/zap"
)

type IBBConditionProperties interface {
	GetOperator() bcore.Operator
	GetKey() string
	GetValue() any
}

// BBConditionProperties 黑板条件节点属性
type BBConditionProperties struct {
	bcore.ObservingProperties
	Operator bcore.Operator `json:"operator"` // 运算符
	Key      string         `json:"key"`      // 黑板键
	Value    any            `json:"value"`
}

func (b *BBConditionProperties) GetOperator() bcore.Operator {
	return b.Operator
}

func (b *BBConditionProperties) GetKey() string {
	return b.Key
}

func (b *BBConditionProperties) GetValue() any {
	return b.Value
}

// BBCondition 黑板条件
//
//	节点将检查给定的 黑板键（Blackboard Key） 上是否设置了值。
type BBCondition struct {
	bcore.ObservingDecorator
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (c *BBCondition) PropertiesClassProvider() any {
	return &BBConditionProperties{}
}

func (c *BBCondition) BBConditionProperties() IBBConditionProperties {
	return c.Properties().(IBBConditionProperties)
}

// StartObserving
//
//	@override bcore.ObservingDecorator .StartObserving
//	@receiver o
//	@param brain
func (c *BBCondition) StartObserving(brain bcore.IBrain) {
	c.ObservingDecorator.StartObserving(brain)
	brain.Blackboard().(bcore.IBlackboardInternal).AddObserver(c.BBConditionProperties().GetKey(), c.getObserver(brain))
}

// StopObserving
//
//	@override bcore.ObservingDecorator .StopObserving
//	@receiver o
//	@param brain
func (c *BBCondition) StopObserving(brain bcore.IBrain) {
	c.ObservingDecorator.StopObserving(brain)
	brain.Blackboard().(bcore.IBlackboardInternal).RemoveObserver(c.BBConditionProperties().GetKey(), c.getObserver(brain))
	c.Memory(brain).DefaultObserver = nil
}

// ConditionMet
//
//	@implement bcore.IObservingWorker .ConditionMet
//	@receiver o
//	@param brain
//	@param args
//	@return bool
//
//nolint:gocyclo
func (c *BBCondition) ConditionMet(brain bcore.IBrain, args ...any) bool {
	// 若委托存在,优先使用委托
	if c.HasDelegatorOrScript() {
		ret := c.Update(brain, bcore.EventTypeOnUpdate, 0)
		return ret == bcore.ResultSucceeded
	}
	v, ok := brain.Blackboard().Get(c.BBConditionProperties().GetKey())
	if !ok {
		return c.BBConditionProperties().GetOperator() == bcore.OperatorIsNotSet
	}
	propValue := c.BBConditionProperties().GetValue()
	// 黑板value转float,配置value转float
	var bbNumber, propNumber float64
	var bbOk, propOk bool
	bbNumber, bbOk = util.Float(v)
	if bbOk {
		propNumber, propOk = util.Float(propValue)
	}
	// 以下类型判断能否成功转成number,不能转则无法比较
	switch c.BBConditionProperties().GetOperator() {
	case bcore.OperatorIsGt, bcore.OperatorIsGte, bcore.OperatorIsLt, bcore.OperatorIsLte:
		if bbOk && propOk {
			break
		}
		c.Log(brain).Error("value cannot compare", zap.Int("operate", int(c.BBConditionProperties().GetOperator())), zap.Any("blackboardValue", v), zap.Any("configValue", propValue))
		return false
	}
	switch c.BBConditionProperties().GetOperator() {
	case bcore.OperatorIsSet:
		return true
	case bcore.OperatorIsEqual:
		if bbOk && propOk {
			return bbNumber == propNumber
		}
		return cmp.Equal(v, propValue)
	case bcore.OperatorIsNotEqual:
		if bbOk && propOk {
			return bbNumber != propNumber
		}
		return !cmp.Equal(v, propValue)
	case bcore.OperatorIsGte:
		return bbNumber >= propNumber
	case bcore.OperatorIsGt:
		return bbNumber > propNumber
	case bcore.OperatorIsLte:
		return bbNumber <= propNumber
	case bcore.OperatorIsLt:
		return bbNumber < propNumber
	}
	c.Log(brain).Error("not support operator", zap.Int("operator", int(c.BBConditionProperties().GetOperator())))
	return false
}

func (c *BBCondition) OnString(brain bcore.IBrain) string {
	return fmt.Sprintf("%s(%d)%s?%v", c.ObservingDecorator.OnString(brain), c.BBConditionProperties().GetOperator(), c.BBConditionProperties().GetKey(), c.BBConditionProperties().GetValue())
}

func (c *BBCondition) getObserver(brain bcore.IBrain) bcore.Observer {
	ob := c.Memory(brain).DefaultObserver
	if ob == nil {
		ob = func(op bcore.OpType, key string, oldValue any, newValue any) {
			c.ObservingDecorator.Evaluate(brain, op, key, oldValue, newValue)
		}
		c.Memory(brain).DefaultObserver = ob
	}
	return ob
}
