# Unified waiters / 统一 Waiter

## English

`waiter.New[T]` combines a context-aware fetcher with an acceptor returning Retry, Success or Failure. `Wait` requires a positive maximum duration including fetches, operation retries and capped exponential sleeps (default 1s to 5s). Caller cancellation/deadlines pass through; the engine's expiry supports both ErrTimeout and context.DeadlineExceeded, preserving the last fetch error. FailureError supports ErrFailure and its fetch cause. Late results cannot succeed after the deadline. Clock/Sleep injection supports deterministic tests; callbacks must be concurrency safe if sharing a waiter. `ecs.NewInstanceRunningWaiter` requires 1..50 distinct IDs, copies input and requests the first status page at size 50. All requested IDs must be present and Running. Empty/missing, Pending, Starting, Stopping and Stopped retry. Unknown states, duplicate response IDs and API errors fail. It does not infer resource deletion or silently treat absence as success; larger sets require separate waiters. No live cloud validation is claimed.

## 中文

`waiter.New[T]` 组合遵守 context 的 fetcher 与返回 Retry、Success、Failure 的 acceptor。`Wait` 要求正数最大时长，包含 fetch、操作重试和有上限的指数睡眠（默认 1s 到 5s）。调用者取消/超时直接传播；引擎自身超时同时支持 ErrTimeout 和 context.DeadlineExceeded，并保留最后 fetch 错误。FailureError 支持 ErrFailure 和 fetch cause。截止后到达的结果不能成功。注入 Clock/Sleep 支持确定测试；共享 waiter 时 callback 必须并发安全。`ecs.NewInstanceRunningWaiter` 要求 1..50 个不同 ID，复制输入，以 size 50 获取首个状态页。所有请求 ID 必须存在且为 Running。空/缺失、Pending、Starting、Stopping、Stopped 会重试。未知状态、重复响应 ID 和 API 错误失败。不会推测资源删除，也不会把缺失视为成功；更大的集合需拆分 waiter。不宣称真实云验证。
