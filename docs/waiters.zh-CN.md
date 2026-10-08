# 统一 Waiter

[English](waiters.md)

- #37 基于审核稀疏策略为完整 DSL `service/ecs` 输出此契约，原 `services/ecs` 参考仍可用，明确导入所需包。
- 覆盖与来源绑定见[能力策略](capability-policy.zh-CN.md)。

- 一次构造 `ecs.NewInstanceRunningWaiter(api, optFns...)`，再调用 `Wait(ctx, input, maxWait, optFns...) error` 或 `WaitForOutput(ctx, input, maxWait, optFns...) (*ecs.DescribeInstanceStatusOutput, error)`。
- 并发等待拥有独立输入、轮询状态与选项快照；每次调用及轮询复制输入和注册项，调用者不得并发修改输入。
- API、HTTP 传输实现、时钟和回调仍共享，须并发安全。
- 构造校验返回 error，这是采用相似调用范式时有意保留的早期 v0 与 AWS 差异。

- `InstanceRunningWaiterOptions` 提供 MinDelay/MaxDelay（零值默认 1s/5s）、每次轮询的 ClientOptions、确定测试的 Now/Sleep 及 Retryable。
- 当前调用覆盖不会保留；nil 回调及非法延迟在轮询前失败。
- Retryable 接收有期限的轮询 context、新输入副本、输出和 fetch 错误： true 继续，false 成功，返回 error 则失败并保留底层错误；nil 使用审核默认规则。
- 失败 fetch 不可转成成功；acceptor 前后均检查取消/超时。
- 回调不得阻塞，覆盖属于调用者明确策略，不构成新服务承诺。

- 每次输入要求 1..50 个非空且不同的 ID，并使用首个状态页；轮询固定 page 1、size 50。
- 默认要求所有请求 ID 均存在且 Running。
- 缺失、空结果及 Pending/Starting/Stopping/Stopped 继续；未知状态、重复响应 ID、API 错误失败。
- 缺失不表示成功或删除；更大集合拆分等待。
- 自定义 acceptor 仍保留输入及页边界。
- 调用示例与所列参考资料一致。

- 共享 `waiter.New[T]` 组合遵守 context 的 fetcher 与 Retry/Success/Failure acceptor， 其 Wait 保持返回 `(T, error)`。
- 正数最大时长包含 fetch、操作重试及有延迟上限的指数退避。
- 调用者取消/期限直接传播；自身超时同时支持 ErrTimeout 和 context.DeadlineExceeded，保留最后可重试 fetch 错误。
- FailureError 支持 ErrFailure 和 fetch/自定义 acceptor 的底层错误； 生成状态等待器失败返回 nil 输出。
- 离线 Example 与回归覆盖行为，不表示真实云验证。

- 迁移：旧生成构造函数绑定输入及`waiter.Options`，Wait 返回输出；改为每次调用传入 input/ maxWait，通过专属函数式选项配置，需要响应时用 WaitForOutput。
- 通用引擎 API 保持。
- 见[修复路径](aws-style-remediation.zh-CN.md)。

## 可运行命令与示例

```go
running, err := ecs.NewInstanceRunningWaiter(client)
if err != nil { return err }
input := &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i-example"}}
if err := running.Wait(ctx, input, time.Minute); err != nil { return err }
out, err := running.WaitForOutput(ctx, input, time.Minute)
```
