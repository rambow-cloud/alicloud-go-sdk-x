# v0.1.0 STS acceptance evidence / v0.1.0 STS 验收证据

## English

Issue #60: technical acceptance and the independent developer gate are separate.
#58/PR #63 and #59/PR #64 are merged. Implementation version for the live/consumer
run is `38cf05ac458d2e6ed3af165350fb77a10e7e8817`, now on main through
`f31ad13e4ed3ba9974dfd5d5a9efff8c01aaf3a6`. Go 1.27.1, Windows/amd64,
Node 22.21.1, parser 2.2.1; Linux race and Windows CI preserve separate results.
The acceptance branch adds maintenance guards/fixtures and docs without changing
accepted STS source/model/policy/generated artifact bytes or signed runtime behavior.

### Scoped coverage

| Native action      | Discovered/lowered/emitted/compiled | Independent offline contracts/Examples                                             | Live                                                                                     |
| ------------------ | ----------------------------------- | ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| AssumeRole         | PASS                                | PASS; provider/cache consumption and narrow mock                                   | PASS #55: three issuances, four ECS reads, real 900-second expiry renewal and cleanup    |
| GetCallerIdentity  | PASS                                | PASS; complete output, wire/presence/errors/cancellation                           | PASS #60: six identity fields/presence match CLI, coherent metadata, single ACS3 attempt |
| AssumeRoleWithOIDC | PASS                                | PASS; unsigned RPC/token encoding/provider isolation/complete output/redaction     | NOT RUN; successful federation outside required v0.1.0 live scope                        |
| AssumeRoleWithSAML | PASS                                | PASS; unsigned RPC/assertion encoding/provider isolation/complete output/redaction | NOT RUN; successful federation outside required v0.1.0 live scope                        |

All four pinned actions emit with 19 reachable named/inline models; the empty identity
Input is not an official DSL model. Emission reports retain compilation/live as
not-assessed; this report supplies actual acceptance evidence separately.
STS has no native pagination/waiter workload in this scope. Other products retain
their own evidence and are not promoted to first-release product acceptance.

### Consumer and maintenance

