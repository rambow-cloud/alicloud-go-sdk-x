# [Maintenance]: Validate the unified runtime foundation before enabling generator development

## English

### Affected areas

core, tools

### Dependencies

#3, #4, #5, #6, #7, #9, #10, #11, #12, #13, #14, #15, #16, #17, #18

### Problem and scope

Complete the shared runtime foundation before product generation. This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Verify all eleven requested foundation capabilities with a handwritten ECS/STS reference.
- [ ] Run formatting, bilingual/API documentation checks, vet, Go tests and Linux race CI.
- [ ] Publish equivalent English/Chinese usage guides and a supported-operation matrix.
- [ ] Keep generator issue #8 blocked until the foundation acceptance criteria pass; do not claim full service coverage or live-cloud validation.

### Documentation and verification

English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.

## 中文

### 领域与依赖

领域：core, tools；依赖：#3, #4, #5, #6, #7, #9, #10, #11, #12, #13, #14, #15, #16, #17, #18。

### 问题与范围

生成产品前完成共享基础。本 issue 交付以下能力，不宣称全量服务覆盖或真实云验证。

### 验收

- [ ] 用手写 ECS/STS 参考验证十一项能力。
- [ ] 运行格式、双语/API 文档、vet、Go 测试和 Linux race CI。
- [ ] 发布对应双语使用指南及已支持操作矩阵。
- [ ] 基础验收通过前 #8 保持阻塞，不宣称全量覆盖或真实云验收。

### 文档与验证

英文为主的 Go 注释、外部可执行 Examples、对应双语指南、离线行为测试及相关 CI 检查。
