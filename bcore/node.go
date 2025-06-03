package bcore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alkaid/behavior/internal"
	"time"

	"github.com/alkaid/behavior/util"

	"github.com/alkaid/behavior/script"

	"github.com/pkg/errors"

	"github.com/samber/lo"

	"github.com/alkaid/behavior/config"
	"github.com/alkaid/behavior/logger"
	"github.com/disiqueira/gotree"
	"go.uber.org/zap"
)

type NodeState int // 节点状态

// INode 节点
//
//	继承自 Node 的自定义节点须实现 INodeWorker
//	一般来说,继承自 Node 的子类不应覆写该接口
type INode interface {
	// InitNodeWorker 设置回调接口,应该委托给继承树中的叶子子类实现
	//  @param worker
	InitNodeWorker(worker INodeWorker) error
	// NodeWorker 获取回调委托(子类对象)
	//  @return INodeWorker
	NodeWorker() INodeWorker
	// NodeWorkerAsNode 获取子类对象
	//
	// @return INode
	NodeWorkerAsNode() INode
	// Init 根据配置加载节点实例
	//  @param cfg
	Init(cfg *config.NodeCfg) error
	// Memory 节点数据
	//  @receiver n
	//  @param brain
	//  @return *NodeMemory
	Memory(brain IBrain) *NodeMemory
	ID() string
	Title() string
	Category() string
	Properties() any
	Name() string
	RawCfg() *config.NodeCfg
	Delegator() config.DelegatorCfg
	// HasDelegatorOrScript 是否存在委托方法或脚本
	//  @return bool
	HasDelegatorOrScript() bool
	State(brain IBrain) NodeState
	IsActive(brain IBrain) bool
	IsInactive(brain IBrain) bool
	IsAborting(brain IBrain) bool
	// Root 获取root
	//  注意由于行为数是单例,如果当前节点是root(子树root),只应该从黑板设置和获取root
	//  @return IRoot
	Root(brain IBrain) IRoot
	// SetRoot 设置根节点
	//  注意由于行为数是单例,这里只能设置当前所在子树的root.
	//  @param root
	SetRoot(brain IBrain, root IRoot)
	SetUpstream(brain IBrain, upstream INode)
	// Start 开始执行 非线程安全,上层须自行封装线程安全的方法
	//  @param brain
	Start(brain IBrain)
	// Abort 中断(终止)执行 非线程安全,上层须自行封装线程安全的方法
	//  一般来说只应该由 IDecorator 调用
	//  一般来说,中断链路为:
	//   节点A Abort-> A INodeWorker.OnAbort-> 子节点B Abort (其他所有子结点同理)-> B INodeWorker.OnAbort-> B Finish-> A IContainer.ChildFinished-> <所有子节点都完成了?true:> A Finish-> 继续按上诉链路回溯上一层
	//  @param brain
	Abort(brain IBrain)
	// Finish 完成执行
	//  这里是真正的停止,会将执行结果 Result 回溯给父节点
	//  注意: Finish 应该是你的最后一个调用,在调用 Finish 后不要再修改任何状态或调用任何东西,否则将产生不可预知的后果
	//  一般来说,停止链路为:
	//  节点B Finish-> 父节点A IContainer.ChildFinished-> <所有子节点都完成了?true:> A Finish-> 继续按上诉链路回溯上一层
	//  @param brain
	//  @param succeeded
	Finish(brain IBrain, succeeded bool)
	// CompositeAncestorFinished 最近的 IComposite 类型的祖先节点关闭时调用
	//  由于直接父节点可能是装饰节点,所以这里特指“最近的 IComposite 类型的祖先节点”
	//  @param brain
	//  @param composite
	CompositeAncestorFinished(brain IBrain, composite IComposite)
	String(brain IBrain) string
	Path(brain IBrain) string
	// Parent 获取父节点
	//  @return IContainer
	Parent(brain IBrain) IContainer
	// SetParent 设置父节点/或者静态挂载.注意不要设置成父节点的父类对象了(若拿到的是父类对象,应该调用 NodeWorker 来获取具体的子类对象).
	//  @param parent
	SetParent(parent IContainer)
	// DynamicMount 将自己作为子树动态挂载到 parent
	//  @param brain
	//  @param parent
	DynamicMount(brain IBrain, parent IContainer)
	Log(brain IBrain, ops ...LogOption) *zap.Logger
	Dump(brain IBrain) string
}

