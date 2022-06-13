package behavior

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"

	"github.com/alkaid/behavior/task"

	"github.com/alkaid/behavior/bcore"
	"github.com/alkaid/behavior/config"
	"github.com/alkaid/behavior/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Tree struct {
	Root            bcore.IRoot
	Ver             string
	Tag             string
	HaveSubtree     bool
	DynamicSubtrees map[string]task.IDynamicSubtree // 所有动态子树容器
}

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

// TreeRegistry 行为树注册器
type TreeRegistry struct {
	TreesByID  map[string]*Tree         // 所有树,索引为ID
	TreesByTag map[string]*Tree         // 所有树,索引为Tag
	subtrees   map[string]task.ISubtree // 所有子树容器
}

func NewTreeRegistry() *TreeRegistry {
	return &TreeRegistry{
		TreesByID:  map[string]*Tree{},
		TreesByTag: map[string]*Tree{},
		subtrees:   map[string]task.ISubtree{},
	}
}

func (r *TreeRegistry) LoadFromPaths(paths []string) error {
	for _, path := range paths {
		file, err := os.ReadFile(path)
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = r.LoadFromJson(string(file))
		if err != nil {
			return err
		}
	}
	err := r.mountSubtree()
	if err != nil {
		return err
	}
	return nil
}

func (r *TreeRegistry) LoadFromJsons(cfgJson []string) error {
	for _, j := range cfgJson {
		_, err := r.LoadFromJson(j)
		if err != nil {
			return err
		}
	}
	err := r.mountSubtree()
	if err != nil {
		return err
	}
	return nil
}

func (r *TreeRegistry) LoadFromJson(cfgJson string) (*Tree, error) {
	var cfg config.TreeCfg
	err := json.Unmarshal([]byte(cfgJson), &cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = cfg.Valid()
	if err != nil {
		return nil, err
	}
	if cfg.Ver == "" {
		cfg.Ver = fmt.Sprintf("%x", md5.Sum([]byte(cfgJson)))
	}
	// 先从缓存中找
	tree, ok := r.TreesByID[cfg.Root]
	if ok && tree.Ver == cfg.Ver {
		logger.Log.Debug("tree already exists", zap.String("ver", cfg.Ver), zap.String("id", cfg.Root))
		return tree, nil
	}
	tree, err = r.Load(&cfg)
	if err != nil {
		return nil, err
	}
	r.TreesByID[cfg.Root] = tree
	r.TreesByTag[tree.Tag] = tree
	return tree, nil
}

func (r *TreeRegistry) Load(cfg *config.TreeCfg) (*Tree, error) {
	tree := &Tree{
		Tag:             cfg.Tag,
		Ver:             cfg.Ver,
		DynamicSubtrees: map[string]task.IDynamicSubtree{},
	}
	nodes := map[string]bcore.INode{}
	for _, nodeCfg := range cfg.Nodes {
		err := nodeCfg.Valid()
		if err != nil {
			return nil, err
		}
		node, err := globalClassLoader.New(nodeCfg.Name, nodeCfg)
		if err != nil {
			return nil, err
		}
		nodes[node.ID()] = node
	}
	for _, node := range nodes {
		chidlrenIDs := cfg.Nodes[node.ID()].Children
		switch node.Category() {
		case bcore.CategoryDecorator:
			// 子树容器特殊处理,暂存待依赖树全部加载完再挂载
			if node.Name() == bcore.NodeNameSubtree || node.Name() == bcore.NodeNameDynamicSubtree {
				r.subtrees[node.ID()] = node.(task.ISubtree)
				if node.Name() == bcore.NodeNameDynamicSubtree {
					tree.DynamicSubtrees[node.ID()] = node.(task.IDynamicSubtree)
				}
				tree.HaveSubtree = true
				continue
			}
			if len(chidlrenIDs) != 1 {
				return nil, errors.New("decorator can have only one child")
			}
			node.(bcore.IDecorator).Decorate(nodes[chidlrenIDs[0]])
		case bcore.CategoryComposite:
			if len(chidlrenIDs) == 0 {
				return nil, errors.New("composite must have one child at least")
			}
			for _, id := range chidlrenIDs {
				node.(bcore.IComposite).AddChild(nodes[id])
			}
		case bcore.CategoryTask:
		default:
			return nil, errors.New("unsupport this category:" + node.Category())
		}
	}
	tree.Root = nodes[cfg.Root].(bcore.IRoot)
	return tree, nil
}

func (r *TreeRegistry) mountSubtree() error {
	for tid, subtree := range r.subtrees {
		id := subtree.GetPropChildID()
		root, ok := r.TreesByID[id]
		if !ok {
			return errors.New("root is nil")
		}
		subtree.Decorate(root.Root)
		delete(r.subtrees, tid)
	}
	return nil
}
