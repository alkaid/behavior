package behavior

import (
	"reflect"
	"time"

	"github.com/alkaid/behavior/bcore"

	"github.com/pkg/errors"

	"github.com/alkaid/behavior/logger"
)

var _ bcore.IBrain = (*Brain)(nil)

// ExecutorFun 委托方法签名例
type ExecutorFun = func(eventType bcore.EventType, delta time.Duration) bcore.Result

type Brain struct {
	blackboard bcore.IBlackboard
	delegates  map[string]*bcore.DelegateMeta
}

func NewBrain(blackboard bcore.IBlackboard, delegates map[string]any) bcore.IBrain {
	b := &Brain{
		blackboard: blackboard,
		delegates:  map[string]*bcore.DelegateMeta{},
	}
	b.SetDelegates(delegates)
	return b
}

func (b *Brain) Blackboard() bcore.IBlackboard {
	return b.blackboard
}

// RegisterDelegate 注册委托对象
//  @receiver b
//  @param name
//  @param delegate
func (b *Brain) RegisterDelegate(name string, delegate any) {
	if delegate == nil {
		logger.Log.Fatal("delegate can't be nil")
	}
	b.delegates[name] = &bcore.DelegateMeta{
		Delegate:     delegate,
		ReflectValue: reflect.ValueOf(delegate),
	}
}

// SetDelegates 注册委托对象
//  @receiver b
//  @param delegates
func (b *Brain) SetDelegates(delegates map[string]any) {
	for name, d := range delegates {
		b.RegisterDelegate(name, d)
	}
}

func (b *Brain) OnExecute(target string, method string, args ...any) ([]any, error) {
	dmeta := b.delegates[target]
	if dmeta == nil {
		return nil, errors.New("target is nil")
	}
	return GlobalHandlerPool().ProcessHandlerMessage(target, dmeta.Delegate, dmeta.ReflectValue, method, args...)
}
