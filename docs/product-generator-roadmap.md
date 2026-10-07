# Product generator roadmap / 产品生成器路线

## English

Accepted direction, 2026-10-07. This roadmap, development-path.md and AGENTS.md supersede
conflicting older generation plans. Current #31 is a bounded compatibility bridge;
product discovery/emission must not require per-operation handcrafted field/model/doc
selection. Historical acceptance records remain accurate, separate from future goals.

The source pipeline is complete official DSL -> official semantic parser -> normalized
operation/model/binding IR -> our Go backend -> accepted runtime. Optional pinned
canonical metadata enriches API requiredness, parameter styles, documentation and
cross-checks. It cannot overwrite exact DSL wire names or dictate CLI response shapes.
Preserve provenance for every fact; missing enrichment is recorded rather than invented.

Sources: [SDK DSL](https://github.com/aliyun/alibabacloud-sdk) and
[CLI metadata](https://github.com/aliyun/aliyun-openapi-meta). The latter declares Apache-2.0
and warns its structure is unstable and currently intended for CLI builds. Pin commits
and hashes and use versioned adapters. Licensed descriptions may seed English Go comments
and equivalent Chinese guides with attribution/license preserved. Web metadata and imported
modules have separate provenance; do not label the entire corpus MIT.

| Stage | Delivery | Acceptance |
| --- | --- | --- |
| 1: normalization | Exact wire names/case; indexed request models; itemName response wrappers; raw provenance and CLI-only attributes | Filter model and Filter.1.Key/Value correspond; true name/type/requiredness changes remain visible |
| 2: discovery/IR | Every product operation and reachable model; protocol/bindings/source locations; versioned deterministic IR and coverage | No per-operation metadata/overlay dependency; discovered/lowered/unsupported counts and reasons |
| 3: batch emission | Complete supported inputs/outputs/models/methods, small interfaces, generic codecs/naming with sparse compatibility exceptions | Compiles; signed-wire/copy/presence tests; deterministic output; selected unsupported behavior fails before writes |
| 4: capability policy | Native token/page pagination, waiter acceptors, idempotency/client tokens, sensitive fields and special validators | Shared engines and AWS conventions; conservative defaults, no invented pagination or guessed write retry |
| 5: docs/expansion | Licensed bilingual doc automation, English Go comments, corresponding guides and offline Examples; further profiles/products | pkg.go.dev, Linux race/Windows, honest coverage and separate live evidence |

Stage 1 retains raw_name separately from CLI name/options, case-sensitive Key/key and
repeatList element fields. Indexes are declared positions, not inferred server limits.
Response array plus itemName is a representation requiring an object/array wrapper;
backendName, nullToEmpty and valueMapping remain source attributes, not automatic SDK
transformations. Only classify material source conflicts after normalization.

Stage 2 scans complete pinned DSL independently of legacy selected snapshots/overlays.
Lowered does not mean generated, compiled or live accepted. Enumerate unsupported ROA,
body/stream/helper patterns with source locations and reasons. An explicit supported
subset is reviewable; never silently omit APIs while claiming a complete SDK.

Stage 3 preserves context, service Options/NewFromConfig, functional call options,
wire containers, absence, input copying and cancellation/error behavior. RPC comes
first. Stage 4 overlays describe exceptions/policies rather than every wire field;
do not infer retry safety just from names, CLI operation_type or HTTP verbs.

Every stage includes its own docs/tests/Examples and offline regeneration; stage 5
expands automation rather than postponing docs. Do not turn account/resource CLI examples
into generated Go tests. Browser checks are performed by the user at exact documented
links; authorized read-only CLI/HTTP evidence is distinct from offline and UI evidence.

First milestone: normalize real representations and publish complete pinned ECS
inventory/coverage, then batch generate supported RPC operations. Track stages under an
open roadmap parent with real child issues created before code and dependencies
1 -> 2 -> 3 -> 4 -> 5. Record returned issue numbers in docs/issues/README.md.
Created tracking: parent [#33](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/33),
stages #34 (normalization) -> #35 (discovery/IR) -> #36 (emission) -> #37 (policy) ->
#38 (docs). These are planned deliveries, not a claim that all stages are implemented.
Keep #31 focused on the compatibility bridge and benchmarks separate. Each issue uses
its own branch; stacked PRs name dependencies and do not imply main contains unmerged work.
Do not close the parent while batch emission, capability policies or docs remain unfinished.

## 中文

2026-10-07 确认的新方向。本路线、development-path.md、AGENTS.md 优先于冲突旧计划。
#31 是有界兼容桥；产品发现/输出不能要求每操作手写字段、模型、说明选择。历史
验收记录保持准确，与未来目标分开。

流水线为完整官方 DSL → 官方语义 parser → 规范化操作/模型/绑定 IR → 本项目 Go
后端 → 已验收 runtime。固定 canonical 元数据可补充 API 必填、参数风格、文档及
交叉验证，不能覆盖准确 DSL 线名或强加 CLI 响应形状。事实保留来源，缺数据明确
记录，不发明信息。

来源为[官方 SDK DSL](https://github.com/aliyun/alibabacloud-sdk)和
[CLI 元数据](https://github.com/aliyun/aliyun-openapi-meta)。后者声明 Apache-2.0，同时
提示格式不稳定、当前用于 CLI 构建。固定 commit/哈希并用版本化 adapter；授权
说明可作为英文 Go 注释和对应中文指南输入，保留归属/许可。网页元数据与导入模块
单独记录来源，不将整体源码标为 MIT。

| 阶段 | 交付 | 验收 |
| --- | --- | --- |
| 1：规范化 | 准确线名/大小写、索引请求模型、itemName 响应包装、原始来源与 CLI 属性 | Filter 模型与 Filter.1.Key/Value 对应，真实名称/类型/必填变化可见 |
| 2：发现/IR | 产品全部操作与可达模型、协议/绑定/来源位置、版本化确定性 IR/覆盖 | 无逐操作元数据/overlay 前提，分别统计发现/降低/不支持及原因 |
| 3：批量输出 | 完整支持输入/输出/模型/方法、小接口、通用编码/命名及少量兼容例外 | 编译、签名/复制/存在测试、确定性；选中不支持行为写前失败 |
| 4：能力策略 | 原生 token/页码分页、waiter、幂等/client token、敏感字段及特殊校验 | 复用引擎/AWS 范式，保守默认，不造分页、不猜写重试 |
| 5：文档/扩展 | 授权双语文档自动化、英文 Go 注释、对应指南/离线 Example、更多协议产品 | pkg.go.dev、Linux race/Windows、准确覆盖及单独真实证据 |

阶段 1 保留 raw_name 和 CLI name/options 的区别、Key/key 大小写及 repeatList element
字段；索引是声明位置，不推断服务上限。array+itemName 表示需恢复对象/数组包装，
backendName/nullToEmpty/valueMapping 作为来源属性，不自动用于 SDK 转换；规范化
后再判断实质冲突。

阶段 2 独立扫描完整固定 DSL，不依赖旧快照/overlay 列表；降低不等于生成、编译或
真实通过。ROA/body/流/helper 不支持模式列来源位置/原因；支持子集明确可审，不
静默漏 API 并宣称完整 SDK。

阶段 3 保留 context、Options/NewFromConfig、functional options、线容器、缺失、
复制、取消/错误；先 RPC。阶段 4 overlay 只记录例外/策略，不逐项重述字段；不
仅凭命名、CLI operation_type、HTTP verb 推断重试安全。

每阶段同步文档/测试/Example/离线再生成，阶段 5 扩展自动化，不推迟文档。不把
含账号/资源的 CLI 示例变成 Go 测试。浏览器由用户按准确链接核验，授权只读 CLI/HTTP
与离线/UI 证据分开。

首个里程碑是规范化真实表示并交付完整固定 ECS inventory/覆盖，再批量输出支持
RPC。以保持打开的路线父任务及代码前创建的真实子 issue 跟踪 1 → 2 → 3 → 4 → 5，
返回编号写入索引；#31 聚焦兼容桥，基准独立。每项独立分支，堆叠 PR 明确依赖，
不宣称未合并代码已在 main；批量输出、能力策略、文档未验收前不关闭父任务。
已建立父任务 [#33](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/33)，阶段为
#34（规范化）→ #35（发现/IR）→ #36（输出）→ #37（策略）→ #38（文档）。这是计划交付，
不表示全部阶段已实现。
