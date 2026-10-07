# Issue 与 label 维护

参考 [rambow-cloud/powertools-lambda-go](https://github.com/rambow-cloud/powertools-lambda-go)
的分类方法：类型标签、module 标签、结构化表单、自动同步目录、PR 关联 issue。
本项目另加 priority 与 status。实现按本项目单一 module 的范围设计；未复制上游脚本。

`.github/labels.json` 是本项目标签目录。自动化创建或更新目录中的标签，不删除
目录外手工标签。新增职责分组时同步目录、表单选项、分类脚本和此说明。

| 维度 | 规则 |
| --- | --- |
| 类型 | bug / enhancement / documentation / maintenance / question / release 至少一种主要分类 |
| 领域 | module:core / credentials / transport / signing / ecs / tools，可多选；表示职责而非独立 go.mod |
| 补充 | cicd、help wanted、good first issue 按实际需要添加，不自动标记初学者任务 |
| 优先级 | priority:p1 当前里程碑必需；p2 后续计划；p3 无近期交付承诺；维护者保留一个 |
| 状态 | status:triage 待评估；ready 前置满足；in-progress 正在实现/评审；blocked 有明确前置；done 验收已交付；维护者保留一个 |

Issue 表单设置初始类型与 triage。自动化从 `[Bug]:`、`[Feature]:`、`[Docs]:`、
`[Maintenance]:` 等标题和 `### Affected areas` 字段同步类型与领域，不通过正文关键词猜测。
无明确优先级时默认 p2；编辑不覆盖维护者设定的优先级与活动状态。缺少领域字段时保留手工领域。
关闭设 done，重开设 triage。done 代表验收交付，不代表已经发布或 pkg.go.dev 已索引。

维护者评估证据、范围、前置条件后，将 triage 改为 ready 或 blocked。开始工作改为
in-progress；前置完成时检查 blocked issue，不因依赖被关闭而自动宣称其已 ready。
PR 合并前核对验收，关联完整交付写 `Closes #N`。状态变动时移除旧 status，避免多个状态。

自动化在 issue 打开/编辑/关闭/重开、目录变更和手动 workflow_dispatch 时执行。
手动运行 Issue labels 会同步标签目录并对所有现有 issues 做 backfill。
流程不发邮件或通知评论，不执行 issue 中的命令，也不改动人工优先级。

初始 issues #1–#9 对应 `docs/issues/` 的同步快照；GitHub 是执行状态的来源。
实现前同时维护正文中的前置依赖与验收条件，不只挂标签。
