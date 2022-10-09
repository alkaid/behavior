package decorator

import (
	"time"

	"go.uber.org/zap"

	"github.com/alkaid/behavior/bcore"
)

type IBBCooldownProperties interface {
	GetKey() string
}

// BBCooldownProperties cd等待装饰器属性
type BBCooldownProperties struct {
	Key string `json:"key"` // 读取冷取时间的黑板KEY
}

func (p *BBCooldownProperties) GetKey() string {
	return p.Key
}

// BBCooldown cd等待装饰器,与 Cooldown 的区别是冷取时间从黑板读取
//
//	每次子节点完成后将等待一段时间才能再次执行
type BBCooldown struct {
	CooldownBase
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (b *BBCooldown) PropertiesClassProvider() any {
	return &BBCooldownProperties{}
}

// CooldownTime
//
//	@implement ICooldownBaseWorker.CooldownTime
//	@receiver b
//	@param brain
//	@return time.Duration
func (b *BBCooldown) CooldownTime(brain bcore.IBrain) time.Duration {
	key := b.Properties().(IBBCooldownProperties).GetKey()
	val, ok := brain.Blackboard().GetDuration(key)
	if !ok {
		b.Log(brain).Error("not found cooldown blackboard key", zap.String("key", key))
		return 0
	}
	return val
}
