# 贡献流程

本项目采用 issue-driven development。代码、文档与验收属于同一交付。

1. 先创建 issue，填写问题、证据、范围、验收条件与文档要求。
   大范围设计先固定协议或 schema，讨论结论写回 issue。
2. 使用 `issue/<number>-<slug>` 分支；初始化工作作为 bootstrap 例外。
3. 针对验收条件实现最小完整改动；行为缺陷添加可复现的回归测试。
4. 同时更新包注释、导出符号文档、可执行 Example 和 README。
5. 运行下列检查各一次；只在失败原因已理解或实现改变时重跑。
6. PR 正文完整交付写 `Closes #<number>`，部分交付写 `Refs #<number>`，说明用户行为、验证结果与限制。
   CI 会检查引用的本仓库 issue 确实存在。维护者确认验收后合并。

```sh
gofmt -w .
go run ./internal/cmd/doccheck
go vet ./...
go test ./...
```

Linux CI 额外使用 race detector；本地有兼容 C 工具链时可使用
`go test -race ./...`。不为本项目修改系统级 Go/Git 配置，不使用真实云凭据运行单元测试。
文档质量检查只覆盖注释与示例的结构，准确性、默认值和限制由评审检查。

没有 GitHub 访问时，先在 `docs/issues/` 放置待同步草稿，联网后用
`gh issue create --repo rambow-cloud/alicloud-go-sdk-x --title "标题" --body-file docs/issues/文件.md`
同步，在打开 PR 前记录真实 issue URL。草稿 ID 不等于 GitHub issue 编号。

公共 API 变更需说明兼容性；协议/生成代码变更需提供固定输入与来源。
任何功能不应只有代码实现而缺少 pkg.go.dev 可读的使用说明。