// INodeWorker Node 中会回调的方法,应该委托给继承树中的叶子子类实现
//
//	继承 Node 须覆写该接口
type INodeWorker interface {
	// OnStart Node.Start 的回调
	//  @param brain
	OnStart(brain IBrain)
	// OnAbort Node.Abort 的回调
	//  @param brain
	OnAbort(brain IBrain)
	// PropertiesClassProvider 提供用于解析 Node.properties 的类的零值. Node.Init 的回调
	//  @param properties
	PropertiesClassProvider() any
	// OnCompositeAncestorFinished Node.CompositeAncestorFinished 的回调
	//  目前仅用来给装饰器移除监察函数
	//  @param brain
	//  @param composite
	OnCompositeAncestorFinished(brain IBrain, composite IComposite)
	// CanMountTo 是否可以挂载到目标容器节点上,会被 Node.Parent 和 Node.SetParent 回调
	//  @return bool
	CanMountTo() bool
	// OnString Node.String 的回调
	//  @param brain
	//  @return string
	OnString(brain IBrain) string
}

// BaseProperties 属性基类
type BaseProperties struct{}

var _ INode = (*Node)(nil)
var _ INodeWorker = (*Node)(nil)

// Node 节点基类
//
//	@implement INode
//	@implement INodeWorker
type Node struct {
	INodeWorker
	parent     IContainer          // 父节点
	root       IRoot               // 当前子树的根节点
	name       string              // 节点名
	id         string              // 唯一ID
	title      string              // 描述
	category   string              // 类型
	properties any                 // 自定义属性
	delegator  config.DelegatorCfg // 委托
	cfg        *config.NodeCfg     // 原始配置
}

// Init
//
//	@implement INode.Init
//	@receiver n
//	@param cfg
func (n *Node) Init(cfg *config.NodeCfg) error {
	err := cfg.Valid()
	if err != nil {
		return err
	}
	n.id = cfg.ID
	n.name = cfg.Name
	n.title = cfg.Title
	n.category = cfg.Category
	n.properties = cfg.Properties
	if cfg.Delegator.Method != "" {
		n.delegator = config.DelegatorCfg{
			Target: cfg.Delegator.Target,
			Method: cfg.Delegator.Method,
			Script: cfg.Delegator.Script,
		}
		// 预编译脚本
		if n.delegator.Script != "" {
			script.RegisterCode(n.id, n.delegator.Script)
		}
	}
	n.cfg = cfg
	return nil
}

// InitNodeWorker
//
//	@implement INode.InitNodeWorker
//	@receiver n
//	@param worker
func (n *Node) InitNodeWorker(worker INodeWorker) error {
	n.INodeWorker = worker
	properties := worker.PropertiesClassProvider()
	if properties == nil {
		n.properties = nil
	}
	err := json.Unmarshal(n.properties.(json.RawMessage), properties)
	if err != nil {
		return errors.WithStack(err)
	}
	n.properties = properties
	return nil
}

// CopyTo 拷贝所有成员到 target @implement INode.CopyTo
//
// @receiver n
// @param target 新的零值 Node
func (n *Node) CopyTo(target *Node) {
	target.category = n.category
	target.title = n.title
	target.delegator = n.delegator
	target.name = n.name
	target.properties = n.properties
	target.id = util.NanoID()
}

