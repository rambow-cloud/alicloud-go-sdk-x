# Supplemental STS consumer review / STS 补充消费者评审

## English

Issue [#60](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/60).
Result: **PASS for the five supplemental technical cases**, subject to exact-head
Linux race/Windows CI before integration. This is an implementation-agent review,
not a Go developer's independent docs-only acceptance. The human result remains
NOT RUN; #60/#61 and publication/indexing gates remain open.

Reviewed main baseline: `3e6a5a720e9ba6afd5c995102b7b7b56a7f955b3`.
Environment: Go 1.27.1, Windows/amd64, Node 22.21.1. The isolated consumer module
uses a local SDK replace and pins official STS v2.1.0, OpenApi v2.1.13, Tea v1.3.13.
This change adds tests and documentation only; SDK runtime, generated outputs,
source pins and policies stay at the reviewed baseline.

### Method and findings

The separate `main_test` package in
[consumer_review_test.go](../examples/stsacceptance/consumer_review_test.go) imports
only public SDK packages, builds its own transport/mocks and does not reuse the
original consumer fixture or SDK internals. Its transports never open a connection.
Credential/identity/token/ARN values are fictional. Consumer construction follows
the [STS composition](sts-credentials.md), [credentials](credentials.md),
[cache contracts](credential-cache.md), [product guide](products/sts.md) and
[anonymous protocol](sts-anonymous-rpc.md). Public declarations and signing evidence
were inspected during review, so this is not labeled an unaided docs-only exercise.

| Case                          | Observation                                                                                                                                                                                                                                                                                            | Result |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ |
| Identity/options/cancellation | Native fields and RequestID/status/attempt metadata; per-call endpoint override leaves the next call unchanged; transport cancellation retains errors.Is and OperationError                                                                                                                            | PASS   |
| Native role/provider/cache    | Caller mutates its session after provider construction; original session and explicit 900 seconds survive. Canceling the initiating waiter leaves the shared issuance alive; 24 concurrent consumer reads use the issued role token with exactly one issuance and no application refresh loop          | PASS   |
| Small mock/errors             | A single AssumeRoleAPI mock suffices; errors.As and errors.Is preserve the exact APIError; already-canceled retrieval never reaches the mock                                                                                                                                                           | PASS   |
| Anonymous federation          | Both OIDC/SAML preserve `+ / = &` and spaces, common query fields and empty POST body; four calls across AnonymousProvider and an unusable custom provider perform zero source retrievals. Signed identity rejects the marker; nil/typed-nil construction fails; default model formatting hides tokens | PASS   |
| Unsafe retry/error formatting | AssumeRole with Standard and HTTP 503 executes once; APIError and OperationError retain code/status/request ID/attempts; the explicit service Message is available while default Error strings omit it                                                                                                 | PASS   |

Initial execution at 2026-10-08 13:15:03 UTC passed four cases; the cache case
failed because the reviewer incorrectly expected the AWS-style slash after
`Credential=<key>`. Alibaba ACS3 uses a comma before SignedHeaders. Correcting
that fixture assertion and adding early-error reporting made that case pass on
its targeted rerun. No SDK fix was needed. This is evidence for the documented
boundary: AWS-like Go calling conventions with Alibaba-native wire semantics.

The pinned official-v2 four-action fixture also passed in this session. Its methods
use RuntimeOptions and response Body envelopes, while this SDK exposes context-first
operations, service functional options, direct native outputs plus Metadata and
small operation interfaces. This review does not measure performance, evaluate all
official versions or evaluate the official credentials library's refresh support.

A documentation defect was found and corrected: the previous acceptance report
grouped AC-06 through AC-12 under mismatched capability descriptions. Its references
now match [the authoritative acceptance table](product-acceptance.md): retry, endpoint,
middleware/OTel, errors, testing, documentation and generation respectively.

### Verification and limits

Commands from the repository root after configuring local Go caches:

```text
go -C examples/stsacceptance test -count=1 -v -run '^TestConsumerReview' ./...
go -C examples/stsacceptance test -count=1 -v -run '^TestConsumerReviewNativeRoleCacheAndCanceledWaiter$' ./...
go -C examples/stsacceptance test -count=1 -v -run '^TestPinnedOfficialWorkload$' ./...
```

The first command's initial FAIL and the targeted correction above are both part
of the record; combined local results cover all five cases. Final full-module
Linux race/Windows runs are supplied by the linked PR's exact-head CI. Documentation,
vet and root tests are recorded there too. Test elapsed time is not independent
developer task time. No fresh cloud calls, IAM/IdP provisioning or source updates
were performed; already accepted [live renewal](live-sts-renewal.md),
[identity evidence](acceptance/sts-identity-live.json) and
[source rehearsal](sts-source-rehearsal.md) remain distinct.

No blocking runtime defect was found in this bounded review. Successful live
OIDC/SAML remains NOT RUN outside the predeclared required live scope; STS has no
native paginator/waiter in this workload. Native Profile/OAuth discovery, broader
product acceptance, human task usability/timing, release publication and same-version
pkg.go.dev browser indexing are not established by this report.

## 中文

本报告归属 [#60](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/60)。
**五项补充技术用例通过**，集成前还需准确提交的 Linux race/Windows CI。本次为实现
代理单独执行的评审，不冒充独立 Go 开发者仅按文档验收。人员结果仍 NOT RUN，#60/#61
及发布/索引门禁保持开放。

评审 main 基线为 `3e6a5a720e9ba6afd5c995102b7b7b56a7f955b3`；Go 1.27.1、
Windows/amd64、Node 22.21.1。隔离消费者模块本地 replace SDK，官方对比固定 STS
v2.1.0、OpenApi v2.1.13、Tea v1.3.13。本次只补测试与文档，运行时、生成物、来源和
策略沿用该基线。

### 方法与发现

[consumer_review_test.go](../examples/stsacceptance/consumer_review_test.go) 的
独立 `main_test` 包只导入 SDK 公开包，自建 transport/mock，不复用原消费者 fixture
或内部工具；transport 不建立网络连接，密钥、身份、token、ARN 全为虚构值。组合依据
上述 STS、凭据、缓存、产品与匿名协议指南。评审还查阅公开声明及签名证据，因此不标为
无人协助、仅按文档的独立人员任务。

| 用例                    | 观察                                                                                                                                                                                                      | 结果 |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---- |
| 身份/options/取消       | 原生字段与 RequestID/status/attempt metadata 正确；操作 endpoint 覆盖不影响下一次调用；调用中取消保留 errors.Is 与 OperationError                                                                         | PASS |
| 原生角色/provider/cache | 构造 provider 后修改调用方 session，不影响原始 session 和显式 900 秒；取消首个等待者不取消共享签发，24 个并发消费者使用角色 token，仅签发一次，不写应用刷新循环                                           | PASS |
| 小 mock/错误            | 只实现 AssumeRoleAPI；errors.As/Is 保留同一 APIError，已取消读取不进入 mock                                                                                                                               | PASS |
| 匿名联邦                | OIDC/SAML 准确保留 `+ / = &` 和空格、公共 query 与空 POST body；AnonymousProvider 及不可用自定义 provider 的四调用均零来源读取；签名 identity 拒绝 marker，构造拒绝 nil/typed-nil；默认模型格式隐藏 token | PASS |
| 非安全重试/错误格式     | Standard 下 AssumeRole 遇 503 只执行一次；结构错误保留 code/status/request ID/attempts，显式 Message 可读取，默认 Error 字符串不带 Message                                                                | PASS |

2026-10-08 13:15:03 UTC 首次执行四项通过；cache 用例因评审者误以为
`Credential=<key>` 后跟 AWS 风格斜线而失败。阿里云 ACS3 在 SignedHeaders 前使用逗号；
修正 fixture 断言并补提前错误报告后，单独重跑该项通过，无需修复 SDK。这印证设计边界：
Go 调用范式参考 AWS，线路语义保留阿里云。

本次官方固定版本的四操作 fixture 也通过。该版本使用 RuntimeOptions 和 Body envelope，
本 SDK 使用 context-first、服务 functional options、原生输出加 Metadata 及窄操作接口。
未测性能、未覆盖所有官方版本、未评估官方 credentials 库刷新能力。

发现并修正一项文档问题：旧报告将 AC-06 至 AC-12 与不准确的能力说明分组对应。
现按[权威验收表](product-acceptance.md)依次对应重试、endpoint、middleware/OTel、
错误、测试、文档和生成，不增加覆盖声明。

### 验证与限制

英文记录中的三个命令在根目录设置本地 Go cache 后运行；首次失败及定向修正均保留记录，
合并本地结果覆盖五项。最终完整隔离模块 Linux race/Windows 由 PR 准确提交 CI 提供，
相关文档/vet/根模块测试也在 PR 记录。测试耗时不是独立开发者任务耗时。本次无新增真实
调用、IAM/IdP 创建或来源升级；既有[真实续期](live-sts-renewal.md)、
[身份记录](acceptance/sts-identity-live.json)、[来源演练](sts-source-rehearsal.md)
仍为分别记录的证据。

本限定评审未发现阻塞运行时缺陷。OIDC/SAML 成功真实联邦仍 NOT RUN，预先位于必需真实
范围外；本 STS 任务无原生分页/waiter。本报告不证明原生 Profile/OAuth 发现、更多产品
验收、独立人员任务易用性/耗时、发布或同版本 pkg.go.dev 浏览器索引。
