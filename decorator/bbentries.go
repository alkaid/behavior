package decorator

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/alkaid/behavior/bcore"
)

// BBEntriesOp 比较黑板条目操作符
type BBEntriesOp int

const (
	BBEntriesOpEqual    BBEntriesOp = iota // 相等
	BBEntriesOpNotEqual                    // 不相等
)

type IBBEntriesProperties interface {
	GetOperator() BBEntriesOp
	GetKeys() []string
	GetQuery() string
}

// BBEntriesProperties 黑板条件节点属性
type BBEntriesProperties struct {
	bcore.ObservingProperties
	Operator BBEntriesOp `json:"operator"` // 运算符
	Keys     []string    `json:"keys"`     // 黑板键
	Query    string      `json:"query"`    // 自定义查询语句
}

func (b *BBEntriesProperties) GetOperator() BBEntriesOp {
	return b.Operator
}

func (b *BBEntriesProperties) GetKeys() []string {
	return b.Keys
}
func (b *BBEntriesProperties) GetQuery() string {
	return b.Query
}

// BBEntries 比较黑板条目（Compare BBEntries）
//
//	节点将比较多个 黑板键 的值，并根据结果（等于或不等）阻止或允许节点的执行。
type BBEntries struct {
	bcore.ObservingDecorator
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (e *BBEntries) PropertiesClassProvider() any {
	return &BBEntriesProperties{}
}

func (e *BBEntries) BBEntriesProperties() IBBEntriesProperties {
	return e.Properties().(IBBEntriesProperties)
}

// StartObserving
//
//	@override bcore.ObservingDecorator .StartObserving
//	@receiver o
//	@param brain
func (e *BBEntries) StartObserving(brain bcore.IBrain) {
	e.ObservingDecorator.StartObserving(brain)
	for _, key := range e.BBEntriesProperties().GetKeys() {
		brain.Blackboard().(bcore.IBlackboardInternal).AddObserver(key, e.getObserver(brain))
	}
}

// StopObserving
//
//	@override bcore.ObservingDecorator .StopObserving
//	@receiver o
//	@param brain
func (e *BBEntries) StopObserving(brain bcore.IBrain) {
	e.ObservingDecorator.StopObserving(brain)
	for _, key := range e.BBEntriesProperties().GetKeys() {
		brain.Blackboard().(bcore.IBlackboardInternal).RemoveObserver(key, e.getObserver(brain))
	}
	e.Memory(brain).DefaultObserver = nil
}

// ConditionMet
//
//	@implement bcore.IObservingWorker .ConditionMet
//	@receiver o
//	@param brain
//	@param args
//	@return bool
func (e *BBEntries) ConditionMet(brain bcore.IBrain, args ...any) bool {
	// 若委托存在,优先使用委托
	if e.HasDelegatorOrScript() {
		ret := e.Update(brain, bcore.EventTypeOnUpdate, 0)
		return ret == bcore.ResultSucceeded
	}
	var strValues []string
	allEqual := true
	for i, key := range e.BBEntriesProperties().GetKeys() {
		v, _ := brain.Blackboard().Get(key)
		str := ""
		if v != nil {
			str = fmt.Sprintf("%v", v)
		}
		if str != strValues[i-1] {
			allEqual = false
		}
		strValues = append(strValues, str)
	}
	switch e.BBEntriesProperties().GetOperator() {
	case BBEntriesOpEqual:
		return allEqual
	case BBEntriesOpNotEqual:
		return !allEqual
	}
	e.Log(brain).Error("not support operator", zap.Int("operator", int(e.BBEntriesProperties().GetOperator())))
	return false
}

func (e *BBEntries) OnString(brain bcore.IBrain) string {
	return fmt.Sprintf("%s(%d)%v", e.ObservingDecorator.OnString(brain), e.BBEntriesProperties().GetOperator(), e.BBEntriesProperties().GetKeys())
}

func (e *BBEntries) getObserver(brain bcore.IBrain) bcore.Observer {
	ob := e.Memory(brain).DefaultObserver
	if ob == nil {
		ob = e.ObservingDecorator.NewBlackboardObserver(brain)
		e.Memory(brain).DefaultObserver = ob
	}
	return ob
}
