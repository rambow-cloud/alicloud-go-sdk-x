# Credentials / 凭据

## English

`Provider` retrieves copied credential snapshots with a context and must be concurrency safe. `NewStaticProvider` copies explicit keys. `EnvProvider` reads ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET and optional ALIBABA_CLOUD_SECURITY_TOKEN each time. All three absent returns `ErrNotFound`; any incomplete configuration returns `ErrMissingCredentials`. `NewChain` copies an explicit ordered provider list, skips only `ErrNotFound` and stops on other failures or invalid snapshots. There is no implicit file, process, metadata or role discovery. Use `ProviderFunc` for a custom source and preserve cancellation errors. Default formatting redacts credential values, but exported fields remain sensitive and must not be logged directly.

## 中文

`Provider` 使用 context 返回复制的凭据快照，必须并发安全。`NewStaticProvider` 复制显式密钥。`EnvProvider` 每次读取 ALIBABA_CLOUD_ACCESS_KEY_ID、ALIBABA_CLOUD_ACCESS_KEY_SECRET 和可选 ALIBABA_CLOUD_SECURITY_TOKEN。三项都不存在返回 `ErrNotFound`；不完整配置返回 `ErrMissingCredentials`。`NewChain` 复制显式有序 provider 列表，只跳过 `ErrNotFound`，其他失败或无效快照立即终止。不会隐式发现文件、进程、元数据或角色。用 `ProviderFunc` 实现自定义来源并保留取消错误。默认格式化隐藏凭据值，但导出字段仍敏感，不得直接记录。
