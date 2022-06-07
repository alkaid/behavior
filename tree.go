package behavior

import (
	"github.com/alkaid/behavior/bcore"
	"github.com/alkaid/behavior/logger"
	"go.uber.org/zap"
)

// NodeRegistry 节点注册器
//  用于加载行为树前保存节点实例
type NodeRegistry struct {
	registry map[string]bcore.INode
}

func NewNodeRegistry() *NodeRegistry {
	return &NodeRegistry{registry: make(map[string]bcore.INode)}
}

func (n *NodeRegistry) Register(node bcore.INode) {
	_, ok := n.registry[node.ID()]
	if ok {
		logger.Log.Debug("node already registered", zap.String("node", node.ID()))
		return
	}
	n.registry[node.ID()] = node
}
