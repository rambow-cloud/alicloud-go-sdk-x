# [Feature]: Implement replaceable endpoint resolution with explicit rules

## English

### Affected areas

endpoints

### Dependencies

None

### Problem and scope

Complete the shared runtime foundation before product generation. This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Expose a context-aware resolver and function adapter; explicit endpoint overrides take precedence.
- [ ] Validate HTTPS and reject userinfo, fragments and unexpected endpoint query strings.
- [ ] Use documented ECS regional and STS endpoints; unsupported service/region combinations fail explicitly.
- [ ] Test resolver cancellation, overrides, rule ownership and request signing against the resolved host.

### Documentation and verification

English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.

## 中文

### 领域与依赖

领域：endpoints；依赖：None。

### 问题与范围

生成产品前完成共享基础。本 issue 交付以下能力，不宣称全量服务覆盖或真实云验证。

### 验收

- [ ] 提供 context-aware resolver 和函数适配器；显式 endpoint 优先。
- [ ] 校验 HTTPS，拒绝 userinfo、fragment 和意外 query。
- [ ] 使用有文档依据的 ECS 地域/STS 地址；不支持的组合明确失败。
- [ ] 测试取消、覆盖、规则所有权和最终 host 签名。

### 文档与验证

英文为主的 Go 注释、外部可执行 Examples、对应双语指南、离线行为测试及相关 CI 检查。
