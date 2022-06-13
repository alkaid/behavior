package task

type IDynamicSubtreeProperties interface {
	GetTag() string
}

// DynamicSubtreeProperties 子树容器属性
type DynamicSubtreeProperties struct {
	SubtreeProperties
	Tag string `json:"tag"` // 标签,用于识别子树
}

func (p *DynamicSubtreeProperties) GetTag() string {
	return p.Tag
}

// IDynamicSubtree 动态子树容器
//  可以在运行时动态更换子节点
type IDynamicSubtree interface {
	ISubtree
	IDynamicSubtreeProperties
}

// DynamicSubtree 动态子树容器
//  可以在运行时动态更换子节点
type DynamicSubtree struct {
	Subtree
}

// PropertiesClassProvider
//  @implement INodeWorker.PropertiesClassProvider
//  @receiver n
//  @return any
func (t *DynamicSubtree) PropertiesClassProvider() any {
	return &DynamicSubtree{}
}

// CanDynamicDecorate 标志可以运行时动态挂载子节点
//  @implement bcore.IDecoratorWorker .CanDynamicDecorate
//  @receiver t
//  @return bool
func (t *DynamicSubtree) CanDynamicDecorate() bool {
	return true
}
func (t *DynamicSubtree) GetTag() string {
	return t.Properties().(IDynamicSubtreeProperties).GetTag()
}