// NodeWorker
//
//	@implement INode.NodeWorker
//	@receiver n
//	@return INodeWorker
func (n *Node) NodeWorker() INodeWorker {
	return n.INodeWorker
}

// NodeWorkerAsNode
//
//	@implement INode.NodeWorkerAsNode
//
// @receiver n
// @return INode
func (n *Node) NodeWorkerAsNode() INode {
	return n.INodeWorker.(INode)
}

// Memory 节点数据
//
//	@implement INode.Memory
//	@receiver n
//	@param brain
//	@return *NodeMemory
func (n *Node) Memory(brain IBrain) *NodeMemory {
	return brain.Blackboard().(IBlackboardInternal).NodeMemory(n.id)
}

func (n *Node) RawCfg() *config.NodeCfg {
	return n.cfg
}
func (n *Node) Delegator() config.DelegatorCfg {
	return n.delegator
}

func (n *Node) HasDelegatorOrScript() bool {
	return n.delegator.Method != "" || n.delegator.Script != ""
}

func (n *Node) Name() string {
	return n.name
}

func (n *Node) ID() string {
	return n.id
}

func (n *Node) Title() string {
	return n.title
}

func (n *Node) Category() string {
	return n.category
}

func (n *Node) Properties() any {
	return n.properties
}
func (n *Node) State(brain IBrain) NodeState {
	return brain.Blackboard().(IBlackboardInternal).NodeMemory(n.id).State
}
func (n *Node) IsActive(brain IBrain) bool {
	return brain.Blackboard().(IBlackboardInternal).NodeMemory(n.id).State == NodeStateActive
}
func (n *Node) IsInactive(brain IBrain) bool {
	return brain.Blackboard().(IBlackboardInternal).NodeMemory(n.id).State == NodeStateInactive
}

func (n *Node) IsAborting(brain IBrain) bool {
	return brain.Blackboard().(IBlackboardInternal).NodeMemory(n.id).State == NodeStateAborting
}

// Parent
//
//	@implement INode.Parent()
//	@receiver n
//	@return IContainer
func (n *Node) Parent(brain IBrain) IContainer {
	if brain == nil {
		return n.parent
	}
	// 注意行为树是单例,故动态挂载的节点从黑板获取父节点,目前只有root可以
	if n.INodeWorker.CanMountTo() {
		mountParent := brain.Blackboard().(IBlackboardInternal).NodeMemory(n.id).MountParent
		// 不为空为动态挂载,否则为静态挂载
		if mountParent != nil {
			return mountParent
		}
	}
	return n.parent
}

// SetParent
//
//	@implement INode.SetParent
//	@receiver n
//	@param parent
func (n *Node) SetParent(parent IContainer) {
	n.parent = parent
}

// DynamicMount
//
//	@implement INode.DynamicMount
//	@receiver n
//	@param brain
//	@param parent
func (n *Node) DynamicMount(brain IBrain, parent IContainer) {
	if !n.INodeWorker.CanMountTo() {
		n.Log(brain).Fatal("cannot mount for this node")
		return
	}
	brain.Blackboard().(IBlackboardInternal).NodeMemory(n.id).MountParent = parent
}

// Root
//
//	@implement INode.Root()
//	@receiver n
//	@return IRoot
func (n *Node) Root(brain IBrain) IRoot {
	// TODO 黑板获取动态root
	return n.root
}

// SetRoot
//
//	@implement INode.SetRoot
//	@receiver n
//	@param root
func (n *Node) SetRoot(brain IBrain, root IRoot) {
	n.root = root
}

type BrainContextKey string

var upstreamKey BrainContextKey = "_upstream"

func (n *Node) SetUpstream(brain IBrain, upstream INode) {
	ibrain := brain.(IBrainInternal)
	ibrain.SetContext(context.WithValue(ibrain.Context(), upstreamKey, upstream))
}

