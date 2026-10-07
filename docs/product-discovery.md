# Product discovery and IR / 产品发现与 IR

## English

Stage [#35](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/35) follows the
[authoritative roadmap](product-generator-roadmap.md) and #34 normalization. Branch
issue/35-product-discovery is stacked on issue/34-source-normalization / PR #39,
which depends on unmerged #32. This specification is committed before implementation.

The build-time command uses Node 22, official parser 2.2.1 and the complete pinned
product/import corpus. It verifies source hashes and import resolution, then performs
semantic parsing offline. It never reads legacy metadata manifests, decisions,
snapshots or overlays; canonical enrichment is optional and not needed for discovery.
No Go runtime dependency, generated service API or automatic retry policy is added.

Operation candidates come from DSL WithOptions functions, API declarations and
functions constructing OpenApi.Params. Constant action names retain their exact case;
ambiguous/dynamic actions remain unsupported with source evidence. Optional pinned
api-info.json is a coverage cross-check: catalog-only and DSL-only entries remain
visible rather than limiting discovery to a catalog whitelist. Ordinary forwarding
functions are not counted as separate APIs. Unexpected candidate behavior is recorded;
source/parser failures and malformed inventory are fatal before output writes.

The versioned product IR records provenance, operation declaration locations, protocol
constants, request/response roots, reachable named/anonymous models and exact wire
members. Model references retain identity instead of recursively duplicating models.
Original numeric type names, optionality, attributes and source coordinates are
preserved; descriptions/examples point to the licensed original source rather than
being reinterpreted as validators. RuntimeOptions/imported transport models remain
explicit external references. Unreachable product declarations are accounted for.

Coverage distinguishes discovered, lowered and unsupported operations. Lowering is
limited to the existing reviewed HTTPS POST RPC direct-query/json-body profile, with
every wire input/response shape checked for unsupported constructs. ROA, body/stream,
helper transforms, unresolved/recursive/inherited wire models or other unsupported
patterns have stable reason codes, messages and source locations. A lowered record is
not a claim of Go emission, compilation, runtime policy or live-cloud acceptance.
Those later acceptance stages remain unassessed here.

Planned repository-root commands:

```sh
node tools/darabonba/discovery.cjs generate
node tools/darabonba/discovery.cjs report ecs
node tools/darabonba/discovery.cjs check
```

Artifacts live under models/{ecs,sts,vpc}/ir.json and coverage.json with a hash-pinned
models/manifest.json. All products preflight before writes; check never writes or
fetches. System write failures can leave partial updates. Optional --operations
selection is a strict acceptance gate: selected unknown/unsupported operations fail
before writes, while unselected unsupported entries remain in the full inventory.
The report prints actual counts/reasons and locations; no invented API coverage.

Acceptance: real full-corpus deterministic regeneration and complete accounting;
discovery with no legacy metadata/overlays/canonical files; malformed/unsupported
protocol, binding, model and source negative tests; strict selection before writes;
paired docs and CLI usage examples. Run frontend/discovery checks and tests, sdkgen
check, doccheck, vet, Go tests, formatting and language checks. CI runs Linux race
and Windows Go 1.27. Public Go packages continue direct JSON v2 and existing offline
Examples. Browser/live calls are not part of this stage; full product Go emission is
#36, capability policies #37 and documentation automation #38.

## 中文

阶段 [#35](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/35) 按
[权威路线](product-generator-roadmap.md) 继 #34 规范化执行。
issue/35-product-discovery 分支叠加在 issue/34-source-normalization / PR #39，后者
依赖未合并 #32。本规格在实现前提交。

构建命令使用 Node 22、官方 parser 2.2.1、完整固定产品/导入源码；先核验来源哈希
和导入路径，再离线语义解析。不读取旧 metadata manifest、决策、快照或 overlay，
发现不需要 canonical 补充。不增加 Go runtime 依赖、生成服务 API 或自动重试策略。

操作候选来自 DSL WithOptions 函数、API 声明及构造 OpenApi.Params 的函数。常量
action 保留准确大小写，歧义/动态 action 带来源证据记为不支持。可选固定 api-info.json
用于覆盖交叉核验，仅 catalog 或仅 DSL 条目均可见，不把 catalog 当白名单。普通
转发函数不单独算 API。不认识的候选行为记入报告；来源/parser 失败、不合法清单
在写输出前整体失败。

版本化产品 IR 记录来源、操作声明位置、协议常量、请求/响应根、可达具名/匿名模型
及准确线成员。引用保留模型身份，不递归复制模型；保留原始数值类型、可选性、属性
和坐标；description/example 指向带许可证的原始来源，不变成 validator。
RuntimeOptions/导入 transport 模型明确为外部引用，不可达声明也有统计。

覆盖区分已发现、已降低、不支持。降低仅接受既有 HTTPS POST RPC 直接 query/json
响应模式，核验全部线输入/响应结构。ROA、body/stream、helper 转换、未解析/递归/
继承线模型及其他不支持行为有稳定原因代码、说明、源码位置。已降低不表示已输出
Go、编译、建立 runtime 策略或真实云验收；后续验收在此均尚未评估。

计划在仓库根目录执行英文章节的 generate、report ecs 和 check 命令。产物为
models/{ecs,sts,vpc}/ir.json、coverage.json 及哈希固定的 models/manifest.json。
全部产品检查后写入，check 不写、不联网；系统写入失败可能部分更新。可选
--operations 提供严格验收：选中未知/不支持操作在写前失败，未选中的不支持操作
仍保留全量清单。报告打印实际数量、原因、位置，不编造覆盖。

验收包含真实全量来源的确定性再生成和完整统计、没有旧元数据/overlay/canonical
仍可发现、协议/绑定/模型/来源异常、严格选择写前失败、双语文档和 CLI 使用示例。
执行前端/发现检查与测试、sdkgen check、doccheck、vet、Go 测试、格式及语言检查；
CI 使用 Linux race、Windows Go 1.27。公共 Go 包继续直接 JSON v2、既有离线 Example。
本阶段不做浏览器/真实调用；全产品 Go 输出属于 #36、能力策略 #37、文档自动化 #38。
