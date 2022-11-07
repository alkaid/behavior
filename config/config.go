package config

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// NodeCfg 节点配置,也是默认解析器所能解析的格式
type NodeCfg struct {
	ID         string          `json:"id"`         // 唯一ID
	Name       string          `json:"name"`       // 节点名
	Category   string          `json:"category"`   // 类型
	Title      string          `json:"title"`      // 描述
	Children   []string        `json:"children"`   // 孩子节点
	Properties json.RawMessage `json:"properties"` // 自定义属性,须由子类自行解析
	Delegator  DelegatorCfg    `json:"delegator"`  // 委托配置
}

func (n *NodeCfg) Valid() error {
	if n.ID == "" {
		return errors.New("id can't nil")
	}
	if n.Name == "" {
		return errors.New("Name can't nil")
	}
	if n.Category == "" {
		return errors.New("Category can't nil")
	}
	return nil
}

// DelegatorCfg 委托配置
type DelegatorCfg struct {
	Target string `json:"target"` // 委托对象,可以为空,为空则使用root的委托对象
	Method string `json:"method"` // 委托方法
	Script string `json:"script"` // 委托脚本
}

// TreeCfg 树配置
type TreeCfg struct {
	Nodes       map[string]*NodeCfg `json:"nodes"`       // 所有节点
	Ver         string              `json:"ver"`         // 数据版本
	Root        string              `json:"root"`        // 根节点nanoID
	Tag         string              `json:"tag"`         // 行为树标志,必须能简要描述业务逻辑且全局唯一
	Description string              `json:"description"` // 业务逻辑详细描述
}

func (c *TreeCfg) Valid() error {
	if c.Root == "" {
		return errors.New(fmt.Sprintf("root cannot be nil,root=%s,tag=%s,desc=%s", c.Root, c.Tag, c.Description))
	}
	if c.Tag == "" {
		return errors.New(fmt.Sprintf("tag cannot be nil,root=%s,tag=%s,desc=%s", c.Root, c.Tag, c.Description))
	}
	if len(c.Nodes) == 0 {
		return errors.New("nodes length is zero")
	}
	return nil
}
