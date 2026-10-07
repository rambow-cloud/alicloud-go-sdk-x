# Issue management / Issue 维护

[English](#english) | [中文](#中文)

## English

Follow the type/module classification of rambow-cloud/powertools-lambda-go, with explicit
priority and status labels. .github/labels.json is the managed catalog; automation updates
catalog labels without deleting outside manual labels. module labels denote responsibilities,
not separate go.mod files. Keep new areas synchronized in catalog, forms and classifier.

Types: bug, enhancement, documentation, maintenance, question, release. Supplement with
cicd, help wanted or good first issue only when justified. Keep exactly one priority:
p1 foundation/current milestone, p2 planned follow-up, p3 uncommitted future work.
Keep one status: triage, ready, in-progress, blocked, done. Record blockers in issue bodies.

English-primary issue forms include Affected areas. Automation classifies title prefixes
and this structured field, preserves manual priority/active status, sets done on closure,
and triage on reopening. Done means delivered acceptance, not release/indexing.
The workflow handles issue events and manual catalog backfill without posting comments.

Translate legacy issues without losing historical completion evidence. Complete work uses
Closes #N, partial work Refs #N; the PR gate checks actual repository issues, excluding
comment templates, fenced examples and PR numbers. Follow development-path.md and the
issue index. Keep generator blocked until the eleven-capability foundation passes.

## 中文

参考 rambow-cloud/powertools-lambda-go 的类型/module 分类，并增加明确的 priority/status。
.github/labels.json 为受管目录；自动化更新目录标签，不删除目录外手工标签。
module 表示职责而非独立 go.mod；新增领域需同步目录、表单和分类器。

类型为 bug、enhancement、documentation、maintenance、question、release；cicd/help wanted/good first issue
按实际需要补充。只保留一个优先级：p1 基础/当前里程碑，p2 后续计划，p3 未承诺工作。
状态只保留一个：triage、ready、in-progress、blocked、done；阻塞原因写入正文。

英文为主的表单保留 Affected areas。自动化根据标题前缀和结构字段分类，保留手工优先级/活动状态，
关闭设 done、重开设 triage。done 是验收交付，不是发布或索引。流程处理 issue 事件与手动 backfill，不发布评论。

翻译历史 issues 时保留已完成证据。完整工作使用 Closes #N，部分用 Refs #N；PR 门禁核验真实本仓库 issue，
忽略注释模板、代码示例和 PR 编号。按开发路径及 issue 索引执行，十一项基础验收前 generator 保持阻塞。
