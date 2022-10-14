package behavior

import (
	"fmt"
	"reflect"

	"github.com/alkaid/behavior/util"

	"github.com/alkaid/behavior/bcore"

	"github.com/alkaid/behavior/config"

	"github.com/alkaid/behavior/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ClassLoader 节点类加加载器
// 作用 1.缓存节点类2.使用时找到缓存类并实例化出来
type ClassLoader struct {
	registry map[string]reflect.Type // 注册器
}

func NewClassLoader() *ClassLoader {
	return &ClassLoader{make(map[string]reflect.Type)}
}

// New 根据 name 实例化节点
//
//	@receiver l
//	@param name
//	@param cfg
//	@return INode
//	@return error
func (l *ClassLoader) New(name string, cfg *config.NodeCfg) (bcore.INode, error) {
	var err error
	v, ok := l.registry[name]
	if !ok {
		return nil, errors.WithStack(fmt.Errorf("not found struct %s", name))
	}
	c := reflect.New(v).Interface().(bcore.INode)
	err = c.Init(cfg)
	if err != nil {
		return nil, err
	}
	err = c.InitNodeWorker(c.(bcore.INodeWorker))
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Clone 克隆一个节点
//
// @receiver l
// @param node
// @return bcore.INode
// @return error
func (l *ClassLoader) Clone(node bcore.INode) (bcore.INode, error) {
	cfg := *node.RawCfg()
	cfg.ID = util.NanoID()
	return l.New(node.Name(), &cfg)
}

// Contains 检查注册器中是否包含指定节点类
//
//	@receiver l
//	@param name
//	@return bool
func (l *ClassLoader) Contains(name string) bool {
	_, ok := l.registry[name]
	return ok
}

// RegisterWithName 根据名字注册节点类
//
//	@receiver rsm
//	@param name 可以为空,为空时使用节点的struct名称
//	@param c
func (l *ClassLoader) RegisterWithName(name string, node bcore.INode) {
	elem := reflect.TypeOf(node).Elem()
	if name == "" {
		name = elem.Name()
	}
	if _, ok := l.registry[name]; ok {
		logger.Log.Warn("node type has already registered,old node type would be replaced", zap.String("name", name))
	}
	l.registry[name] = elem
}

// Register 注册节点类,注册名为节点的struct名称
//
//	@receiver l
//	@param node
func (l *ClassLoader) Register(node bcore.INode) {
	l.RegisterWithName("", node)
}
