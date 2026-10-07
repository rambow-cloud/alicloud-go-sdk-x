# 建立 pkg.go.dev 文档与 issue 驱动 CI 门禁

### Affected areas

core, credentials, tools

### 前置依赖

无；初始化交付。

状态：GitHub issue #2（执行状态以 GitHub 为准）。

## 问题与证据

OpenAPI #221 的非原生注释引起工具解析失败；#225 报告依赖许可证缺失。文档必须随代码交付。

来源：[上游证据](https://github.com/aliyun/darabonba-openapi/issues/221)。具体边界见 [调研](../research.md)。

## 范围

包/符号/字段注释检查、可执行外部示例、CI、issue 表单、PR 模板、发布步骤；不自动发布版本。

## 验收条件

- [ ] 公开包必须有 doc.go，导出类型、方法、字段、常量均有以名称开头的注释
- [ ] 每个公开包至少一个带确定输出的外部 Example，go test 执行示例
- [ ] CI 在 Go 1.27 Linux/Windows 运行格式、文档、vet、测试；Linux 执行 race
- [ ] 按 powertools-lambda-go 的类型与 module 分类维护 issue/label，加入 priority/status
- [ ] PR 必须用 Closes/Refs 引用本仓库现有 issue；关联失败使门禁失败
- [ ] 标准 LICENSE；发布说明包含 pkg.go.dev Request 步骤与版本规则

## 文档要求

每项公开 API 同时交付 doc.go、导出符号/字段注释和可执行外部 Example。
更新 README 支持范围与相关设计/发布说明。文档完成与代码完成属于同一交付。

## 验证计划

按验收条件建立离线行为测试，运行文档检查、vet 和相关测试一次。
涉及并发时由 Linux CI 运行 race；上线或 pkg.go.dev 页面由浏览器确认。
