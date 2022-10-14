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
	Root              bcore.IRoot
	Ver               string
	Tag               string
	StaticSubtrees    map[string]task.ISubtree        // 所有静态子树容器,索引为id
	DynamicSubtrees   map[string]task.IDynamicSubtree // 所有动态子树容器,key为tag
	AllSubtreeMounted bool                            // 是否所有子树已经全部挂载完(不包括childTag为空的)
}

// Clone 拷贝整个树
//
// @receiver t
// @return *Tree
// @return error
func (t *Tree) Clone() (*Tree, error) {
	tree := &Tree{
		Tag:             t.Tag,
		Ver:             t.Ver,
		StaticSubtrees:  map[string]task.ISubtree{},
		DynamicSubtrees: map[string]task.IDynamicSubtree{},
	}
	_, err := t.backtrackingClone(t.Root, tree)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (t *Tree) backtrackingClone(originNode bcore.INode, newTree *Tree) (bcore.INode, error) {
	newNode, err := globalClassLoader.Clone(originNode)
	if err != nil {
		return nil, err
	}
	// 子树容器类型 终止
	switch newNode.Name() {
	case bcore.NodeNameSubtree:
		dst, _ := newNode.(task.ISubtree)
		newTree.StaticSubtrees[dst.ID()] = dst
		return newNode, nil
	case bcore.NodeNameDynamicSubtree:
		dst, _ := newNode.(task.IDynamicSubtree)
		newTree.DynamicSubtrees[dst.Tag()] = dst
		return newNode, nil
	}
	// 非container类型 终止
	if newNode.Category() != bcore.CategoryDecorator && newNode.Category() != bcore.CategoryComposite {
		return newNode, nil
	}
	// root类型 赋值
	if newTree.Root == nil && originNode.Name() == bcore.NodeNameRoot {
		newTree.Root = newNode.(bcore.IRoot)
	}
	// 处理子节点
	var children []bcore.INode
	switch v := originNode.(type) {
	case bcore.IComposite:
		children = v.Children()
	case bcore.IDecorator:
		children = append(children, v.Decorated(nil))
	default:
		return nil, errors.New(fmt.Sprintf("unSupport category:%s", originNode.Category()))
	}
	for _, child := range children {
		newChild, err := t.backtrackingClone(child, newTree)
		if err != nil {
			return nil, err
		}
		newChild.SetRoot(nil, newTree.Root)
		newChild.SetParent(newNode.(bcore.IContainer))
	}
	return newNode, nil
}

// TreeRegistry 行为树注册器
type TreeRegistry struct {
	TreesByID  map[string]*Tree   // 所有树,索引为 IRoot.ID
	TreesByTag map[string][]*Tree // 所有树,索引为 IRoot.Tag
}

func NewTreeRegistry() *TreeRegistry {
	return &TreeRegistry{
		TreesByID:  map[string]*Tree{},
		TreesByTag: map[string][]*Tree{},
	}
}

func (r *TreeRegistry) LoadFromPaths(paths []string) error {
	for _, path := range paths {
		file, err := os.ReadFile(path)
		if err != nil {
			return errors.WithStack(err)
		}
		err = r.LoadFromJson(file)
		if err != nil {
			return err
		}
	}
	err := r.MountAll()
	if err != nil {
		return err
	}
	return nil
}

func (r *TreeRegistry) LoadFromJsons(cfgJson [][]byte) error {
	for _, j := range cfgJson {
		err := r.LoadFromJson(j)
		if err != nil {
			return err
		}
	}
	err := r.MountAll()
	if err != nil {
		return err
	}
	return nil
}

func (r *TreeRegistry) LoadFromJson(cfgJson []byte) error {
	var cfg config.TreeCfg
	err := json.Unmarshal(cfgJson, &cfg)
	if err != nil {
		return errors.WithStack(err)
	}
	if cfg.Ver == "" {
		cfg.Ver = fmt.Sprintf("%x", md5.Sum(cfgJson))
	}
	return r.Load(&cfg)
}

// Remove 根据tag移除树,移除前请务必:1.停止使用该树运行的AI 2.同时移除关联树(该树的静态子树和动态子树)
//
// @receiver r
// @param tag
func (r *TreeRegistry) Remove(tag string) {
	for _, tree := range r.TreesByTag[tag] {
		delete(r.TreesByID, tree.Root.ID())
	}
	delete(r.TreesByTag, tag)
}

// Load 加载树,加载前请务必:1.停止使用该树运行的AI 2.移除该树旧版及其关联树(该树的静态子树和动态子树)
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
	// 先从缓存中找 tag+ver重复时返回旧树忽略加载
	trees, ok := r.TreesByTag[cfg.Tag]
	var tree *Tree
	if ok {
		tree = trees[0]
	}
	if tree != nil && tree.Ver == cfg.Ver {
		logger.Log.Warn("tree id already exists,ignore load", zap.String("ver", cfg.Ver), zap.String("tag", cfg.Tag), zap.String("id", cfg.Root))
		return nil
	}
	r.Remove(cfg.Tag)
	tree = &Tree{
		Tag:             cfg.Tag,
		Ver:             cfg.Ver,
		StaticSubtrees:  map[string]task.ISubtree{},
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
		switch node.Name() {
		case bcore.NodeNameSubtree:
			dst, _ := node.(task.ISubtree)
			tree.StaticSubtrees[dst.ID()] = dst
			continue
		case bcore.NodeNameDynamicSubtree:
			dst, _ := node.(task.IDynamicSubtree)
			tree.DynamicSubtrees[dst.Tag()] = dst
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
	r.TreesByTag[tree.Tag] = append(r.TreesByTag[tree.Tag], tree)
	return nil
}

// MountAll 遍历所有未挂载子树的子树容器,挂载子树
//
//	@receiver r
//	@return error
func (r *TreeRegistry) MountAll() error {
	for _, tree := range r.TreesByID {
		err := r.mountAllSubtree(tree)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetNotParentTreeWithoutClone 获取一个还未分配静态父节点的树，多用于获取该tag的主树.
//
// @receiver r
// @param tag
// @return *Tree
func (r *TreeRegistry) GetNotParentTreeWithoutClone(tag string) *Tree {
	for _, tree := range r.TreesByTag[tag] {
		parent := tree.Root.Parent(nil)
		if parent == nil {
			return tree
		}
	}
	return nil
}

// getNotParentTree 获取一个还未分配的静态父节点的树，若没有未分配的则clone一个
//
// @receiver r
// @param tag 树的tag
// @return utree 无父节点的树
// @return cloned utree 是否是clone出来的
// @return err
func (r *TreeRegistry) getNotParentTree(tag string) (utree *Tree, cloned bool, err error) {
	if len(r.TreesByTag[tag]) == 0 {
		return nil, false, nil
	}
	for _, tree := range r.TreesByTag[tag] {
		parent := tree.Root.Parent(nil)
		if parent == nil {
			return tree, false, nil
		}
	}
	child, err := r.TreesByTag[tag][0].Clone()
	if err != nil {
		return nil, false, err
	}
	return child, true, nil
}

// getNotDynamicParentTree 获取一个还未分配的动态父节点的树，若没有未分配的则clone一个
//
// @receiver r
// @param tag
// @param brain
// @return utree
// @return cloned
// @return err
func (r *TreeRegistry) getNotDynamicParentTree(tag string, brain bcore.IBrain) (utree *Tree, cloned bool, err error) {
	if len(r.TreesByTag[tag]) == 0 {
		return nil, false, nil
	}
	for _, tree := range r.TreesByTag[tag] {
		parent := brain.Blackboard().(bcore.IBlackboardInternal).NodeMemory(tree.Root.ID()).MountParent
		if parent == nil {
			return tree, false, nil
		}
	}
	child, err := r.TreesByTag[tag][0].Clone()
	if err != nil {
		return nil, false, err
	}
	// 注意clone出来的树要处理未挂载的节点
	err = r.mountAllSubtree(child)
	if err != nil {
		return nil, false, err
	}
	return child, true, nil
}

func (r *TreeRegistry) mountAllSubtree(tree *Tree) error {
	if tree.AllSubtreeMounted {
		return nil
	}
	allMounted := true
	each := func(container task.ISubtree) error {
		if container.Decorated(nil) != nil {
			return nil
		}
		tag := container.GetPropChildTag()
		// 无子树tag配置,不挂载
		if tag == "" {
			return nil
		}
		child, cloned, err := r.getNotParentTree(tag)
		if err != nil {
			allMounted = false
			return err
		}
		// 找不到子树 可能还没加载
		if child == nil {
			allMounted = false
			return nil
		}
		// 找到子树,装饰
		container.Decorate(child.Root)
		// 如果是克隆出来的,存储并递归处理子树的挂载
		if !cloned {
			return nil
		}
		r.TreesByTag[tag] = append(r.TreesByTag[tag], child)
		r.TreesByID[child.Root.ID()] = child
		return r.mountAllSubtree(child)
	}
	for _, container := range tree.StaticSubtrees {
		err := each(container)
		if err != nil {
			return err
		}
	}
	for _, container := range tree.DynamicSubtrees {
		err := each(container)
		if err != nil {
			return err
		}
	}
	tree.AllSubtreeMounted = allMounted
	return nil
}
