# [Feature]: Add optional OpenTelemetry operation and attempt instrumentation

## English

### Affected areas

telemetry, middleware

### Dependencies

#3, #11, #13

### Problem and scope

Complete the shared runtime foundation before product generation. This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Expose an opt-in middleware adapter with injected TracerProvider.
- [ ] Create one logical-operation span and child attempt spans with request metadata.
- [ ] Never set global providers/exporters or record credentials, raw bodies or query values by default.
- [ ] Keep the core import graph standard-library-only and test spans with an in-memory exporter.

### Documentation and verification

English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.

## 中文

### 领域与依赖

领域：telemetry, middleware；依赖：#3, #11, #13。

### 问题与范围

生成产品前完成共享基础。本 issue 交付以下能力，不宣称全量服务覆盖或真实云验证。

### 验收

- [ ] 提供注入 TracerProvider 的可选 middleware 适配器。
- [ ] 建立逻辑操作 span 和含请求元数据的子尝试 span。
- [ ] 不设置全局 provider/exporter，不默认记录凭据、原始 body 或 query 值。
- [ ] 保持核心导入仅标准库，使用内存 exporter 验证 spans。

### 文档与验证

英文为主的 Go 注释、外部可执行 Examples、对应双语指南、离线行为测试及相关 CI 检查。
