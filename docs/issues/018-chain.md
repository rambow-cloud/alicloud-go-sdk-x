# [Feature]: Add an explicit composable credential chain

## English

### Affected areas

credentials

### Dependencies

None

### Problem and scope

Complete the shared runtime foundation before product generation. This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Distinguish an absent credential source from an incomplete or invalid source.
- [ ] Skip only absent sources and stop on configuration, retrieval or cancellation errors.
- [ ] Support environment and explicitly supplied providers without implicit metadata or process execution.
- [ ] Test provider precedence, cancellation and errors without leaking secrets.

### Documentation and verification

English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.

## 中文

### 领域与依赖

领域：credentials；依赖：None。

### 问题与范围

生成产品前完成共享基础。本 issue 交付以下能力，不宣称全量服务覆盖或真实云验证。

### 验收

- [ ] 区分凭据来源不存在与来源不完整/无效。
- [ ] 仅跳过不存在；配置、获取或取消错误立即停止。
- [ ] 支持环境与显式来源，不隐式执行进程或访问 metadata。
- [ ] 测试优先级、取消和错误，避免秘密泄漏。

### 文档与验证

英文为主的 Go 注释、外部可执行 Examples、对应双语指南、离线行为测试及相关 CI 检查。
