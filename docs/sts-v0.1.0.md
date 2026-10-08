# v0.1.0 STS delivery / v0.1.0 STS 交付

[English](#english) | [中文](#中文)

## English

### Authority and scope

The user-approved 2026-10-08 first release target is **v0.1.0: STS through the
complete Darabonba pipeline**. This scoped experimental release takes priority over
the earlier ECS/VPC/STS candidate Beta schedule, while preserving its broader
[acceptance criteria](product-acceptance.md) and historical evidence. Completing
this milestone does not establish that broader Beta. No release tag or new cloud
resources are created by this planning change.

Use pinned official STS 2015-04-01 DSL/imports, the official semantic parser,
normalized complete IR, our Go backend and the shared runtime. Ship all four actions
in that pinned source under `service/sts`: AssumeRole, GetCallerIdentity,
AssumeRoleWithOIDC and AssumeRoleWithSAML. Reconcile the source inventory during
an update rather than claiming perpetual or whole-cloud coverage. Keep Go 1.27,
direct encoding/json/v2, AWS-style calling conventions, explicit credential
providers, source ownership, safe errors and licensed bilingual documentation.

### Baseline and remaining work

The planning-time baseline discovered four actions and emitted only AssumeRole. OIDC/SAML
report DSL_PROTOCOL_PROFILE (Anonymous RPC and doRPCRequest); GetCallerIdentity
reports DSL_OPERATION_SIGNATURE (no request-model argument). Fix the underlying
frontend/IR/backend/runtime contracts under separate issues. Do not handwrite those
operations, silently sign anonymous requests, introduce implicit key discovery or
edit generated outputs. Anonymous authentication is an explicit per-operation
protocol contract, not permission to omit credentials from signed operations.

Accepted baseline: #51 full-DSL provider/cache composition, #53 STS-first guidance
and explicit source registration, and #55 real 900-second expiry renewal with
three STS issuances/four generated ECS reads and complete temporary-IAM cleanup.
Those closed issues remain delivered evidence; do not recreate their cloud setup
or repeat the same live run merely to populate a new milestone.

### Delivery order and evidence

1. Support requestless RPC signatures and generate GetCallerIdentity, preserving
   operation names, context, Options, mock seam, exact complete responses and metadata.
2. Support reviewed Anonymous RPC/doRPCRequest lowering and generate OIDC/SAML;
   preserve exact fields and omission, avoid signing/provider retrieval and redact
   request tokens/assertions as well as response credentials in default formatting.
   Items 1 and 2 can proceed independently on separate issue branches.
3. After both, accept the four-operation generated product: reproducible offline
   wire/response/negative contracts, provider/cache regression, a pinned external
   consumer and official-v2 comparison, authorized GetCallerIdentity live evidence,
   and a real source-update rehearsal with reviewed drift. Existing #55 remains
   the scoped live AssumeRole/renewal evidence unless a relevant change invalidates it.
4. Prepare compatibility/migration notes, licenses, English/Chinese release notes
   and the versioned pkg.go.dev inspection. Publish only after required prepublication
   gates; close the milestone only after publication/indexing evidence is recorded.

| Gate                  | Required evidence                                                                                                                                                          |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Generation            | Four discovered/lowered/emitted actions; complete reachable models; deterministic frontend/IR/Go checks; unsupported selected behavior fails before writes                 |
| Behavior              | All four compile and have independent offline protocol/field/response/error/cancellation tests and external Examples; signed/anonymous separation and sensitive formatting |
| Credentials           | Explicit providers; generated AssumeRole -> provider/cache -> consumer; existing ownership/concurrency/cancellation contracts and scoped #55 live renewal                  |
| Consumer/maintenance  | Version-pinned external STS workloads and official-v2 examples; docs-only independent Go-user task evidence; reviewed real upstream revision rehearsal                     |
| Live scope            | AssumeRole/renewal #55 plus authorized GetCallerIdentity; SDK/CLI and Explorer browser evidence recorded separately                                                        |
| Documentation/release | English pkg.go.dev comments, paired guides, licenses, Linux race/Windows CI, immutable v0.1.0 tag/release and same-version pkg.go.dev browser evidence                     |

OIDC/SAML successful live federation requires separately authorized IdP/provider/
assertion fixtures. Those live cases are **outside this first release's required
live scope**, declared before execution; mark them NOT RUN and disclose the limit
in the release notes. All four actions' offline correctness remains mandatory.
The user's subsequent #68 correction makes native default configuration/Profile/OAuth
renewal required before publication; follow [the updated route](default-configuration.md).
Federation credential-provider helpers remain separate future scope. STS has no pagination/waiter workload here; preserve shared engines
without inventing STS pagination. ECS is only the existing role-credential consumer;
ECS/VPC product-wide acceptance and benchmark #20 remain independent.

### GitHub coordination

Use a GitHub milestone named `v0.1.0` and a repository-linked Projects v2 project
named `alicloud-go-sdk-x: v0.1.0 STS`. Milestone membership identifies release scope;
issues contain acceptance/dependencies/evidence; Project Status tracks execution.
Keep one status label per issue: ready prerequisites -> status:ready/Todo;
active work -> status:in-progress/In Progress; unmet issue dependencies ->
status:blocked/Todo with named prerequisites; delivered acceptance -> status:done/Done.
The Project is maintained when issue state changes; this change does not claim
automatic synchronization. Never mark release Done because only code is merged.

The parent tracks child completion and remains open through the release gate.
Children refer to the parent as membership, not as a blocking prerequisite, keeping
the dependency graph acyclic. Prior closed STS evidence joins the milestone/Project;
historical generation/foundation issues and unrelated projects retain their scope.
No arbitrary deadline is imposed. Read this document before choosing the next
v0.1.0 issue.

[Milestone v0.1.0](https://github.com/rambow-cloud/alicloud-go-sdk-x/milestone/4) |
[Project](https://github.com/orgs/rambow-cloud/projects/3) |
[Parent #57](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/57)

| Issue | Work                                                                       | Blocking prerequisites         |
| ----- | -------------------------------------------------------------------------- | ------------------------------ |
| #58   | Requestless GetCallerIdentity generation                                   | Accepted generator baseline    |
| #59   | Anonymous OIDC/SAML RPC generation                                         | Accepted generator baseline    |
| #60   | Four-operation acceptance, consumer comparison and source-update rehearsal | #58, #59; accepted #51/#53/#55 |
| #61   | Scoped release and versioned pkg.go.dev evidence                           | #60                            |

Next implementation: #58, then #59; no API is handcrafted to bypass either gate.
The parent aggregates delivery and is not a prerequisite that blocks its children.

## 中文

### 权威与范围

用户于 2026-10-08 指定首版目标为 **v0.1.0：STS 完整走通 Darabonba 生成链路**。
这个限定的实验版本优先于旧 ECS/VPC/STS 候选 Beta 的排期，但保留其更广泛的
[验收标准](product-acceptance.md)及历史证据；完成本里程碑不代表整体 Beta。
本次规划不创建版本标签或新云资源。

采用固定官方 STS 2015-04-01 DSL/导入模块、官方语义 parser、规范化完整 IR、
本项目 Go 后端及共享 runtime。`service/sts` 交付固定来源的四操作：AssumeRole、
GetCallerIdentity、AssumeRoleWithOIDC、AssumeRoleWithSAML；升级来源时重新核对清单，
不宣称永久全量或全云覆盖。保留 Go 1.27、直接 encoding/json/v2、AWS 调用范式、
显式凭据 provider、所有权、安全错误和带许可的双语文档。

### 基线及缺口

规划时基线发现四操作，只输出 AssumeRole；OIDC/SAML 的 DSL_PROTOCOL_PROFILE 对应
Anonymous RPC/doRPCRequest，GetCallerIdentity 的 DSL_OPERATION_SIGNATURE 对应
没有请求模型参数。分 issue 修复前端/IR/后端/runtime 契约，不手写操作、不静默签名
匿名请求、不隐式发现密钥、不直接改生成文件。匿名认证是显式逐操作协议契约，
不允许需要签名的操作省略凭据。

已接受 #51 完整 DSL provider/cache 组合、#53 STS 优先指南及显式来源、#55 真实
900 秒到期续期（三次签发、四次生成 ECS 读取，临时 IAM 已全部清理）。这些关闭 issue
作为交付证据纳入，不为新里程碑重建云环境或重复同一真实测试。

### 顺序与证据

1. 支持无请求模型 RPC 签名，生成 GetCallerIdentity，保留操作名、context、Options、
   mock 接口、准确完整响应及元数据。
2. 支持审核的 Anonymous RPC/doRPCRequest，生成 OIDC/SAML，保留线路字段及缺失语义，
   不签名/不读取 provider，默认格式隐藏 token/assertion 及响应凭据。前两项可在独立
   issue 分支推进。
3. 两项完成后验收四操作：可重现的离线线路/响应/负面契约、provider/cache 回归、固定
   版本外部消费者及官方 v2 对比、授权 GetCallerIdentity 真实证据，以及真实上游版本
   升级/漂移演练。已有 #55 继续作为 AssumeRole/续期证据，除非相关修改使其失效。
4. 准备兼容/迁移说明、许可、双语版本说明、对应版本 pkg.go.dev 检查；发布前门槛
   全通过才发布，发布/索引证据完成后才能关闭里程碑。

| 门槛        | 必需证据                                                                                             |
| ----------- | ---------------------------------------------------------------------------------------------------- |
| 生成        | 四操作发现/降低/输出，完整可达模型，确定性前端/IR/Go 检查，选中不支持行为写前失败                    |
| 行为        | 四操作编译及独立离线协议/字段/响应/错误/取消测试与外部 Example，签名/匿名分离和敏感格式              |
| 凭据        | 显式 provider，生成 AssumeRole→provider/cache→消费者，保留所有权/并发/取消契约及 #55 限定真实续期    |
| 消费者/维护 | 固定版本的外部 STS 任务与官方 v2 示例，独立 Go 用户仅按文档完成任务的证据，真实上游升级/漂移审核     |
| 真实范围    | #55 AssumeRole/续期及授权 GetCallerIdentity；SDK/CLI 与 Explorer 浏览器证据分别记录                  |
| 文档/发布   | 英文 pkg.go.dev 注释、配对指南、许可、Linux race/Windows CI、不可变 v0.1.0 标签/发布及同版本网页证据 |

OIDC/SAML 真实成功联邦调用需要另行授权的 IdP/provider/assertion 环境，**预先明确不属于
首版必需真实范围**；记录 NOT RUN，并在版本说明披露，不替代四操作必需离线正确性。
用户后续 #68 修正使默认配置/原生 Profile/OAuth 续期成为发布前必需，遵循
[更新路线](default-configuration.md)；联邦 credential-provider helper 仍为后续范围。本例 STS 没有
分页/waiter 任务，不编造分页，保留共享引擎。ECS 仅作为已有角色凭据消费者；ECS/VPC
全产品验收和基准 #20 独立。

### GitHub 协调

建立 `v0.1.0` milestone 和关联本仓库的 Projects v2 项目
`alicloud-go-sdk-x: v0.1.0 STS`。Milestone 表示发布范围，issue 承载验收/依赖/证据，
Project Status 表示执行进度。每 issue 只有一个状态 label：满足前提为 ready/Todo；
执行中为 in-progress/In Progress；依赖未完成为 blocked/Todo 并列出前提；验收完成为
done/Done。随 issue 状态变化维护 Project，不宣称已自动同步；代码合并不等于发布完成。

父 issue 跟踪子项直到发布门槛通过；子项关联父 issue 表示归属，不把父 issue 作为阻塞
依赖，保持无环。已有关闭 STS 证据加入里程碑/项目；历史生成器/基础 issue 和其他项目
保留原范围。不随意设截止日期；选择 v0.1.0 下一项前阅读本文。

[Milestone v0.1.0](https://github.com/rambow-cloud/alicloud-go-sdk-x/milestone/4) |
[Project](https://github.com/orgs/rambow-cloud/projects/3) |
[父 issue #57](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/57)

| Issue | 工作                                 | 阻塞前提                     |
| ----- | ------------------------------------ | ---------------------------- |
| #58   | 无请求模型 GetCallerIdentity 生成    | 已接受生成器基线             |
| #59   | 匿名 OIDC/SAML RPC 生成              | 已接受生成器基线             |
| #60   | 四操作验收、消费者对照及来源升级演练 | #58、#59；已完成 #51/#53/#55 |
| #61   | 限定发布及对应版本 pkg.go.dev 证据   | #60                          |

下一实现先 #58，再 #59，不手写 API 绕过门槛。父项汇总交付，不阻塞子项。

## STS delivery status / STS 交付状态

### English

#58/PR #63 and #59/PR #64 are merged: all four pinned STS actions emit with native signed/anonymous separation. #60 delivers consumer/real-source/live-identity evidence and the independent developer handoff; keep its UX gate open until the user-arranged Go developer records actual results. #61 publication/indexing follows that required acceptance. Historical planning counts below describe the earlier baseline, not current coverage. See [acceptance evidence](sts-v010-acceptance-report.md).

### 中文

#58/PR #63 与 #59/PR #64 已合并，四个固定 STS 操作输出并保留原生签名/匿名分离。#60 交付消费者/真实来源/真实 identity 证据及独立开发者验收包，用户安排的 Go 开发者提交实际结果前 UX 门槛保持开放。#61 发布/索引以此必需验收为前提，下文规划数量是旧基线不是当前覆盖。见[验收证据](sts-v010-acceptance-report.md)。