// Start
//
//	@implement INode.Start
//	@receiver n
//	@param brain
func (n *Node) Start(brain IBrain) {
	// TODO debug info
	nodeData := brain.Blackboard().(IBlackboardInternal).NodeMemory(n.id)
	if nodeData.State != NodeStateInactive {
		n.Log(brain).Error("can only start inactive nodes")
		return
	}
	nodeData.State = NodeStateActive
	n.INodeWorker.OnStart(brain)
}

// Abort
//
//	@implement INode.Abort
//	@receiver n
//	@param brain
func (n *Node) Abort(brain IBrain) {
	nodeData := brain.Blackboard().(IBlackboardInternal).NodeMemory(n.id)
	if nodeData.State != NodeStateActive {
		n.Log(brain).Error("can only abort active nodes")
		return
	}
	nodeData.State = NodeStateAborting
	n.INodeWorker.OnAbort(brain)
}

// Finish
//
//	@implement INode.Finish
//	@receiver n
//	@param brain
//	@param succeeded
func (n *Node) Finish(brain IBrain, succeeded bool) {
	nodeData := brain.Blackboard().(IBlackboardInternal).NodeMemory(n.id)
	if nodeData.State == NodeStateInactive {
		n.Log(brain).Error("called 'Finish' while in state NodeStateInactive, something is wrong!")
		return
	}
	nodeData.State = NodeStateInactive
	n.Log(brain).Debug("Finish", zap.Bool("succeeded", succeeded))
	// TODO debug info
	parent := n.Parent(brain)
	if parent != nil {
		parent.ChildFinished(brain, n, succeeded)
	}
}

func (n *Node) Update(brain IBrain, eventType EventType, delta time.Duration) Result {
	return n.OnUpdate(brain, eventType, delta)
}

func (n *Node) OnUpdate(brain IBrain, eventType EventType, delta time.Duration) Result {
	log := n.Log(brain).With(zap.Int("event", int(eventType)), zap.String("target", n.delegator.Target), zap.String("method", n.delegator.Method), zap.Bool("script", n.delegator.Script != ""))
	log.Debug("OnUpdate")
	// 优先使用脚本,脚本不存在才使用委托
	if n.delegator.Script != "" {
		out, err := script.RunCode(n.id, n.scriptEnv(brain, eventType, delta))
		if err != nil {
			log.Error("OnUpdate run script error", zap.Error(err))
			return ResultFailed
		}
		// 无返回值则默认成功
		if out == nil {
			return ResultSucceeded
		}
		switch v := out.(type) {
		case bool:
			return lo.If(v, ResultSucceeded).Else(ResultFailed)
		case int64:
			return Result(v)
		case Result:
			return v
		default:
			log.Error("run script error:unSupport this type", zap.Any("return", v))
		}
	}
	// 脚本不存在,改用委托
	if n.delegator.Method == "" {
		log.Debug("method not found,return ResultFailed")
		if internal.GlobalConfig.ActionSuccessIfNotDelegate {
			return ResultSucceeded
		}
		return ResultFailed
	}
	// 若当前节点没有delegator target,使用root的delegator target作为默认
	target := lo.If(n.delegator.Target != "", n.delegator.Target).Else(n.root.Delegator().Target)
	if target == "" {
		log.Debug("delegator not found,return ResultFailed")
		return ResultFailed
	}
	// 交给委托执行
	ret := brain.(IBrainInternal).OnNodeUpdate(n.delegator.Target, n.delegator.Method, brain, eventType, delta)
	return ret
}

func (n *Node) scriptEnv(brain IBrain, eventType EventType, delta time.Duration) map[string]any {
	env := brain.(IBrainInternal).GetDelegates()
	env["blackboard"] = brain.Blackboard()
	env["eventType"] = eventType
	env["delta"] = delta
	return env
}

// CompositeAncestorFinished
//
//	@implement INode.CompositeAncestorFinished
//	@receiver n
//	@param brain
//	@param composite
func (n *Node) CompositeAncestorFinished(brain IBrain, composite IComposite) {
	n.INodeWorker.OnCompositeAncestorFinished(brain, composite)
}

