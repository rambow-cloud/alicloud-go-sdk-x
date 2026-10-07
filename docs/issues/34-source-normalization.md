# Issue 34: source normalization / Issue 34：来源规范化

## English

Actual issue: [#34](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/34), stage
of parent [#33](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/33). Depends
on #31 and its unmerged PR #32; branch issue/34-source-normalization is stacked on
issue/31-official-darabonba-frontend. Scope, commands and evidence limits are in
[source-normalization.md](../source-normalization.md). The authoritative route was
committed before implementation at a206ea0, with actual issue links at 92095fe.

Delivery includes pinned Apache-2.0 canonical fixtures, the versioned Node adapter,
exact CLI/wire case separation, recursive itemName wrappers, verified Filter aliases
in Node/Go and an explicit optional-string exception for metadata-only Tag[].key/value.
New discrepancies and source tampering fail before writes; no generated Go API or
runtime changes. Regression tests include signed-wire/public contract tests already
in the repository and cross-backend generated Go byte comparison.

Acceptance uses frontend offline check/tests, sdkgen check, doccheck, vet, Go tests,
formatting and paired-document checks; record actual results in the linked PR.
Linux race/Windows CI remain required; local Windows checks do not replace them.
Browser/live behavior is unverified for this stage. #35 owns full operation/model
discovery; #36 owns batch emission. Parent #33 stays open for remaining stages.

## 中文

真实 issue 为 [#34](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/34)，父任务
为 [#33](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/33)。依赖 #31 及未合并
PR #32；issue/34-source-normalization 分支叠加在 issue/31-official-darabonba-frontend。
范围、命令、证据边界见[来源规范化](../source-normalization.md)。权威路线在实现前
提交 a206ea0，真实 issue 链接提交为 92095fe。

交付包括固定 Apache-2.0 canonical 样本、版本化 Node 适配器、准确 CLI/线大小写
分离、递归 itemName 包装、Node/Go 的 Filter 绑定核验，以及 metadata-only
Tag[].key/value 的可选字符串例外。新偏差和来源修改在写前失败；生成 Go API 和
runtime 不变。回归包括仓库既有签名/公开契约测试、跨后端 Go 产物字节比较。

验收使用前端离线检查/测试、sdkgen check、doccheck、vet、Go 测试、格式和双语
检查，实际结果记入关联 PR。Linux race/Windows CI 仍必须通过，本地 Windows 检查
不代替它们。本阶段未核验浏览器/真实行为；全量操作/模型发现属于 #35、批量输出
属于 #36。父任务 #33 保持打开以跟踪后续阶段。
