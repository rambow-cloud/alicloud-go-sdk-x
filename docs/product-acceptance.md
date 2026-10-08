# Product acceptance / 产品验收标准

[English](#english) | [中文](#中文)

## English

### Objective and authority

Acceptance establishes whether developers can replace the official Alibaba Cloud Go
SDK v2 for an explicit workload with consistent Go APIs, correct behavior and manageable
maintenance. The criteria below govern future scoped issues alongside
[development-path.md](development-path.md). Code generation, compilation, tool selection,
stars and contributor counts do not establish product acceptance.

Keep complete pinned official Darabonba DSL -> official semantic parser -> normalized IR
-> our Go backend -> shared runtime. Smithy remains an isolated local experiment unless
a separate issue/decision changes the route. Require Go 1.27+, direct JSON v2,
standard-library core imports and optional OTel. Issue/PR text and Go comments remain
English-primary; authored Markdown stays English/Chinese equivalent.

### Candidate Beta boundary

The first version is now the user-selected [v0.1.0 STS](sts-v0.1.0.md), tracked
by #57-#61, milestone v0.1.0 and Project 3. Its scoped experimental release gate
requires all four pinned STS actions generated/compiled/offline-tested, credential
composition, scoped live evidence, consumer/maintenance/docs and publication/indexing.
Successful OIDC/SAML live federation is declared outside that required live scope
before execution, never promoted from NOT RUN to PASS. This explicit route revision
supersedes the broader Beta release prerequisite for v0.1.0 only; the candidate
boundary/AC/UX definitions below remain the broader goal. Neither release claims
broader Beta or unverified ECS/VPC coverage. The later user-approved #68 correction
adds required native default configuration/Profile/OAuth behavior before publication;
the earlier exclusion is superseded by [the new route](default-configuration.md).

The candidate scope below is the starting acceptance workload, not a release claim.
Adding/removing required cases needs an issue, evidence and paired updates. Passing
shared contracts does not extend product coverage automatically.

| Boundary          | Candidate requirement                                                                                                                                  |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Products/packages | ECS 2014-05-26, VPC 2016-04-28, STS 2015-04-01 under `service/`                                                                                        |
| Actions           | ECS DescribeRegions/DescribeImages/DescribeInstances/DescribeInstanceStatus; VPC DescribeVpcs; STS AssumeRole                                          |
| Protocol/region   | Reviewed RPC/ACS3 over HTTPS; cn-hangzhou; other modeled regions retain separate evidence                                                              |
| Capabilities      | Four native paginators; ECS InstanceRunningWaiter; reviewed opt-in read retry; STS provider/cache composition; optional OTel                           |
| Identity          | Explicit static/env providers, explicit chain/cache and AssumeRole using separate source credentials                                                   |
| Exclusions        | Automatic process/metadata discovery, arbitrary write retry, ROA/OSS/streaming, whole-cloud parity; native CLI Profile/OAuth is now required under #68 |

Historical local-profile validation injected a checked snapshot after CLI
authentication; it does not establish native loading or renewal. #68 adds native
loading and scoped live exchange/persistence evidence under
[the new contract](default-configuration.md). No cloud resource creation,
write operation or target-role use is authorized by this document.

Application guidance prioritizes renewable STS role providers/cache. Long-lived keys
require explicit StaticProvider/EnvProvider/profile-provider registration; complete
temporary environment credentials and native profiles are discoverable through
LoadDefaultConfig. Custom providers remain supported. Config/Options have no bare
credential fields, construction never makes credential HTTP calls, and nil/typed-nil
sources fail before requests. Present invalid sources stop resolution. Historical
[#53](credentials.md) rejection/signing/cache evidence remains under AC-01/05/11
and UX-01/04; #68 supersedes its blanket discovery exclusion.

### Required criteria

| ID    | Contract                                                                                                                                                              | Required proof                                                                                                                                                        |
| ----- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| AC-01 | Consistent typed Options/NewFromConfig, context-first operations and call options; documented absence/defaults and ownership                                          | External consumer examples; input/config isolation; cancellation/deadline identity; Windows and Linux race CI                                                         |
| AC-02 | Exact Alibaba action/version/wire names, complete supported models, absence/zero/false/int64 and response containers; JSON v2 strictness with unknown-field tolerance | Independent wire fixtures, signed requests, malformed JSON cases and selected CLI/Explorer comparison; report all-field vs selected-field evidence                    |
| AC-03 | Native HasMorePages/NextPage; no invented token; reject mixed modes; stable cursor on errors/cancel; bounded cycle handling                                           | Multi-page token/page, empty page, repeated token, metadata bounds and ownership cases; live continuation evidence for both candidate modes                           |
| AC-04 | Reusable Wait/WaitForOutput; all requested IDs required for success; bounded/context-aware polling and isolated concurrent waits                                      | Missing/partial/duplicate/unknown states, transitions, errors and expiry; authorized live waiter success and transition evidence                                      |
| AC-05 | Explicit credential precedence, missing vs invalid distinction, shared bounded refresh, valid role keys/token/expiration and separate source identity                 | Full-DSL STS client -> provider -> cache -> generated consumer integration; rotation/concurrency/cancel/invalid-response tests; authorized role/refresh live evidence |
| AC-06 | Opt-in retries only for reviewed replayable/idempotent operations; bounded attempts/time, jitter/Retry-After and fresh per-attempt signing                            | Offline fault injection, canceled backoff, budget and unsafe-write rejection; live fault injection is separate and never implied                                      |
| AC-07 | Deterministic product/region endpoints and explicit override; unsupported combinations fail                                                                           | Rule/HTTPS/signing tests and candidate-region live reads; custom and other-region evidence kept separate                                                              |
| AC-08 | Ordered middleware lifecycle and optional injected OTel; operation/attempt hierarchy, propagation and no secret/raw-body labels                                       | Hook lifecycle and test-exporter integration; no global provider mutation; private error/body/query exclusion                                                         |
| AC-09 | errors.Is/As, typed service/operation errors, request ID/status/attempt metadata and safe default formatting                                                          | Error chains, cancellation, status/decode failures and sensitive fmt/log/span tests; no string matching required in examples                                          |
| AC-10 | Small operation/paginator/waiter mock seams, scripted HTTP and virtual time                                                                                           | Real consumer-style tests with no network/account; test business results and failure handling                                                                         |
| AC-11 | English Go docs and runnable offline Examples, equivalent Chinese guidance, defaults/limits/migration/license notices                                                 | doccheck/Examples; source-prose gaps recorded; tagged release and user browser inspection of the same version on pkg.go.dev                                           |
| AC-12 | Complete discovery, deterministic regeneration, explicit unsupported reasons, reviewed sparse policy and safe source updates                                          | Frontend/check/product-check, invalid selected cases leave outputs untouched, compatible/incompatible drift report and one real upstream revision rehearsal           |

For AC-11, local docs/Examples/licenses are Beta requirements; tagging and same-version
pkg.go.dev inspection belong to the release gate. They are not prerequisites for Beta.

AC-03/04/05 use product-specific reviewed rules, not field-name inference. The basic
Smithy allStringEquals waiter and stock paginator ownership observed in the PoC do
not satisfy these contracts automatically. Our policy semantics remain authoritative.

### Developer experience tasks and comparison

| ID    | Consumer task                                                        | Acceptance observation                                                                                                                            |
| ----- | -------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| UX-01 | New module -> explicit credentials -> first ECS read using only docs | Initial target <=15 minutes with Go installed and credentials already authorized; exclude browser/approval wait; record actual time and obstacles |
| UX-02 | Traverse ECS token plus image/VPC page results                       | Zero handwritten cursor advancement/termination logic; caller may use the documented paginator loop                                               |
| UX-03 | Wait for multiple instances                                          | Zero handwritten polling/backoff loop; handle missing IDs, timeout and structured failure                                                         |
| UX-04 | Use renewable AssumeRole credentials                                 | Source STS client, provider, cache and generated consumer compose without a handwritten credential translator or refresh loop                     |
| UX-05 | Test the application and add observability                           | Fake only needed operations; no cloud access in tests; injected tracing and errors.Is/As without SDK internals                                    |

Use Go developers who did not implement the feature. Pin this SDK and official v2
module versions, Go/OS/architecture, API inputs, result projection and workload. Record
task success, time, application-code lines (excluding generated SDK/vendor/setup),
documentation lookups, manual glue and obstacles. Show both solutions; do not infer
superiority from lines alone. No independent-user result has yet been recorded.

[Benchmark #20](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/20) separately
tracks imports/build/binary costs; encoding/decoding latency and allocations require
equal fixtures and published methodology. Set regression thresholds after obtaining
a reproducible baseline, before optimization. No speed or percentage claim is accepted
without measured evidence. Contributor docs, issue/label hygiene, reproduction guidance
and update workflow are maintainer acceptance; community popularity is not a release gate.

### Evidence, gates and current gaps

Every case records criterion/task ID, required scope, SDK/source/official-tool versions,
Go/OS/architecture, commit, check command or browser action, fixture origin, result,
redacted evidence link, limitations and responsible issue. Separate discovered, lowered,
emitted, compiled, offline-tested, live-tested and published/indexed evidence.

Results are PASS, FAIL, SKIP (with reason) or NOT RUN. SKIP/NOT RUN is not PASS. A skipped
required live/UX/publication case keeps its gate open; do not relabel it as optional
after execution. Expected-negative tests pass only when the required rejection is
observed; an audit recording a weakness does not approve that weakness.

| Gate         | Required completion                                                                                                                          |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Foundation   | Eleven shared capabilities, meaningful offline contracts/Examples/doccheck, standard-library core, Linux race and Windows CI                 |
| Product Beta | Candidate AC contracts, all five consumer tasks, scoped live cases and real source-update rehearsal; zero unresolved required failures/skips |
| Release      | Product Beta plus immutable version/release notes, migration/compatibility statement, licenses and same-version pkg.go.dev browser evidence  |

Historical baseline: [foundation mapping](foundation-acceptance.md) and
[integration record](generator-integration.md). The latter records 579 emitted actions,
four paginators, one waiter and seven reviewed policies; 572 actions remain unreviewed
for those capabilities. These are recorded snapshots, not a Beta completion count.
[Live #47 / PR #48](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/48) records
two image pages; instance token/page/status and VPC traversals end on an empty first
page, waiter/AssumeRole are skipped and Explorer browser verification is not run.

STS composition #51 and explicit-provider guidance #53 are merged.
[Live STS #55](live-sts-renewal.md) passes scoped issuance/reuse, forced refresh and
automatic real-time expiry renewal with generated ECS reads and temporary-IAM cleanup.
Background/concurrent live refresh, other roles/conditions and native Profile/OAuth
renewal are not established by this run.

Current work queue, in priority order: reproducible
consumer workloads and independent UX/official-v2 comparison; authorized token/waiter/
remaining scoped role live gaps; real upstream-update rehearsal; release/indexing. Benchmark #20 stays
independent. The older STS helper only accepts `services/sts`; its foundation acceptance
does not establish composition with the new `service/sts` API. #51 adds that composition
with local doccheck, product-check, vet, all Go tests/Examples and 50 frontend tests
passing. #51 is offline evidence; #55 adds scoped live renewal. Independent UX remains open.

Closing an implementation or this definition issue does not close these overall gates.
Review all tracked cases at the target release commit. This change creates no version
tag or switch the production generator; this acceptance definition itself authorizes
no live calls. The separately user-authorized #55 scope is recorded above.

## 中文

### 目标与约束地位

验收判断开发者能否在明确业务范围内替代官方阿里云 Go SDK v2，获得一致 Go 接口、正确行为和
可控维护成本。本标准与[开发路径](development-path.md)共同约束后续 issue。生成、编译、工具
选择、star 或贡献者数量不足以证明产品验收。

保留完整固定官方 Darabonba DSL → 官方语义 parser → 规范化 IR → 本项目 Go 后端 → 公共 runtime。
Smithy 仍是独立本地实验，改变路线需单独 issue/决策。要求 Go 1.27+、直接 JSON v2、核心仅标准库
导入、OTel 可选。Issue/PR 和 Go 注释英文为主，项目 Markdown 中英对应。

### 候选 Beta 边界

用户指定首版现为 [v0.1.0 STS](sts-v0.1.0.md)，由 #57-#61、同名 milestone、Project 3
跟踪。限定实验发布必需四个固定 STS 操作生成/编译/离线测试、凭据组合、限定真实证据、
消费者/维护/文档及发布/索引。OIDC/SAML 真实联邦在执行前声明不属于必需真实范围，
不把 NOT RUN 写成 PASS。此明确路线调整仅替代 v0.1.0 的旧整体 Beta 发布前提，下文
候选边界/AC/UX 保留为更广目标，不宣称整体 Beta 或未经验证的
ECS/VPC 覆盖。

用户后续 #68 要求发布前补默认配置/原生 Profile/OAuth，旧排除范围以
[更新路线](default-configuration.md)为准，不用历史 #55 角色续期冒充原生 OAuth 证据。

以下是首批验收业务，不是发布声明。增减必需项需 issue、证据和双语更新，共享契约通过不自动
扩大产品覆盖。

| 边界      | 候选要求                                                                                                                             |
| --------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| 产品/包   | ECS 2014-05-26、VPC 2016-04-28、STS 2015-04-01，使用 service/                                                                        |
| 操作      | ECS DescribeRegions/DescribeImages/DescribeInstances/DescribeInstanceStatus；VPC DescribeVpcs；STS AssumeRole                        |
| 协议/地域 | 已审核 RPC/ACS3、HTTPS、cn-hangzhou；其他已建模地域证据独立                                                                          |
| 能力      | 四个原生分页器、ECS InstanceRunningWaiter、审核读取的显式重试、STS provider/cache、可选 OTel                                         |
| 身份      | 显式 static/env、有序 chain/cache、来源凭据独立的 AssumeRole                                                                         |
| 范围外    | 自动进程/metadata 发现、任意写重试、ROA/OSS/streaming、全云等价；用户后续 #68 修正要求原生 CLI Profile/OAuth，旧排除范围以新路线为准 |

历史 Profile 验证在 CLI 登录后注入检查过期的凭据快照，不证明原生加载/续期。
#68 按[新契约](default-configuration.md)新增原生加载及真实交换/持久化证据。
本文件不授权云资源创建、写操作或目标角色调用。

### 必需标准

应用指南优先采用可刷新的 STS role provider/cache；长期密钥必须显式注册
StaticProvider/EnvProvider/profile provider，LoadDefaultConfig 可发现完整临时环境凭据及
原生 Profile，仍支持自定义来源。Config/Options 无裸凭据字段，构造期间不进行凭据 HTTP，
请求前拒绝 nil/typed-nil；已配置无效来源停止解析。#53 离线拒绝/签名/缓存证据仍关联
AC-01/05/11、UX-01/04，#68 优先于其全面禁止发现范围。

| ID    | 契约                                                                                             | 必需证据                                                                                      |
| ----- | ------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------- |
| AC-01 | 一致的强类型 Options/NewFromConfig、context 优先和操作 options；明确缺省/默认/所有权             | 外部消费者示例、输入/配置隔离、取消/超时识别、Windows/Linux race CI                           |
| AC-02 | 准确 action/version/线名、完整支持模型、缺省/零/false/int64/响应容器；JSON v2 严格且容忍未知字段 | 独立 wire fixture、签名请求、非法 JSON 和选定 CLI/Explorer 对照；区分全字段/选定字段          |
| AC-03 | 原生 HasMorePages/NextPage，不造 token，拒绝混用，错误/取消不推进，有界循环保护                  | token/页码多页、空页、重复 token、元数据边界、所有权；候选两种模式真实续页                    |
| AC-04 | 可复用 Wait/WaitForOutput，全部请求 ID 达标才成功，有界可取消，并发等待隔离                      | 缺失/部分/重复/未知状态、转换/错误/超时；授权的真实 waiter 成功和状态转换                     |
| AC-05 | 明确凭据顺序、缺失与无效区分、合并有界刷新、角色密钥/token/过期校验、来源身份分离                | 完整 DSL STS→provider→cache→生成消费者；轮换/并发/取消/非法响应；授权角色/刷新真实证据        |
| AC-06 | 仅审核幂等可重放操作显式重试，限制次数/时间，jitter/Retry-After，每次重新签名                    | 离线故障注入、取消退避、预算、非安全写拒绝；真实故障注入独立，不默认宣称                      |
| AC-07 | 产品/地域 endpoint 和覆盖确定，不支持明确失败                                                    | 规则/HTTPS/签名测试、候选地域真实读取；自定义/其他地域证据独立                                |
| AC-08 | 有序 middleware 生命周期，可选注入 OTel，operation/attempt 层级与传播，不含敏感/raw-body 标签    | Hook 生命周期、测试 exporter；不修改全局 provider，不暴露错误/body/query 原文                 |
| AC-09 | errors.Is/As、服务/操作错误、RequestID/status/attempt 元数据、安全格式化                         | 错误链、取消、状态/解码失败、敏感 fmt/log/span；示例不靠字符串匹配                            |
| AC-10 | 小操作/分页/waiter mock 接口、脚本 HTTP、虚拟时间                                                | 真实消费者风格离线测试，无网络/账号，验证业务结果和失败处理                                   |
| AC-11 | 英文 Go docs/离线可运行 Example、对应中文、默认/限制/迁移/许可                                   | doccheck/Examples，报告上游缺说明；带 tag 发布及用户浏览器检查同版本 pkg.go.dev               |
| AC-12 | 完整发现、确定性生成、明确未支持、审核稀疏策略、安全来源更新                                     | frontend/check/product-check，无效选中项写前失败，兼容/破坏漂移报告，一次真实上游版本升级演练 |

AC-11 的本地文档/Example/许可证属于 Beta 要求；tag 及同版本 pkg.go.dev 检查属于发布门槛，
不作为 Beta 的前置条件。

AC-03/04/05 使用审核产品规则，不猜字段。PoC 中基础 allStringEquals waiter 和官方分页器输入
所有权不自动满足这些契约，现有能力策略仍是行为依据。

### 开发者体验与对比

| ID    | 用户任务                                 | 验收观测                                                                               |
| ----- | ---------------------------------------- | -------------------------------------------------------------------------------------- |
| UX-01 | 新 module→显式凭据→仅按文档首次 ECS 读取 | 初始目标不超过 15 分钟，前提 Go 已安装/凭据已授权，不计浏览器/审批等待；记录实测与障碍 |
| UX-02 | 遍历 ECS token 和镜像/VPC 页码           | 不手写游标推进/结束逻辑，可以使用文档中的分页循环                                      |
| UX-03 | 等待多个实例                             | 不手写轮询/退避循环，处理缺失 ID、超时和结构化错误                                     |
| UX-04 | 可刷新的 AssumeRole 身份                 | 来源 STS client/provider/cache/生成消费者直接组合，无手写凭据翻译或刷新循环            |
| UX-05 | 业务测试与可观测                         | 只 fake 所需操作，无云测试，注入 tracing，errors.Is/As，无需 SDK 内部实现              |

邀请未参与实现的 Go 开发者，固定本项目与官方 v2 module 版本、Go/OS/架构、输入、结果投影和
业务。记录成功率、时间、业务代码行（排除 SDK 生成/vendor/环境准备）、文档查询、手写胶水和
障碍，展示双方实现，不单凭代码行判断优劣。目前未记录独立用户结果。

[基准 #20](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/20) 独立跟踪依赖/构建/二进制
成本；编码/解码延迟和分配需相同 fixture 与公开方法。先取得可复现基线，再在优化前定回归门槛。
没有实测证据不接受速度或百分比声明。贡献文档、issue/label、复现指南、更新流程属于维护验收，
社区热度不作为发布门槛。

### 证据、阶段与当前缺口

每项记录标准/任务编号、必需范围、SDK/来源/官方工具版本、Go/OS/架构、提交、命令或浏览器动作、
fixture 来源、结果、脱敏证据链接、限制和负责 issue。发现、降低、输出、编译、离线、真实、发布/
索引分别记录。

结果为 PASS、FAIL、SKIP（需原因）、NOT RUN。跳过/未执行不是通过；必需真实/体验/发布项跳过
时门槛保持未通过，不能执行后改成可选。负例必须观察到规定拒绝才通过；缺陷审计不代表批准缺陷。

| 门槛      | 必需完成                                                                           |
| --------- | ---------------------------------------------------------------------------------- |
| 基础      | 十一项共享能力、有意义离线契约/Example/doccheck、核心标准库、Linux race/Windows CI |
| 产品 Beta | 候选 AC 契约、五项用户任务、限定真实用例、真实来源升级；必需项无未解决失败/跳过    |
| 发布      | Beta 加不可变版本/发布说明、迁移兼容声明、许可证、同版本 pkg.go.dev 浏览器证据     |

历史依据：[基础映射](foundation-acceptance.md)、[集成记录](generator-integration.md)。后者记录
579 个输出操作、4 分页器、1 waiter、7 个审核策略，572 个操作相关能力未审核；这些是历史快照，
不是 Beta 完成数。[真实 #47 / PR #48](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/48)
记录镜像两页，实例 token/页码/状态及 VPC 均空结果第一页结束，waiter/AssumeRole 跳过，Explorer
浏览器未执行。

STS 组合 #51 和显式 provider 指南 #53 已合并。[真实 STS #55](live-sts-renewal.md)通过限定
签发/复用、强制刷新、真实时间到期后自动续期、生成 ECS 读取和临时 IAM 清理；不证明真实
后台/并发刷新、其他角色/条件或原生 Profile/OAuth 续期。

按优先级继续：可复现消费者业务与独立用户/官方 v2 对比；授权 token/waiter 和剩余限定 role
真实缺口；真实上游升级；发布/索引。基准 #20 独立。旧 STS helper 只接受
services/sts，基础通过不代表与新 service/sts 可组合。#51 增加该组合，本地 doccheck、
product-check、vet、全 Go 测试/Example、50 项前端测试通过；#51 为离线证据，#55 增加限定真实
续期，独立体验仍待验证。

关闭实现或本定义 issue 不代表这些总门槛通过；在目标发布提交复核所有用例。本次不打 tag、
不切换生产生成器；本验收定义本身不授权真实调用，用户独立授权的 #55 范围已在上文记录。