func (n *Node) String(brain IBrain) string {
	return n.INodeWorker.OnString(brain)
}

func (n *Node) Path(brain IBrain) string {
	parent := n.Parent(brain)
	if parent != nil {
		return parent.Path(brain) + "/" + n.name
	}
	return n.name
}

// OnStart
//
//	@implement INodeWorker.OnStart
//	@receiver n
//	@param brain
func (n *Node) OnStart(brain IBrain) {
	n.Log(brain).Debug(" OnStart")
}

// OnAbort
//
//	@implement INodeWorker.OnAbort
//	@receiver n
//	@param brain
func (n *Node) OnAbort(brain IBrain) {
	n.Log(brain, LogWithUpstream).Debug(" OnAbort")
}

// OnCompositeAncestorFinished
//
//	@implement INodeWorker.OnCompositeAncestorFinished
//	@receiver n
//	@param brain
//	@param composite
func (n *Node) OnCompositeAncestorFinished(brain IBrain, composite IComposite) {
}

// PropertiesClassProvider
//
//	@implement INodeWorker.PropertiesClassProvider
//	@receiver n
//	@return any
func (n *Node) PropertiesClassProvider() any {
	return &map[string]any{}
}

// CanMountTo
//
//	@implement INodeWorker.CanMountTo
//	@receiver n
//	@return bool
func (n *Node) CanMountTo() bool {
	return false
}

// OnString
//
//	@implement INodeWorker.OnString 的回调
//	@param brain
//	@return string
func (n *Node) OnString(brain IBrain) string {
	return n.name + "(" + n.title + ")"
}

type LogOption = func(brain IBrain, logg *zap.Logger) *zap.Logger

func LogWithUpstream(brain IBrain, logg *zap.Logger) *zap.Logger {
	if brain == nil {
		return logg
	}
	upstreamAny := brain.(IBrainInternal).Context().Value(upstreamKey)
	if upstreamAny != nil {
		upstream := upstreamAny.(INode)
		return logg.With(zap.Namespace("upstream"), zap.String("name", upstream.Name()), zap.String("title", upstream.Title()), zap.Int("state", int(upstream.State(brain))))
	}
	return logg
}

func (n *Node) Log(brain IBrain, ops ...LogOption) *zap.Logger {
	logg := logger.Log.With(zap.String("name", n.name), zap.String("title", n.title))
	if brain == nil {
		return logg
	}
	logg = logger.Log.With(zap.Int("id", brain.Blackboard().(IBlackboardInternal).ThreadID()), zap.String("name", n.name), zap.String("title", n.title), zap.Int("state", int(n.State(brain))))
	for _, op := range ops {
		logg = op(brain, logg)
	}
	return logg
}

func Print(node INode, brain IBrain) {
	fmt.Print(node.Dump(brain))
}
func (n *Node) Dump(brain IBrain) string {
	nodeContent := fmt.Sprintf("%s<%d>", n.Title(), n.Memory(brain).State)
	tree := gotree.New(nodeContent)
	printStep(n.NodeWorkerAsNode(), brain, tree)
	return tree.Print()
}

func printStep(node INode, brain IBrain, printerParent gotree.Tree) {
	if node.Category() == CategoryDecorator {
		dec := node.(IDecorator)
		if dec.Decorated(brain) != nil {
			printStep(dec.Decorated(brain), brain, printerParent.Add(fmt.Sprintf("%s:%d", dec.Decorated(brain).Title(), node.State(brain))))
		}
		return
	}
	if node.Category() == CategoryComposite {
		comp := node.(IComposite)
		for _, child := range comp.Children() {
			printStep(child, brain, printerParent.Add(fmt.Sprintf("%s:%d", child.Title(), node.State(brain))))
		}
	}
}
