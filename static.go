package behavior

import (
	"github.com/alkaid/behavior/logger"
	gonanoid "github.com/matoous/go-nanoid"
	"go.uber.org/zap"
)

var globalClassLoader = NewClassLoader()

// GlobalClassLoader 全局类加载器
//  @return *ClassLoader
func GlobalClassLoader() *ClassLoader {
	return globalClassLoader
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
