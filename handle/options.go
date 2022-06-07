package handle

type (
	options struct {
		name           string              // component name
		nameFunc       func(string) string // rename handler name
		TaskGoProvider func(task func())   // 异步任务派发线程提供者
	}

	// Option used to customize handler
	Option func(options *options)
)

// WithName used to rename component name
func WithName(name string) Option {
	return func(opt *options) {
		opt.name = name
	}
}

// WithNameFunc override handler name by specific function
// such as: strings.ToUpper/strings.ToLower
func WithNameFunc(fn func(string) string) Option {
	return func(opt *options) {
		opt.nameFunc = fn
	}
}

// WithTaskGoProvider 注册延迟动态绑定的异步任务派发线程
//  @param taskGoProvider
//  @return Option
func WithTaskGoProvider(taskGoProvider func(task func())) Option {
	return func(opt *options) {
		opt.TaskGoProvider = taskGoProvider
	}
}
