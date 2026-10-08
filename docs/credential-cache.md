# Credential cache

[中文](credential-cache.zh-CN.md)

- NewCache requires an explicit non-nil source and rejects typed-nil providers without retrieval.
- Prefer cached renewable STS role providers for applications; static/env sources remain explicit secondary choices, as described in [credentials](credentials.md).

- Wrap a source with `credentials.NewCache`.
- Zero options mean a one-minute early refresh window and a ten-second source deadline.
- Zero `ExpiresAt` is cached until `Invalidate`; known expired values are never returned.
- A still-valid value inside the early window is returned while one background refresh runs.
- Without a valid value, callers wait on the same refresh; a caller cancellation does not cancel that source call.
- Provider implementations must honor the shared deadline.
- Refresh errors do not publish a new value.
- Invalidation prevents an older refresh from publishing, then permits a new refresh after it finishes.
- Clock functions must be concurrency safe; `sdktest.Clock` supports deterministic expiry tests.
- Static providers check known expiration but never refresh.
- Use STS helpers for renewable temporary credentials.
