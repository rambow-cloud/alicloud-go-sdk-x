# 参与开发

[English](CONTRIBUTING.md)

## 工作流程

- 先阅读 [AGENTS.md](AGENTS.zh-CN.md) 和[开发路线](docs/development-path.zh-CN.md)。

- 开始编码前建立英文 issue，写清证据、范围、依赖、验收条件和文档要求。

- 分支使用 `issue/<number>-<slug>`，每个分支只处理一项可独立评审的工作。

- 完整交付使用 `Closes #N`，部分交付使用 `Refs #N`。issue 和 PR 只写英文。

- 离线草稿放在 docs/issues/，打开 PR 前同步；草稿编号不是 GitHub issue 编号。

## 文档

- 按[写作规范](docs/documentation-style.zh-CN.md)分别维护 name.md 和 name.zh-CN.md，互相链接并同步更新。

- 英文使用简短要点，中文自然表达。每份指南独立包含命令、默认值、限制和证据。

- Go 文档使用英文；实现时同步 doc.go、导出符号和字段注释，以及不依赖网络的外部 Example。

- 上游源码、README 和许可证保持原始内容。

## 生成文件

- 修改固定元数据、审核策略、模板或手写校验，不直接修改生成文件。

- 使用 `go run ./internal/cmd/sdkgen generate` 重新生成；完整 DSL 产品另执行 `go run ./internal/cmd/sdkgen product-generate`，然后运行检查。

- 导入来源需要明确的网络访问；生成和检查离线执行。见[生成器流程](docs/generator.zh-CN.md)。

- 保留已经验收的基础行为，评审模型变化后再检查公共 API。

## 验证

- 有意义的修改后，只执行相关检查一次。修改生成器时，先用 Node 22 执行前端检查和测试，再运行 Go 检查。

```powershell
# Node 22; only needed for generator changes
npm --prefix tools/darabonba run check
npm --prefix tools/darabonba test

go run ./internal/cmd/sdkgen check
go run ./internal/cmd/sdkgen product-check
go run ./internal/cmd/doccheck
node .github/scripts/check-doc-language.cjs
node --test .github/scripts/*.test.cjs
go vet ./...
go test ./...
```

- 使用 gofmt 格式化。Linux CI 增加竞态检查，Windows CI 验证可移植性。

- 单元测试不使用真实凭据，不调用云 API；不修改全局 Go/Git 配置。

- 只有发生有意义的修改或明确失败原因后才重复检查。工具检查结构，评审确认翻译、行为、兼容性和限制。
