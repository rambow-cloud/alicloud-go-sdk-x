GitHub issue: #23.

# [Maintenance]: Enforce deterministic regeneration and generator acceptance in CI

## English

### Problem and evidence

Foundation #19 passed. The four reference operations remain handwritten. Official source: https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/. Development path: docs/generator.md; parent #8.

### Scope

Implement offline write/check commands with an explicit owned file set, full render before mutation, unmarked-file protection, stale-file detection and stable output. Add CI regeneration checks, drift/failure tests, bilingual contributor workflow and an acceptance map for parent #8. Record Linux race and Windows results before closing completed issues.

### Dependencies

#22

### Acceptance criteria

- [ ] Check mode detects edited/missing/stale files without mutation; invalid inputs produce no output changes. Run regeneration, documentation/language gates, vet, all tests/Examples and CI once after integration. Keep benchmark #20 separate.
- [ ] Original English-primary code comments and equivalent English/Chinese Markdown updated together.
- [ ] Record exact commit and relevant successful verification before closure.

### Affected areas

module:tools, module:docs, module:ci

## 中文

基础 #19 已通过，父任务 #8，开发路径见 docs/generator.md，依赖 #22。

实现离线 write/check、明确文件所有权、全量渲染后写入、无标记保护、多余文件检测及确定性输出；CI 添加再生成门禁、失败/漂移测试和双语贡献指南，验收后维护 #8。基准 #20 独立。

验收包含英文注释、对应双语文档、行为测试及实际提交和验证证据；不仅交付接口。
