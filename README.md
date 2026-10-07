# alicloud-go-sdk-x

[English](#english) | [中文](#中文)

## English

[Go Reference](https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x) ·
[CI](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/workflows/ci.yml) ·
[Development path](docs/development-path.md)

An independent Alibaba Cloud SDK for Go. Requires Go 1.27 and encoding/json/v2.
Development is issue-driven, runtime-first, with English-primary Go docs and paired
English/Chinese guides. It is not an official SDK. APIs may change before v1.

The runtime foundation is being implemented; check the GitHub issue milestone and
supported-operation matrix for actual completed coverage. Do not assume full product
coverage or live-cloud acceptance. No version tag is published by this task.

Foundation: shared middleware, endpoints, structured errors, bounded retry, credential
providers/cache/chain, STS helper, unified pagination/waiters, mock interfaces, testing
helpers and opt-in OpenTelemetry. Generator implementation follows foundation acceptance.
See docs/design.md, docs/research.md, docs/issue-management.md and docs/releasing.md.

Run go run ./internal/cmd/doccheck, node .github/scripts/check-doc-language.cjs,
go vet ./..., go test ./..., and node --test .github/scripts/*.test.cjs.
Linux CI runs race detection; Windows CI verifies portability. Public packages provide
offline external Examples. Documentation checks cover structure; reviewers check semantics.
Read AGENTS.md and CONTRIBUTING.md before contributing. MIT: see LICENSE.

## 中文

[Go 文档](https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x) ·
[CI](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/workflows/ci.yml) ·
[开发路径](docs/development-path.md)

独立阿里云 Go SDK，要求 Go 1.27 和直接使用 JSON v2。采用 issue 驱动、runtime 优先，
Go 注释英文为主，使用指南中英文对应。本项目非官方 SDK，v1 前 API 可能变化。

基础运行时正在实现，实际完成范围以 GitHub 里程碑和支持矩阵为准，不假定全产品覆盖或真实云验收。
本次任务不发布版本标签。基础包括共享 middleware、endpoint、结构化错误、有界重试、凭据 provider/cache/chain、
STS helper、统一分页/waiter、mock 接口、测试辅助和可选 OpenTelemetry；基础验收后才建设 generator。
设计、调研、issue 维护和发布步骤见英文章节所列文档。

运行英文章节的文档、双语、vet、Go 与自动化测试命令；Linux CI 使用 race，Windows 验证可移植性。
公共包提供离线外部 Examples；检查工具验证结构，评审检查语义。贡献前阅读约束与贡献指南，MIT 许可证见 LICENSE。
