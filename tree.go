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
	DynamicSubtrees map[string]task.IDynamicSubtree // 所有动态子树容器,key为tag
}

// Start 启动行为树,内部会派发到 bcore.IBrain 实例化时指定的线程
//
//	若 Root 状态错误将抛出异常,注意调用方应保证 Start 线程安全以避免 Root 状态读取时为脏数据
//	@receiver t
//	@param brain
func (t *Tree) Start(brain bcore.IBrain) error {
	if !t.Root.IsInactive(brain) {
		return errors.New("tree is not inactive")
	}
	t.Root.Start(brain)
	return nil
}

// Abort 终止树
//
//	若 Root 状态错误将抛出异常,注意调用方应保证 Abort 线程安全以避免 Root 状态读取时为脏数据
//	@receiver t
//	@param brain
func (t *Tree) Abort(brain bcore.IBrain) error {
	if !t.Root.IsActive(brain) {
		return errors.New("tree is not active")
	}
	t.Root.Abort(brain)
	return nil
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
		err = r.LoadFromJson(string(file))
		if err != nil {
			return err
		}
	}
	err := r.MountSubtree()
	if err != nil {
		return err
	}
	return nil
}

func (r *TreeRegistry) LoadFromJsons(cfgJson []string) error {
	for _, j := range cfgJson {
		err := r.LoadFromJson(j)
		if err != nil {
			return err
		}
	}
	err := r.MountSubtree()
	if err != nil {
		return err
	}
	return nil
}

func (r *TreeRegistry) LoadFromJson(cfgJson string) error {
	var cfg config.TreeCfg
	err := json.Unmarshal([]byte(cfgJson), &cfg)
	if err != nil {
		return errors.WithStack(err)
	}
	if cfg.Ver == "" {
		cfg.Ver = fmt.Sprintf("%x", md5.Sum([]byte(cfgJson)))
	}
	return r.Load(&cfg)
}

// Load
//
//	@receiver r
//	@param cfg
//	@return *Tree
//	@return error
//
//nolint:gocyclo
func (r *TreeRegistry) Load(cfg *config.TreeCfg) error {
	err := cfg.Valid()
	if err != nil {
		return err
	}
	// 先从缓存中找 id重复时返回旧树忽略加载 tag重复则加载更新
	tree, ok := r.TreesByID[cfg.Root]
	if ok && tree.Ver == cfg.Ver {
		logger.Log.Warn("tree id already exists,ignore load", zap.String("ver", cfg.Ver), zap.String("tag", cfg.Tag), zap.String("id", cfg.Root))
		return nil
	}
	_, ok = r.TreesByTag[cfg.Tag]
	if ok {
		logger.Log.Debug("tree tag already exists,will load and replace", zap.String("ver", cfg.Ver), zap.String("tag", cfg.Tag), zap.String("id", cfg.Root))
	}
	tree = &Tree{
		Tag:             cfg.Tag,
		Ver:             cfg.Ver,
		DynamicSubtrees: map[string]task.IDynamicSubtree{},
	}
	nodes := map[string]bcore.INode{}
	for _, nodeCfg := range cfg.Nodes {
		err = nodeCfg.Valid()
		if err != nil {
			return err
		}
		node, err := globalClassLoader.New(nodeCfg.Name, nodeCfg)
		if err != nil {
			return err
		}
		nodes[node.ID()] = node
	}
	tree.Root = nodes[cfg.Root].(bcore.IRoot)
	for _, node := range nodes {
		chidlrenIDs := cfg.Nodes[node.ID()].Children
		// 子树容器特殊处理,暂存待依赖树全部加载完再挂载
		if node.Name() == bcore.NodeNameSubtree || node.Name() == bcore.NodeNameDynamicSubtree {
			r.subtrees[node.ID()] = node.(task.ISubtree)
			tree.HaveSubtree = true
			if node.Name() == bcore.NodeNameDynamicSubtree {
				dst, ok := node.(task.IDynamicSubtree)
				if !ok {
					return errors.New("node is not DynamicSubtree")
				}
				tree.DynamicSubtrees[dst.Tag()] = node.(task.IDynamicSubtree)
			}
			continue
		}
		switch node.Category() {
		case bcore.CategoryComposite:
			if len(chidlrenIDs) == 0 {
				return errors.New("composite must have one child at least")
			}
			for _, id := range chidlrenIDs {
				node.(bcore.IComposite).AddChild(nodes[id])
			}
		case bcore.CategoryDecorator:
			if len(chidlrenIDs) != 1 {
				return errors.New("decorator can have only one child")
			}
			node.(bcore.IDecorator).Decorate(nodes[chidlrenIDs[0]])
		case bcore.CategoryTask:
			// do nothing
		default:
			return errors.New("unsupport this category:" + node.Category())
		}
	}
	for _, node := range nodes {
		node.SetRoot(nil, tree.Root)
	}
	r.TreesByID[cfg.Root] = tree
	r.TreesByTag[tree.Tag] = tree
	return nil
}

// MountSubtree 遍历所有未挂载子树的子树容器,挂载子树
//
//	@receiver r
//	@return error
func (r *TreeRegistry) MountSubtree() error {
	for tid, subtree := range r.subtrees {
		var tree *Tree
		var ok = false
		if tag := subtree.GetPropChildTag(); tag != "" {
			tree, ok = r.TreesByTag[tag]
		}
		if id := subtree.GetPropChildID(); !ok && id != "" {
			tree, ok = r.TreesByID[id]
		}
		if !ok && !subtree.CanDynamicDecorate() {
			return errors.New("tree is nil")
		}
		if ok {
			subtree.Decorate(tree.Root)
		}
		// 非动态子树可以删除
		if !subtree.CanDynamicDecorate() {
			delete(r.subtrees, tid)
		}
	}
	return nil
}
