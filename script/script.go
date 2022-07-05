package script

import (
	"fmt"

	"github.com/alkaid/behavior/logger"
	"github.com/bilibili/gengine/engine"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var pool *engine.GenginePool    // 引擎池单例
var codes = map[string]string{} // 代码,索引为nodeID
var apiLib = map[string]any{}   // 公共api库

// InitPool 初始化规则引擎池
//  @param poolMinLen 最小容量
//  @param poolMaxLen 最大容量
//  @param api 注入的API
//  @return error
func InitPool(poolMinLen, poolMaxLen int, api map[string]any) error {
	if pool != nil {
		logger.Log.Warn("init goroutine pool duplicate")
		return nil
	}
	RegisterApi(api)
	var err error
	// 实例化pool时脚本不能为空
	codeWrap := fmt.Sprintf(`
rule "%s"
begin
	%s
end
	`, "donothing", "")
	pool, err = engine.NewGenginePool(int64(poolMinLen), int64(poolMaxLen), engine.SortModel, codeWrap, apiLib)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// RegisterApi 向引擎池注册公共API
//  @param api
func RegisterApi(api map[string]any) {
	for name, f := range api {
		_, ok := apiLib[name]
		if ok {
			logger.Log.Warn("api exits,ignore inject it", zap.String("name", name))
			continue
		}
		apiLib[name] = f
	}
}

// RegisterCode 注册代码
//  @param name 规则名称 一般是nodeID
//  @param code 代码
func RegisterCode(name string, code string) {
	codeWrap := fmt.Sprintf(`
rule "%s"
begin
	%s
end
	`, name, code)
	codes[name] = codeWrap
}

// UpdateCode 全量更新池子里的代码
//  @return error
func UpdateCode() error {
	all := ""
	for _, code := range codes {
		all += code
	}
	err := pool.UpdatePooledRules(all)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
func ExistsCode(name string) bool {
	return pool.IsExist([]string{name})[0]
}

// RunCode 执行代码
//  @param name
//  @param env
func RunCode(name string, env map[string]any) (out any, err error) {
	err, rets := pool.ExecuteSelectedRules(env, []string{name})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return rets[name], nil
}
