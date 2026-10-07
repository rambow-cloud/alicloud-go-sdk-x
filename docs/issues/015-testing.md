# [Feature]: Provide small mock interfaces and deterministic testing helpers

## English

### Affected areas

testing, tools

### Dependencies

None

### Problem and scope

Complete the shared runtime foundation before product generation. This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Provide operation-sized client interfaces and mockable paginator/waiter contracts.
- [ ] Provide scripted HTTP transport, response sequences and request assertions without cloud access.
- [ ] Provide deterministic clock/sleep helpers for retry, waiter and credential tests.
- [ ] Use helpers to verify the real shared runtime rather than only mocked application results.

### Documentation and verification

English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.

## 中文

### 领域与依赖

领域：testing, tools；依赖：None。

### 问题与范围

生成产品前完成共享基础。本 issue 交付以下能力，不宣称全量服务覆盖或真实云验证。

### 验收

- [ ] 提供操作级小接口和可 mock 的分页/waiter 契约。
- [ ] 提供不访问云的脚本 HTTP transport、响应序列和请求断言。
- [ ] 提供确定性 clock/sleep 用于重试、waiter 和凭据测试。
- [ ] 使用辅助工具验证真实共享 runtime，而不仅是业务 mock。

### 文档与验证

英文为主的 Go 注释、外部可执行 Examples、对应双语指南、离线行为测试及相关 CI 检查。
