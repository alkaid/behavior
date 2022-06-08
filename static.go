package behavior

import (
	"github.com/alkaid/behavior/handle"
	"github.com/alkaid/behavior/logger"
	gonanoid "github.com/matoous/go-nanoid"
	"go.uber.org/zap"
)

var globalClassLoader = NewClassLoader()
var handlerPool = handle.NewHandlerPool()

// GlobalClassLoader 全局类加载器
//  @return *ClassLoader
func GlobalClassLoader() *ClassLoader {
	return globalClassLoader
}

// GlobalHandlerPool 全局反射代理缓存池
//  @return *handle.HandlerPool
func GlobalHandlerPool() *handle.HandlerPool {
	return handlerPool
}

// RegisterDelegatorType 注册代理类的反射信息
//  @param name
//  @param target
func RegisterDelegatorType(name string, target any) error {
	return handlerPool.Register(name, target)
}

// NanoID 随机唯一ID like UUID
//  @return string
func NanoID() string {
	id, err := gonanoid.ID(defaultNanoIDLen)
	if err != nil {
		logger.Log.Error("", zap.Error(err))
	}
	return id
}
