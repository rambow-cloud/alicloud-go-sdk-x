# Issue 维护

[English](issue-management.md)

- 参考 rambow-cloud/powertools-lambda-go 的类型/module 分类，并增加明确的 priority/status。
- .github/labels.json 为受管目录；自动化更新目录标签，不删除目录外手工标签。
- module 表示职责而非独立 go.mod；新增领域需同步目录、表单和分类器。

- 类型为 bug、enhancement、documentation、maintenance、question、release；cicd/help wanted/good first issue 按实际需要补充。
- 只保留一个优先级：p1 基础/当前里程碑，p2 后续计划，p3 未承诺工作。
- 状态只保留一个：triage、ready、in-progress、blocked、done；阻塞原因写入正文。

- issue 标题、正文和表单只使用英文；中文说明放在独立的 `.zh-CN.md` 指南中。表单保留 Affected areas 字段。
- 自动化根据标题前缀和结构字段分类，保留手工优先级/活动状态， 关闭设 done、重开设 triage。
- done 是验收交付，不是发布或索引。
- 流程处理 issue 事件与手动 backfill，不发布评论。

- 翻译历史 issues 时保留已完成证据。
- 完整工作使用 Closes #N，部分用 Refs #N；PR 门禁核验真实本仓库 issue， 忽略注释模板、代码示例和 PR 编号。
- 按开发路径及 issue 索引执行，十一项基础验收前生成器保持阻塞。

- issue 工作流拒绝标题或正文中的中文说明；代码示例中的原始 API 值可以保留。
- 修改语言时保留 issue 状态、标签、里程碑、关联链接和验收证据。
