# Retry and backoff

[中文](retry.zh-CN.md)

- Interrupted response reads are wrapped in ResponseReadError.
- EOF/UnexpectedEOF and network read failures may retry under the bounded idempotent/replayable policy; cancellation, JSON decoding failures and nonexistent DNS names do not retry.

- Retries are disabled by default.
- Opt in with `retry.NewStandard` through client configuration.
- Defaults: three total attempts, 200ms base, 20s cap, full jitter and twenty shared budget tokens.
- Only explicitly idempotent operations with replayable bodies retry selected transient errors (429/500/502/503/504, throttling codes and network errors).
- Cancellation, deadlines, authentication and decoding errors do not retry.
- A retry consumes one token; success refunds one up to the initial budget.
- Sharing a policy shares its budget.
- Retry-After seconds and dates are honored up to the delay cap.
- Runtime context bounds operation time including all attempts and backoff. `retry.Wait` honors cancellation; injectable Sleep/Jitter enable deterministic tests.
- Custom implementations must be concurrency safe and bounded.
- AssumeRole is conservatively non-idempotent and is not retried by Standard.
