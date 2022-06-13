package behavior

import (
	"github.com/alkaid/behavior/handle"
)

var globalClassLoader = NewClassLoader()
var handlerPool = handle.NewHandlerPool()
var treeRegistry = NewTreeRegistry()

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

func GlobalTreeRegistry() *TreeRegistry {
	return treeRegistry
}

// RegisterDelegatorType 注册代理类的反射信息
//  @param name
//  @param target
func RegisterDelegatorType(name string, target any) error {
	return handlerPool.Register(name, target)
}
