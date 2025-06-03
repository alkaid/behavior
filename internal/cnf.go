package internal

type globalConfig struct {
	ActionSuccessIfNotDelegate bool // 当委托不存在时,action是否返回成功. 常用于debug时,避免每次都要写委托方法.
}

var GlobalConfig = &globalConfig{}
