# Unified waiters / 统一 Waiter

## English

Construct `ecs.NewInstanceRunningWaiter(api, optFns...)` once, then call
`Wait(ctx, input, maxWait, optFns...) error` or
`WaitForOutput(ctx, input, maxWait, optFns...) (*ecs.DescribeInstanceStatusOutput, error)`.
Concurrent waits have private inputs, polling state and option snapshots. Inputs and
option registrations are copied per invocation and poll; callers must not mutate their
input concurrently. APIs, transports, clocks and callbacks remain shared and must be
concurrency safe. Constructor validation returns an error, an intentional early-v0
difference from AWS alongside the similar calling convention.

`InstanceRunningWaiterOptions` supplies MinDelay/MaxDelay (zero selects 1s/5s),
ClientOptions applied to every poll, deterministic Now/Sleep seams and Retryable.
Invocation overrides do not persist. Nil callbacks and invalid delays fail before
polling. Retryable receives the bounded poll context, a fresh input copy, output and
fetch error: true retries, false succeeds, and an error fails with its inspectable
cause. Nil retains the reviewed default. A failed fetch cannot become success;
cancellation/expiry are checked before and after acceptance. Callbacks must not block.
Changing acceptance is an explicit caller policy, not a new service guarantee.

Each invocation requires 1..50 nonempty distinct IDs and a first-page request. It polls
page 1 at size 50. The default requires every requested ID present and Running.
Missing IDs, empty results and Pending/Starting/Stopping/Stopped retry. Unknown states,
duplicate response IDs and API errors fail. Absence never implies success or deletion.
Larger sets require separate waits. A custom acceptor retains these input/page bounds.

```go
running, err := ecs.NewInstanceRunningWaiter(client)
if err != nil { return err }
input := &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i-example"}}
if err := running.Wait(ctx, input, time.Minute); err != nil { return err }
out, err := running.WaitForOutput(ctx, input, time.Minute)
```

The shared `waiter.New[T]` engine combines a context-aware fetcher with Retry/Success/
Failure acceptance. Its Wait continues returning `(T, error)`. A positive maximum
duration includes fetches, operation retries and capped exponential sleeps. Caller
cancellation/deadlines pass through; own expiry supports ErrTimeout and
context.DeadlineExceeded, preserving the last retryable fetch error. FailureError
supports ErrFailure and a fetch/custom-acceptor cause. Failed generated waits return
nil output. Offline examples and regression fixtures cover behavior; no live cloud
validation is claimed.

Migration: the old generated constructor bound input and waiter.Options, and Wait
returned output. Move input/maxWait to each call, configure dedicated options through
functional callbacks, and use WaitForOutput when the response is needed. The generic
engine API is preserved. See [the remediation path](aws-style-remediation.md).

## 中文

一次构造 `ecs.NewInstanceRunningWaiter(api, optFns...)`，再调用
`Wait(ctx, input, maxWait, optFns...) error` 或
`WaitForOutput(ctx, input, maxWait, optFns...) (*ecs.DescribeInstanceStatusOutput, error)`。
并发等待拥有独立输入、轮询状态与选项快照；每次调用及轮询复制输入和注册项，调用者不得并发
修改输入。API、transport、时钟和回调仍共享，须并发安全。构造校验返回 error，这是采用相似
调用范式时有意保留的早期 v0 与 AWS 差异。

`InstanceRunningWaiterOptions` 提供 MinDelay/MaxDelay（零值默认 1s/5s）、每次轮询的
ClientOptions、确定测试的 Now/Sleep 及 Retryable。当前调用覆盖不会保留；nil 回调及非法
延迟在轮询前失败。Retryable 接收有期限的轮询 context、新输入副本、输出和 fetch 错误：
true 继续，false 成功，返回 error 则失败并保留 cause；nil 使用审核默认规则。失败 fetch 不可
转成成功；acceptor 前后均检查取消/超时。回调不得阻塞，覆盖属于调用者明确策略，不构成新服务承诺。

每次输入要求 1..50 个非空且不同的 ID，并使用首个状态页；轮询固定 page 1、size 50。
默认要求所有请求 ID 均存在且 Running。缺失、空结果及 Pending/Starting/Stopping/Stopped
继续；未知状态、重复响应 ID、API 错误失败。缺失不表示成功或删除；更大集合拆分等待。
自定义 acceptor 仍保留输入及页边界。调用示例与英文章节相同。

共享 `waiter.New[T]` 组合遵守 context 的 fetcher 与 Retry/Success/Failure acceptor，
其 Wait 保持返回 `(T, error)`。正数最大时长包含 fetch、操作重试及有上限指数睡眠。
调用者取消/期限直接传播；自身超时同时支持 ErrTimeout 和 context.DeadlineExceeded，保留
最后可重试 fetch 错误。FailureError 支持 ErrFailure 和 fetch/自定义 acceptor 的 cause；
生成 waiter 失败返回 nil 输出。离线 Example 与回归覆盖行为，不宣称真实云验证。

迁移：旧生成构造函数绑定输入及 waiter.Options，Wait 返回输出；改为每次调用传入 input/
maxWait，通过专属 functional options 配置，需要响应时用 WaitForOutput。通用引擎 API
保持。见[修复路径](aws-style-remediation.md)。
