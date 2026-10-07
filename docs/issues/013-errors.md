# [Feature]: Provide structured operation errors and response metadata

## English

### Affected areas

core

### Dependencies

None

### Problem and scope

Complete the shared runtime foundation before product generation. This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Retain APIError compatibility and add Unwrap-capable operation errors.
- [ ] Expose service, operation, request ID, HTTP status and attempt count without default raw-body logging.
- [ ] Preserve errors.Is for cancellation and errors.As for service and transport errors.
- [ ] Handle non-JSON service failures and reject malformed successful JSON using encoding/json/v2.

### Documentation and verification

English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.

## 中文

### 领域与依赖

领域：core；依赖：None。

### 问题与范围

生成产品前完成共享基础。本 issue 交付以下能力，不宣称全量服务覆盖或真实云验证。

### 验收

- [ ] 保留 APIError，增加可 Unwrap 的操作错误。
- [ ] 提供产品、操作、request ID、HTTP 状态和尝试次数，不默认记录原始 body。
- [ ] 保留 errors.Is 取消及 errors.As 服务/传输错误。
- [ ] 处理非 JSON 服务错误；用 JSON v2 拒绝无效成功响应。

### 文档与验证

英文为主的 Go 注释、外部可执行 Examples、对应双语指南、离线行为测试及相关 CI 检查。
