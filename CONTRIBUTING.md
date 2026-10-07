# Contributing / 贡献流程

[English](#english) | [中文](#中文)

## English

Start every behavior change with an issue describing evidence, scope, dependencies,
observable acceptance criteria, and documentation requirements. Use issue/<number>-<slug>.
Keep issue and PR content English-primary. Complete PRs use Closes #N; partial work uses Refs #N.
Keep English and Chinese Markdown sections equivalent and Go comments English-primary.
Provide doc.go, exported symbol/field comments, and runnable external Examples together.

Run relevant checks once after a meaningful change:
```
go run ./internal/cmd/sdkgen check
go run ./internal/cmd/doccheck
node .github/scripts/check-doc-language.cjs
go vet ./...
go test ./...
node --test .github/scripts/*.test.cjs
```
Use gofmt before checking formatting; Linux CI adds race detection. Never run live cloud
tests with real credentials as part of unit tests. Do not change global Go/Git configuration.
Only repeat checks after a meaningful change or a understood failure. Documentation checks
verify structure; reviewers verify accuracy, compatibility, defaults and limitations.
Offline drafts under docs/issues/ must be synchronized before a PR; draft IDs are not issue numbers.
Read AGENTS.md and docs/development-path.md before implementing. Generator work is gated on the runtime foundation.

For generated clients edit reviewed metadata/overlays, templates or handwritten validation
extensions, then run `go run ./internal/cmd/sdkgen generate` before the checks. Generated
files are owned by sdkgen; do not edit them directly. The importer is explicitly networked,
but generation/check are offline. See [the generator workflow](docs/generator.md).

## 中文

每项行为修改先建 issue，写明证据、范围、依赖、可观察的验收条件及文档要求。分支使用 issue/<number>-<slug>。
Issue/PR 以英文为主；完整交付使用 Closes #N，部分交付使用 Refs #N。
Markdown 的中英文章节保持语义对应，Go 注释以英文为主；doc.go、导出符号/字段注释和可执行外部 Example 一起交付。

有意义的变更后，文档检查、双语检查、vet、Go 测试及自动化测试各运行一次，命令见上方英文章节。
格式检查前使用 gofmt；Linux CI 增加 race。单元测试不使用真实凭据访问云资源，也不修改全局 Go/Git 配置。
仅在有意义的修改或已理解的失败后重复检查。检查工具验证结构，评审验证准确性、兼容性、默认值与限制。
离线草稿放 docs/issues/，PR 前同步；草稿 ID 不是 GitHub 编号。
实现前阅读 AGENTS.md 与开发路径；基础 runtime 验收后才开发 generator。

生成客户端修改审核的元数据/overlay、模板或手写校验扩展，检查前运行
`go run ./internal/cmd/sdkgen generate`。生成文件由 sdkgen 管理，不直接修改。
导入需显式联网，生成/check 离线；参见[生成器流程](docs/generator.md)。
