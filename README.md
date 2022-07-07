# behavior
golang event driver behavior tree
features:
- 事件驱动
- 共享实例的行为数:所有节点无状态,状态由黑板管理
- 并发:各个AI在独立子纤程执行互不干扰
- 脚本:任务节点和条件节点的委托支持使用脚本