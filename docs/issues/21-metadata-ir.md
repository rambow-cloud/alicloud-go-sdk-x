GitHub issue: #21.

# [Feature]: Import pinned OpenAPI protocol metadata and validate generator IR

## English

### Problem and evidence

Foundation #19 passed. The four reference operations remain handwritten. Official source: https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/. Development path: docs/generator.md; parent #8.

### Scope

Implement a bounded explicit metadata importer, protocol-only snapshots, source/raw/snapshot SHA-256 provenance, strict manifests/overlays and validated RPC IR. Preserve selected public APIs through reviewed overlays; reject removed or changed selected fields, unsupported encoding/references and newly required inputs. Generation must be offline and independent of Go map iteration order.

### Dependencies

#19

### Acceptance criteria

- [ ] Offline importer HTTP fixtures, checksum tampering, schema drift, strict overlay errors and a non-ECS synthetic operation. Pin four official ECS/STS operations; supply paired design/usage docs.
- [ ] Original English-primary code comments and equivalent English/Chinese Markdown updated together.
- [ ] Record exact commit and relevant successful verification before closure.

### Affected areas

module:tools

## 中文

基础 #19 已通过，父任务 #8，开发路径见 docs/generator.md，依赖 #19。

实现有界显式元数据导入、仅协议快照、来源与双 SHA-256、严格 manifest/overlay 和 RPC IR；保留公开子集，拒绝选中字段删除/类型变化、不支持的编码/引用及新增必填字段。离线测试覆盖导入、篡改、schema 漂移、overlay 错误及非 ECS 合成操作。

验收包含英文注释、对应双语文档、行为测试及实际提交和验证证据；不仅交付接口。
