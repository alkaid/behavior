package behavior

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/alkaid/behavior/logger"
)

const DefaultDelegateName = "__default"

// IBrain 大脑=行为树+记忆+委托对象集合. 也可以理解为上下文
type IBrain interface {
	// Tree() (*BehaviorTree, error)

	// Blackboard 获取黑板
	//  @return IBlackboard
	Blackboard() IBlackboard
	// Execute 执行委托方法
	//  @param target 委托对象名
	//  @param method 委托方法名
	//  @param args 参数列表,注意若多个请加上...
	//  @return []any
	//  @return error
	Execute(target string, method string, args ...any) ([]any, error)
}

var _ IBrain = (*Brain)(nil)

type Brain struct {
	blackboard IBlackboard
	delegates  map[string]*delegateMeta
}

func NewBrain(blackboard IBlackboard, delegates map[string]any) IBrain {
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

func (b *Brain) Blackboard() IBlackboard {
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
