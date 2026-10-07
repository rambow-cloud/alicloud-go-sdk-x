GitHub issue: #22.

# [Feature]: Generate typed ECS and STS clients and reviewed paginator/waiter adapters

## English

### Problem and evidence

Foundation #19 passed. The four reference operations remain handwritten. Official source: https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/. Development path: docs/generator.md; parent #8.

### Scope

Emit formatted models, request codecs, operation methods, minimal mock interfaces, client construction, English Go docs, deterministic offline Examples and bilingual product guides from validated IR. Generate reviewed pagination/waiter adapters and sensitive model redaction. Replace four handwritten reference operations while retaining prose-only validation extensions. Delegate all shared capabilities to the foundation runtime; preserve current public API and endpoint scope.

### Dependencies

#21

### Acceptance criteria

- [ ] Existing ECS/STS protocol fixtures and paginator/waiter/STS helper tests must pass unchanged. Test operation encoding, shared runtime idempotency and a synthetic service. No live accounts or network calls in generated Examples.
- [ ] Original English-primary code comments and equivalent English/Chinese Markdown updated together.
- [ ] Record exact commit and relevant successful verification before closure.

### Affected areas

tools, ecs, sts

## 中文

基础 #19 已通过，父任务 #8，开发路径见 docs/generator.md，依赖 #21。

由 IR 生成类型、编码、方法、小 mock 接口、构造器、英文 Go 注释、离线 Example 和双语指南，以及审核分页/waiter 适配器和脱敏；替换四个手写操作，保留说明规则校验扩展。已有协议及集成测试保持通过，不扩大端点范围。

验收包含英文注释、对应双语文档、行为测试及实际提交和验证证据；不仅交付接口。
