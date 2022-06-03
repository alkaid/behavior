package config

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// NodeCfg 节点配置,也是默认解析器所能解析的格式
type NodeCfg struct {
	Id         string          `json:"id"`         // 唯一ID
	Name       string          `json:"name"`       // 节点名
	Category   string          `json:"category"`   // 类型
	Title      string          `json:"title"`      // 描述
	Children   []NodeCfg       `json:"children"`   // 孩子节点
	Properties json.RawMessage `json:"properties"` // 自定义属性,须由子类自行解析
}

func (n *NodeCfg) Valid() error {
	if n.Id == "" {
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
