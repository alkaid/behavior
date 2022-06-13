package decorator

import (
	"time"

	"github.com/alkaid/behavior/util"

	"github.com/alkaid/behavior/bcore"
)

type ICooldownProperties interface {
	GetCooldownTime() time.Duration
}

// CooldownProperties cd等待装饰器属性
type CooldownProperties struct {
	CooldownTime util.Duration `json:"cooldownTime"` // 冷却时间
}

func (p *CooldownProperties) GetCooldownTime() time.Duration {
	return p.CooldownTime.Duration
}

// Cooldown cd等待装饰器
//  每次子节点完成后将等待一段时间才能再次执行
type Cooldown struct {
	CooldownBase
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (b *Cooldown) PropertiesClassProvider() any {
	return &CooldownProperties{}
}

// CooldownTime
//  @implement ICooldownBaseWorker.CooldownTime
//  @receiver b
//  @param brain
//  @return time.Duration
func (b *Cooldown) CooldownTime(brain bcore.IBrain) time.Duration {
	return b.Properties().(ICooldownProperties).GetCooldownTime()
}
