package behavior

// IObservingProperties 观察者装饰器属性
type IObservingProperties interface {
	GetAbortMode() AbortMode
}

// ObservingProperties 观察者装饰器属性
type ObservingProperties struct {
	BaseProperties
	AbortMode AbortMode `json:"abortMode"`
}

func (o *ObservingProperties) GetAbortMode() AbortMode {
	return o.AbortMode
}

// ObservingDecorator 观察者装饰器,实现了对条件的监听,和条件变化时的各种中断模式
//                 节点处于启用状态且条件不再被满足：
//                        Stops.NONE： 无
//                        Stops.SELF： 中断当前节点
//                        Stops.LOWER_PRIORITY： 无
//                        Stops.BOTH： 中断当前节点
//                节点处于停用状态且条件被满足时：
//                        Stops.NONE： 无
//                        Stops.SELF： 无
//                        Stops.LOWER_PRIORITY： 关闭当前启用的分支，启动此分支
//                        Stops.BOTH： 关闭当前启用的分支，启动此分支
type ObservingDecorator struct {
	Decorator
	IObservingWorker
}

// IObservingWorker ObservingDecorator 的回调,ObservingDecorator 的子类必须实现该接口
type IObservingWorker interface {
	// StartObserving 开始监听
	//  @param brain
	StartObserving(brain IBrain)
	// StopObserving 停止监听
	//  @param brain
	StopObserving(brain IBrain)
	// ConditionMet 计算条件是否满足
	//  @param brain
	//  @return bool
	ConditionMet(brain IBrain) bool
}

var _ IObservingWorker = (*ObservingDecorator)(nil)

// InitNodeWorker
//  @override Node.InitNodeWorker
//  @receiver c
//  @param worker
func (o *ObservingDecorator) InitNodeWorker(worker INodeWorker) {
	o.Decorator.InitNodeWorker(worker)
	// 强转,由框架本身保证实例化时传进来的worker是自己(自己实现了IContainerWorker接口,故强转不会panic)
	o.IObservingWorker = worker.(IObservingWorker)
}

// AbortMode 中断模式
//  @receiver o
//  @return AbortMode
func (o *ObservingDecorator) AbortMode() AbortMode {
	return o.properties.(IObservingProperties).GetAbortMode()
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (o *ObservingDecorator) PropertiesClassProvider() any {
	return &ObservingProperties{}
}

// OnStart
//  @override Node.OnStart
//  @receiver n
//  @param brain
func (o *ObservingDecorator) OnStart(brain IBrain) {
	o.Decorator.OnStart(brain)
	if o.AbortMode() != AbortModeNone {
		if !o.Memory(brain).Observing {
			o.Memory(brain).Observing = true
			o.IObservingWorker.StartObserving(brain)
		}
	}
	if !o.IObservingWorker.ConditionMet(brain) {
		o.Finish(brain, false)
	} else {
		o.decorated.Start(brain)
	}
}

// OnChildFinished
//  @override Container.OnChildFinished
//  @receiver r
//  @param brain
//  @param child
//  @param succeeded
func (o *ObservingDecorator) OnChildFinished(brain IBrain, child INode, succeeded bool) {
	o.Decorator.OnChildFinished(brain, child, succeeded)
	if o.IsInactive(brain) {
		o.Log().Fatal("ObservingDecorator cannot be inactive")
	}
	abortMode := o.AbortMode()
	if abortMode == AbortModeNone || abortMode == AbortModeSelf {
		o.stopObserving(brain)
	}
}
func (o *ObservingDecorator) stopObserving(brain IBrain) {
	if !o.Memory(brain).Observing {
		return
	}
	o.Memory(brain).Observing = false
	o.IObservingWorker.StopObserving(brain)
}

// OnCompositeAncestorFinished
//  @override Node.OnCompositeAncestorFinished
//  @receiver n
//  @param brain
//  @param composite
//
func (o *ObservingDecorator) OnCompositeAncestorFinished(brain IBrain, composite IComposite) {
	o.Decorator.OnCompositeAncestorFinished(brain, composite)
	o.stopObserving(brain)
}

// Evaluate 根据节点状态和条件满足来评估后续中断流程
//  @receiver o
//  @param brain
func (o *ObservingDecorator) Evaluate(brain IBrain) {
	conditionMet := o.IObservingWorker.ConditionMet(brain)
	mode := o.AbortMode()
	// 当条件变为不满足时,根据mode中断自己
	if o.IsActive(brain) && !conditionMet {
		if mode == AbortModeSelf || mode == AbortModeBoth {
			o.Abort(brain)
		}
		return
	}
	// 当条件变为满足时,根据mode中断低优先级分支
	if !o.IsActive(brain) && conditionMet {
		if mode != AbortModeLowerPriority && mode != AbortModeBoth {
			return
		}
		parent := o.Parent(brain)
		var child INode = o
		var composite IComposite
		for {
			var ok bool
			composite, ok = parent.(IComposite)
			if parent == nil || ok {
				break
			}
			child = parent
			parent = parent.Parent(brain)
		}
		if parent == nil {
			o.Log().Fatal("AbortMode is only valid when attached to a parent composite")
		}
		// TODO 平行节点是否要特殊限制
		o.stopObserving(brain)
		// 通知最近的组合祖先节点停止低优先级分支
		composite.AbortLowerPriorityChildrenForChild(brain, child)
	}
}

// StartObserving
//  @implement IObservingWorker.StartObserving
//  @receiver o
//  @param brain
func (o *ObservingDecorator) StartObserving(brain IBrain) {
	o.Log().Debug(o.String(brain) + " StartObserving")
}

// StopObserving
//  @implement IObservingWorker.StopObserving
//  @receiver o
//  @param brain
func (o *ObservingDecorator) StopObserving(brain IBrain) {
	o.Log().Debug(o.String(brain) + " StartObserving")
}

// ConditionMet panic
//  @implement IObservingWorker.ConditionMet
//  @receiver o
//  @param brain
//  @return bool
func (o *ObservingDecorator) ConditionMet(brain IBrain) bool {
	return false
}
