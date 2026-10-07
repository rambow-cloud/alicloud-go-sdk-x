# Implement bounded retry and full-jitter backoff

[English](#english) | [中文](#中文)

## English

GitHub issue: #6.

Implement a replaceable policy, bounded attempts and per-policy retry budget; support Retry-After/context. Keep retries off by default and require idempotent replayable operations. Test independently, then integrate under #3.

Acceptance includes implementation, meaningful offline behavior tests, English Go comments, executable Examples and equivalent bilingual guides. GitHub remains the source of execution status.

## 中文

GitHub issue：#6。

提供可替换策略、有界尝试、每策略预算、Retry-After/context；默认关闭，仅用于幂等可重放操作。先独立测试，再由 #3 集成。

验收包含实现、有意义的离线行为测试、英文 Go 注释、可执行示例和对应双语指南；执行状态以 GitHub 为准。