`examples/stsacceptance` pins the official [STS v2.1.0](https://github.com/alibabacloud-go/sts-20150401/tree/v2.1.0),
darabonba-openapi/v2 v2.1.13 and tea v1.3.13; go.sum locks its isolated dependency
graph. Our explicit local replace uses the checked-out SDK revision, recorded by
the independent developer. Both fixtures execute four actions; our workload also
performs native AssumeRole -> provider/cache -> two role-signed identity reads.
Three external tests and runnable consumer PASS. No translator or application
refresh loop is needed; performance and official credential-library refresh are
not measured. Fixed versions have concrete context/options/envelope/mock differences,
not a source-compatibility or superiority claim.

Real source rehearsal compares `c321394a58d9b6e513fabb898ee9857a2c6df852` with
`d2c0338636a58a6cafc5316d2ed5d158f1f5b162`, retaining original licensed bytes,
blob hashes and pinned imports. Native fields/models/types/requiredness and semantic
prose are unchanged. Real initializer signing/endpoint mappings and anonymous handoffs
change; prose coordinates move. Historical v2 signed initialization and Anonymous/
callApi are unsupported, not accepted as ACS3. Candidate four-action discovery,
emission and standalone contracts/Examples PASS; stale policy and historical signed
selection fail before writes. Current candidate STS bytes match production; original
source/IR/policies/generated outputs remain unchanged. [Commands/review](sts-source-rehearsal.md)
and [machine evidence](acceptance/sts-source-rehearsal.json) distinguish actual
historical drift, the no-change production update and synthetic negative cases.

The authorized read-only live identity run at 2026-10-08 11:33:27 UTC uses `oss-sftp`
OAuth/STS only as an explicitly injected in-memory snapshot after CLI authentication.
[Sanitized evidence](acceptance/sts-identity-live.json) contains no keys/account IDs/
ARNs/tokens/raw responses; no cloud objects were created. This is not native Profile/
OAuth renewal. Explorer browser evidence is NOT RUN, distinct from CLI/SDK comparison.
#55 natural-renewal evidence is reused: signed credential retrieval/signing/cache and
AssumeRole source behavior are preserved; anonymous additions do not alter that path.

### Gate status and handoff

AC-01/02: typed context/options, all modeled fields/presence, independent protocol
and strict-JSON fixtures. AC-05: explicit provider/cache plus scoped #55 live renewal.
AC-06/07/08: structured safe errors, ownership/middleware/provider isolation and OTel.
AC-09/10/11/12: docs/Examples, pinned consumer/real-source rehearsal, dependency/credential
constraints and licensed provenance. These are scoped STS mappings, not broader Beta.

Local verification is recorded in the linked PR: frontend/IR checks, 56 Node cases,
both generators, doccheck, formatting/bilingual checks, vet, Go tests/Examples,
isolated consumer and standalone regenerated STS. Exact-head Linux race/Windows CI
must pass before merging the technical delivery.

UX-04/05 independent docs-only tasks: **NOT RUN — user arranging another Go developer**.
The author's fixture execution does not establish independent task time/success.
Provide [instructions](sts-consumer-acceptance.md) and the
[paired result template](sts-independent-result-template.md) at the exact PR commit.
Keep #60 open until actual independent results pass. #61 release/indexing is blocked
by this required evidence; no tag or pkg.go.dev indexing is claimed.

## 中文

#60 的技术验收和独立开发者门槛分别记录。#58/PR #63、#59/PR #64 已合并。
真实/消费者运行的实现版本为 `38cf05ac458d2e6ed3af165350fb77a10e7e8817`，通过
`f31ad13e4ed3ba9974dfd5d5a9efff8c01aaf3a6` 进入 main。Go 1.27.1、Windows/amd64、
Node 22.21.1、parser 2.2.1；Linux race/Windows CI 单独记录。验收分支补维护门禁/
fixture/文档，不改变已接受 STS 来源、模型、策略、生成字节或签名 runtime 行为。

### 限定覆盖

| 原生操作           | 发现/降低/输出/编译 | 独立离线契约/Examples                                     | 真实                                                                     |
| ------------------ | ------------------- | --------------------------------------------------------- | ------------------------------------------------------------------------ |
| AssumeRole         | PASS                | PASS；provider/cache 消费及小 mock                        | PASS #55：三次签发、四次 ECS 读取、真实 900 秒到期续期与清理             |
| GetCallerIdentity  | PASS                | PASS；完整输出、线路/存在/错误/取消                       | PASS #60：六身份字段及缺失状态与 CLI 一致、metadata 一致、一次 ACS3 尝试 |
| AssumeRoleWithOIDC | PASS                | PASS；匿名 RPC/token 编码/provider 隔离/完整响应/脱敏     | NOT RUN；成功联邦预先位于 v0.1.0 必需真实范围外                          |
| AssumeRoleWithSAML | PASS                | PASS；匿名 RPC/assertion 编码/provider 隔离/完整响应/脱敏 | NOT RUN；成功联邦预先位于 v0.1.0 必需真实范围外                          |

四个固定操作输出 19 个可达命名/内联模型；空 identity Input 不算官方 DSL 模型。
生成报告的编译/真实仍为 not-assessed，本报告另提供验收证据。本范围 STS 无原生
分页/waiter 任务，其他产品保留各自证据，不升级为首版产品验收。

### 消费者与维护

隔离模块固定上述官方 STS v2.1.0、OpenApi v2.1.13、tea v1.3.13 和 go.sum；本 SDK
显式本地 replace，独立开发者记录检出提交。两 SDK fixture 都跑四操作，本任务另用
原生 AssumeRole → provider/cache → 两次角色签名 identity 读取。三个外部测试及程序
PASS，无应用 translator/刷新循环；未测性能或官方 credentials 库刷新。不宣称源码
兼容或优劣，只比较固定版本 context/options/envelope/mock 的实际差异。

真实修订演练比较上述 c321 与 d2c，保留原始许可/blob/固定导入。原生字段/模型/类型/
必需性及语义说明不变，initializer 签名/endpoint 映射、匿名 handoff 和说明坐标变化。
历史 v2 初始化与 Anonymous/callApi 不支持，不能当作 ACS3 验收。当前候选四操作发现、
生成、独立编译契约/Examples PASS；旧策略及历史签名选择写前失败。候选 STS 字节与
生产相同，原来源/IR/策略/生成物未变。[命令/审核](sts-source-rehearsal.md)及
[机器证据](acceptance/sts-source-rehearsal.json)区分真实历史变化、无变化生产更新与合成负面案例。

2026-10-08 11:33:27 UTC 的授权只读 identity 使用 CLI 登录后的 `oss-sftp` OAuth/STS
内存快照显式注入。[脱敏证据](acceptance/sts-identity-live.json)无密钥/账号/ARN/token/
原始响应，不建云资源，不代表原生 Profile/OAuth 续期。Explorer 浏览器记 NOT RUN，
与 CLI/SDK 证据分开。签名读取/签名/cache/AssumeRole 来源行为保留，匿名扩展不影响该
路径，因此复用 #55 自然续期证据，不重复已完成云操作。

### 门槛与交接

限定映射：AC-01/02 强类型 context/options、完整建模字段/存在/协议/严格 JSON；AC-05
显式 provider/cache 及 #55；AC-06/07/08 安全结构错误、所有权/middleware/provider 隔离/
OTel；AC-09/10/11/12 文档/Examples、固定消费者/真实来源、依赖/凭据约束及许可来源，
均不表示更广 Beta。PR 记录前端/56 Node 案例、两生成器、doccheck、格式/双语、vet、
Go 测试/Examples、隔离消费者及再生成 STS；技术合并前准确提交 Linux race/Windows CI
必须通过。

UX-04/05 独立文档任务：**NOT RUN，用户安排另一位 Go 开发者**。实现者 fixture 不证明
独立耗时/成功。按固定 PR 提交交付[任务](sts-consumer-acceptance.md)和
[双语模板](sts-independent-result-template.md)，实际独立结果通过前 #60 开放；必需证据
阻塞 #61 发布/索引，不宣称 tag 或 pkg.go.dev 已索引。
