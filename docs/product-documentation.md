# Licensed product documentation / 授权产品文档

## English

Stage #38 follows #37 / PR #42 on `issue/38-licensed-product-docs`, stacked on
`issue/37-capability-policies`. The original dependency stack is now
[integrated into main](generator-integration.md). This specification preceded implementation.
Keep the existing supported RPC
profile and operation/model counts; broader protocols require separate scoped work.

Use the official semantic parser's field description strings and operation annotation
tokens from checksum-pinned product DSL. Extend IR with description/summary records
and source coordinates; do not parse Tea source using a second compiler or require
per-operation documentation overlays. Upstream example values remain coordinates only:
never copy account IDs, credentials, resource IDs, CLI commands or policy examples into
executable Go Examples. Existing invocation/capability Examples stay deterministic,
offline and explicitly illustrative, rather than valid cloud request sets.

Emit English-primary native Go paragraphs on operations and fields, retaining our
ownership, pointer, retry and cancellation contracts. Convert Markdown links/emphasis,
headings and lists into readable Go prose; preserve safe HTTPS links, discard HTML/
unsafe URL markup, normalize controls/newlines, and escape compiler directives. Every
source line is emitted inside // comments; upstream prose cannot create code or become
validation, requiredness, deprecation or retry policy. Mark descriptions as upstream
service documentation distinct from SDK contracts and link to pinned source lines.
Missing/empty/non-English-only descriptions are reported, never invented or translated.

Generate paired English/Chinese guides with the same usage/contracts, operation/source
indexes and documentation coverage. Semantic prose is available in English in the
fixed corpus; no pinned Chinese translation is assumed. Both language guides point to
the same licensed source/Go documentation and state this language limitation. Do not
label English excerpts as translated Chinese. Retain machine-readable operation/field
documentation coverage, missing descriptions and source attribution independently of
emission, reviewed capability and live acceptance.

Preserve upstream source/license bytes, attach copyright, Apache-2.0 attribution,
prominent transformation notices and full license terms to product output. Root MIT
continues to cover original runtime/tooling; upstream-derived definitions/prose are
not relabeled MIT. No imported module implementation is emitted. Package notices and
source-lock references make redistributed generated packages self-describing.

Acceptance: official parser extraction with real corpus and missing/empty annotations;
deterministic output, source coordinates, English prose, safe Markdown/HTML/directive
handling, no promoted source examples or behavioral changes, preserved notices and
accurate language/description counts. Existing protected-output/reconciliation tests
also cover generated docs/licenses. Run frontend/discovery checks and relevant Node
tests before Go gates; both generation checks, full Go tests including isolated product
compilation/Examples, doccheck, vet, tracked formatting, bilingual/whitespace checks,
then Linux race/Windows CI. No live calls, release, pkg.go.dev indexing or merges.

Evidence: pinned [product corpus](https://github.com/aliyun/alibabacloud-sdk/tree/ec489e5c3deae95496daae2b41503ac58b221adb),
[upstream license](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/LICENSE),
and [Apache-2.0 terms](https://www.apache.org/licenses/LICENSE-2.0.txt), inspected 2026-10-07.

## 中文

阶段 #38 继 #37 / PR #42，在 `issue/38-licensed-product-docs` 叠加
`issue/37-capability-policies`；原依赖链现已[集成 main](generator-integration.md)。本规格
先于实现。保持当前 RPC 支持范围及操作/模型数量，更多协议另行明确范围。

复用官方语义 parser 的字段 description 字符串及操作 annotation token，从哈希固定
的产品 DSL 提取；IR 增加说明/摘要及来源坐标，不另写 Tea 编译器，不要求逐操作文档
overlay。上游 example 值仍仅保留坐标，不将账号、凭据、资源、CLI 命令或权限策略
复制到可执行 Go Example；既有示例保持确定、离线且明确只演示调用，不冒充有效云参数。

为操作/字段输出英文为主的原生 Go 段落，保留本 SDK 的所有权、指针、重试、取消
契约。将 Markdown 链接/强调、标题、列表转成可读 Go 说明，保留安全 HTTPS 链接，
去掉 HTML/不安全 URL 标记，规范控制字符/换行并阻止编译器指令。每行均为 // 注释；
原文不能生成代码，也不自动变成校验、必填、弃用或重试策略。明确区分上游服务说明
与 SDK 契约，并链接固定来源行。缺失/空/纯非英文说明明确记录，不编造或自动翻译。

生成对应中英文指南，同步使用/行为契约、操作/来源索引及文档覆盖。固定语料的
语义说明为英文，不假设有固定中文翻译；两个语言指南指向同一授权来源/Go 文档，
明确此语言限制，不把英文摘录标成已翻译中文。机器报告单独记录操作/字段说明覆盖、
缺失及归属，不将其混同于代码输出、能力审核或真实验收。

保留上游源码/许可原始字节；产品输出附版权、Apache-2.0 归属、明确转换声明及
完整许可。根 MIT 仍适用于原创 runtime/工具，上游派生定义/说明不改标 MIT，不
输出导入模块实现。包通知及来源锁引用让分发后的生成包自身包含来源/许可信息。

验收覆盖官方 parser 提取、真实语料/缺失/空注释、确定性、来源坐标、英文说明、
Markdown/HTML/指令安全转换、不复制原始示例或改变运行行为、许可/语言覆盖准确。
保护文件/产物清理测试同步覆盖文档/许可。先运行前端/发现检查及相关 Node 测试，
再运行新旧生成检查、完整 Go 测试（含隔离编译/Example）、doccheck、vet、已跟踪
格式、双语/空白检查及 Linux race/Windows CI。不做真实调用、发布、索引或合并。
证据对应英文章节三处固定产品来源、上游许可及完整 Apache 条款，2026-10-07 阅读。
