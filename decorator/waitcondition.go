package decorator

import (
	"github.com/alkaid/behavior/bcore"
)

type IWaitConditionProperties interface {
}

// WaitConditionProperties 条件等待装饰器属性
type WaitConditionProperties struct {
	ConditionProperties
}

// WaitCondition 条件等待装饰器
//  将阻塞等待条件成功才开始执行分支
type WaitCondition struct {
	bcore.Decorator
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (o *WaitCondition) PropertiesClassProvider() any {
	return &WaitConditionProperties{}
}

// TODO 未完,同样的还有 黑板条件等待 黑板条目等待,考虑抽象一个等待监听器基类
