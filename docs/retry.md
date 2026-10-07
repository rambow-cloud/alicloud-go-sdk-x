# Retry and backoff / 重试与退避

## English

Interrupted response reads are wrapped in ResponseReadError. EOF/UnexpectedEOF and
network read failures may retry under the bounded idempotent/replayable policy;
cancellation, JSON decoding failures and nonexistent DNS names do not retry.

Retries are disabled by default. Opt in with `retry.NewStandard` through client configuration. Defaults: three total attempts, 200ms base, 20s cap, full jitter and twenty shared budget tokens. Only explicitly idempotent operations with replayable bodies retry selected transient errors (429/500/502/503/504, throttling codes and network errors). Cancellation, deadlines, authentication and decoding errors do not retry. A retry consumes one token; success refunds one up to the initial budget. Sharing a policy shares its budget. Retry-After seconds and dates are honored up to the delay cap. Runtime context bounds operation time including all attempts and backoff. `retry.Wait` honors cancellation; injectable Sleep/Jitter enable deterministic tests. Custom implementations must be concurrency safe and bounded. AssumeRole is conservatively non-idempotent and is not retried by Standard.

## 中文

响应读取中断由 ResponseReadError 包装，EOF/UnexpectedEOF 及网络读取失败可以在明确幂等、
可重放、有尝试与预算限制的策略下重试；取消、JSON 解码失败及不存在的 DNS 名称不重试。

默认禁用重试。通过客户端配置显式选择 `retry.NewStandard`。默认共三次尝试、200ms 基础退避、20s 上限、full jitter 和二十个共享预算 token。只有明确幂等且请求体可重放的操作，才重试选定瞬态错误（429/500/502/503/504、限流码、网络错误）。取消、超时、认证和解码错误不重试。重试消耗一个 token；成功返还一个，最多达到初始预算。共享策略就共享预算。Retry-After 秒数或日期在延迟上限内生效。运行时 context 限制包含所有尝试和退避的总时间。`retry.Wait` 支持取消；注入 Sleep/Jitter 可确定测试。自定义实现必须并发安全且有界。AssumeRole 保守视为非幂等，Standard 不重试。
