package behavior

import (
	"reflect"

	"github.com/alkaid/behavior/bcore"

	"github.com/pkg/errors"

	"github.com/alkaid/behavior/logger"
)

const DefaultDelegateName = "__default"

var _ bcore.IBrain = (*Brain)(nil)

type Brain struct {
	blackboard bcore.IBlackboard
	delegates  map[string]*delegateMeta
}

func NewBrain(blackboard bcore.IBlackboard, delegates map[string]any) bcore.IBrain {
	b := &Brain{
		blackboard: blackboard,
		delegates:  map[string]*delegateMeta{},
	}
	b.SetDelegates(delegates)
	return b
}

type delegateMeta struct {
	delegate     any
	reflectValue reflect.Value
}

func (b *Brain) Blackboard() bcore.IBlackboard {
	return b.blackboard
}

// DefaultDelegate 默认委托对象，当后续的节点配置中若没有提供委托对象,将使用该默认委托
//  一般是业务方提供的AI组件,内部实现了行为树各个 INode 执行时的委托
//  @return any
func (b *Brain) DefaultDelegate() *delegateMeta {
	return b.Delegate(DefaultDelegateName)
}

// Delegate 比 Pawn 更灵活的委托对象，节点在执行委托时应该优先使用该对象执行委托方法,若配置中没有提供委托对象名,则使用 DefaultDelegate
//  @param name 可以为空,为空则调用 DefaultDelegate
//  @return any
func (b *Brain) Delegate(name string) *delegateMeta {
	if name == "" {
		return b.DefaultDelegate()
	}
	return b.delegates[name]
}

func (b *Brain) RegisterDefaultDelegate(delegate any) {
	b.RegisterDelegate(DefaultDelegateName, delegate)
}

func (b *Brain) RegisterDelegate(name string, delegate any) {
	if delegate == nil {
		logger.Log.Fatal("delegate can't be nil")
	}
	b.delegates[name] = &delegateMeta{
		delegate:     delegate,
		reflectValue: reflect.ValueOf(delegate),
	}
}

func (b *Brain) SetDelegates(delegates map[string]any) {
	for name, d := range delegates {
		b.RegisterDelegate(name, d)
	}
}

func (b *Brain) Execute(target string, method string, args ...any) ([]any, error) {
	dmeta := b.Delegate(target)
	if dmeta == nil {
		return nil, errors.New("target is nil")
	}
	return GlobalHandlerPool().ProcessHandlerMessage(target, dmeta.delegate, dmeta.reflectValue, method, args...)
}
