# [Feature]: Implement a shared staged middleware pipeline

## English

### Affected areas

middleware, core

### Dependencies

None

### Problem and scope

Complete the shared runtime foundation before product generation. This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Provide Initialize, Build, Finalize and Deserialize stages with named middleware and function adapters.
- [ ] Separate once-per-operation hooks from once-per-attempt hooks; freeze client registrations.
- [ ] Preserve context, short-circuit errors and deterministic ordering; reject duplicate IDs per stage.
- [ ] Test ordering, retries, cancellation and concurrent client use.

### Documentation and verification

English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.

## 中文

### 领域与依赖

领域：middleware, core；依赖：None。

### 问题与范围

生成产品前完成共享基础。本 issue 交付以下能力，不宣称全量服务覆盖或真实云验证。

### 验收

- [ ] 提供 Initialize、Build、Finalize、Deserialize 阶段，以及具名 middleware 和函数适配器。
- [ ] 区分每操作一次与每尝试一次；冻结客户端注册。
- [ ] 保持 context、短路错误与确定顺序；同阶段拒绝重复 ID。
- [ ] 测试顺序、重试、取消和客户端并发。

### 文档与验证

英文为主的 Go 注释、外部可执行 Examples、对应双语指南、离线行为测试及相关 CI 检查。
