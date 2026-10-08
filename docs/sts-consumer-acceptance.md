# STS consumer acceptance / STS 消费者验收

## English

Issue #60 validates the four generated actions and consumer maintenance, separately
from publication #61. The isolated `examples/stsacceptance` module pins the official
STS v2.1.0 comparison; the SDK uses an explicit local replacement to this checked-out
repository. Record `git rev-parse HEAD`, `go version`, OS and module versions in the
report. Official comparison dependencies never enter the SDK runtime module.

The implementation author's offline execution is technical evidence. The user has
arranged an independent Go developer; their docs-only result remains required.
Successful live federation is NOT RUN/outside release scope; never supply real
credentials to this offline kit or create cloud resources.

### Independent developer handoff

Check out the exact commit supplied with the handoff, use Go 1.27+, and read
[STS composition](sts-credentials.md), [credentials](credentials.md),
[the generated guide](products/sts.md) and [anonymous RPC](sts-anonymous-rpc.md).
From `examples/stsacceptance`, run `go test -v ./...` and `go run .`.
These run with scripted local HTTP fixtures and fixed placeholders. Dependencies
may download on the first build; operation execution has no account/network access.

In a separate scratch Go module, without implementation-author assistance:

1. Call GetCallerIdentity with an injected scripted transport, context and operation
   options; inspect the concrete output and Metadata.
2. Call native AssumeRole, then compose the generated client with
   AssumeRoleProvider and Cache to consume temporary credentials, with no response
   translator or application refresh loop. Use fixtures, not a cloud role.
3. Substitute a small AssumeRoleAPI mock; classify an APIError with errors.As and a
   canceled operation with errors.Is.
4. Configure AnonymousProvider and call OIDC/SAML with explicit fixture fields;
   explain why signed operations cannot use this marker, nil is rejected, and STS
   supplies no paginator or waiter.
5. Run the pinned official-v2 comparison workload; record the concrete differences
   in context/options/envelope/provider/mock use. Do not infer timing or superiority.

Record each task's PASS/FAIL, start/end/elapsed minutes, docs used, compilation
errors, obstacles, assistance and proposed fixes in the paired
[result template](sts-independent-result-template.md). Report source revision and
versions. Any required FAIL/NOT RUN keeps #60 and release #61 open. Do not publish
account identity, credentials, assertions or raw requests/responses in evidence.

## 中文

#60 验收四个生成操作与消费者维护，与 #61 发布分开。隔离的 `examples/stsacceptance`
模块固定官方 STS v2.1.0；本 SDK 通过显式本地 replace 使用当前检出源码。报告记录
`git rev-parse HEAD`、`go version`、系统和模块版本；对比依赖不进入 SDK 运行时模块。

实现者离线运行只是技术证据；用户安排的独立 Go 开发者仍需仅按文档完成任务。
真实成功联邦调用预先范围外且记 NOT RUN；离线包不要传入真实凭据或创建云资源。

### 独立开发者交接

检出交接时提供的准确提交，使用 Go 1.27+，阅读上述 STS 组合、凭据、生成产品和匿名协议
指南，在 `examples/stsacceptance` 执行 `go test -v ./...` 与 `go run .`。固定占位值与脚本
HTTP fixture 无需账号，第一次构建可能下载依赖，操作运行不访问网络。

在独立临时 Go 模块内、不接受实现者协助，完成：

1. 使用脚本 transport、context、操作 options 调用 GetCallerIdentity，检查具体输出与 Metadata。
2. 调用原生 AssumeRole，再用生成 client、AssumeRoleProvider 与 Cache 消费临时凭据，不写
   response translator 或应用刷新循环；使用 fixture，不使用云角色。
3. 替换小 AssumeRoleAPI mock，用 errors.As 识别 APIError、errors.Is 识别取消。
4. 配置 AnonymousProvider，显式传入 fixture 调用 OIDC/SAML，解释签名操作为何不能使用
   此标记、nil 被拒绝及 STS 为什么不提供分页或 waiter。
5. 运行固定官方 v2 对比任务，记录 context/options/envelope/provider/mock 的具体差异，
   不推断速度或优劣。

按[结果模板](sts-independent-result-template.md)记录各项 PASS/FAIL、起止/分钟、所用文档、
编译错误、障碍、协助、改进和源码/环境版本。必需项 FAIL/NOT RUN 使 #60/#61 保持开放。
证据不得发布账号身份、凭据、assertion 或原始请求/响应。
