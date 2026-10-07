# Implement expiry-aware shared credential caching

[English](#english) | [中文](#中文)

## English

GitHub issue: #9.

Extend credentials with expiration; coalesce bounded refreshes, refresh early and never serve expired values. One canceled waiter must not cancel other waiters. Test with an injected clock/source and integrate the STS provider separately.

Acceptance includes implementation, meaningful offline behavior tests, English Go comments, executable Examples and equivalent bilingual guides. GitHub remains the source of execution status.

## 中文

GitHub issue：#9。

增加凭据过期信息，合并有界刷新、提前刷新、不返回已过期值；一个等待者取消不影响其他人。注入时钟/来源测试，STS provider 单独集成。

验收包含实现、有意义的离线行为测试、英文 Go 注释、可执行示例和对应双语指南；执行状态以 GitHub 为准。
