# behavior
golang event driver behavior tree
features:
- 事件驱动
- 共享实例的行为数:所有节点无状态,状态由黑板管理
- 并发:各个AI在独立子纤程执行互不干扰
- 脚本:任务节点和条件节点的委托支持使用脚本

## v1.1.30 依赖迁移

- 直接依赖 `github.com/alkaid/ants/v2 v2.12.1-ak-4`，调用方传入的 `PoolWithID` 也需使用此模块路径。不再需要旧 ants 路径的 `replace`。
- 时间轮改用 `github.com/zeromicro/go-zero v1.8.5` 的 `core/collection`，移除 `github.com/alkaid/timingwheel`。
- `Brain.After`、`Brain.Cron` 和 `NodeMemory.CronTask` 的类型改为 `*github.com/alkaid/behavior/timer.Timer`；显式声明返回类型的调用方需同步修改。调度选项改用 `timer.WithPool` 和 `timer.WithGoID`。
- 时间轮池在构造时启动，不再调用 `TimeWheelPool.Start`。`TimeWheelInstance`、`Get` 和 `GetRandom` 返回 go-zero 时间轮，通常应通过 `timer.After` / `timer.Cron` 调度行为树回调。

随机离差只在创建任务时采样一次。循环任务按原计划时间推进，在回调开始前安排下一次；调度精度受配置的时间轮 tick 限制。非正延时在下一 tick 异步执行，不进行无间隔循环。

`Brain` 的回调仍投递到黑板 `ThreadID` 对应的 ants 池。`Timer.Stop` 可以取消尚未开始执行的排队回调和后续循环，不等待已开始的回调；回调内可以调用 `Stop`。取消会立即释放定时器保存的闭包，避免 go-zero 延迟清理槽位期间继续持有 AI 对象。

新版 ants 的满队列会对阻塞提交施加背压。behavior 的 `Brain.Go`、黑板通知和指定池的定时器通过同一内部 FIFO 派发：每个池、ID 的任务按入队顺序提交，准入等待发生在独立派发协程中，业务回调仍在 ants worker 上串行执行。这样，黑板在回调内批量写入不会等待自身队列腾出空间，一个 AI 满队列也不会阻塞同 tick 的其他 AI。派发队列清空后会释放；持续超出消费能力的输入仍需要业务侧限流，待投递任务会占用内存。

直接调用 `PoolWithID.Submit` 仍遵循 ants 的阻塞语义；同 ID 的回调不能同步等待排在自己后面的回调完成。新版 ants 将运行中任务替换阈值独立为 `RunningTaskTimeout`（默认 5 分钟），与旧版 `ExpiryDuration` 的语义不同。

注入池关闭前应停止行为树和定时任务。派发层收到 `ErrPoolClosed` 后，会拒绝该 ID 尚未投递的整条队列；ants 的 `ReleaseTimeout` 只等待其已接收的任务，不等待 behavior 尚未投递的任务或错误回调。behavior 不支持调用方关闭后直接 `Reboot` 复用同一池实例：ants 未公开生命周期代次，无法识别派发协程运行前发生的快速关闭与重启。需要再次使用时应创建新池，避免旧任务进入新的执行环境。

## 验证

```sh
go build ./...
go vet ./...
go mod verify
go test -race -count=5 -timeout 60s ./internal/dispatch ./timer ./thread
go test -race -count=20 -timeout 60s . -run '^TestBrainTimersShareThread$'
go test -race -count=20 -timeout 60s . -run '^TestBlackboardBatchInCallback$'
go test -race -timeout 15s . -run '^TestRunTree_LoadFromStr$'
make pre-commit
```

`tree_test.go` 还包含依赖本机 `/tmp/export/*.json` 或持续运行的旧演示测试，全量 `go test ./...` 需要相应环境和退出条件，不能作为当前可独立完成的自动化测试集。原 `make pre-commit` 使用 golangci-lint v1 配置；使用 v2 时需先迁移配置，并保留原 depguard 的 logrus 黑名单语义。以上限制在本次迁移前已存在。
