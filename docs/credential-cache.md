# Credential cache / 凭据缓存

## English

NewCache requires an explicit non-nil source and rejects typed-nil providers without
retrieval. Prefer cached renewable STS role providers for applications; static/env
sources remain explicit secondary choices, as described in [credentials](credentials.md).

Wrap a source with `credentials.NewCache`. Zero options mean a one-minute early refresh window and a ten-second source deadline. Zero `ExpiresAt` is cached until `Invalidate`; known expired values are never returned. A still-valid value inside the early window is returned while one background refresh runs. Without a valid value, callers wait on the same refresh; a caller cancellation does not cancel that source call. Provider implementations must honor the shared deadline. Refresh errors do not publish a new value. Invalidation prevents an older refresh from publishing, then permits a new refresh after it finishes. Clock functions must be concurrency safe; `sdktest.Clock` supports deterministic expiry tests. Static providers check known expiration but never refresh. Use STS helpers for renewable temporary credentials.

## 中文

NewCache 要求显式非 nil 来源，构造期间拒绝 typed-nil provider，且不读取凭据。应用优先采用
缓存的可刷新 STS role provider；静态/环境来源仍为显式次要选项，见[凭据](credentials.md)。

使用 `credentials.NewCache` 包装来源。零选项表示提前一分钟刷新、来源调用限时十秒。零 `ExpiresAt` 缓存到 `Invalidate`；绝不返回已知过期值。提前窗口内仍有效的值立即返回，同时进行一次后台刷新。没有有效值时，调用者等待同一次刷新；某调用者取消不会取消来源调用。Provider 必须遵守共享超时。刷新错误不发布新值。失效操作防止旧刷新发布，旧调用结束后才允许新刷新。时钟函数必须并发安全；`sdktest.Clock` 支持确定的过期测试。静态 provider 检查已知过期时间，但不刷新。可更新临时凭据使用 STS helper。
