# Anonymous STS RPC / 匿名 STS RPC

## English

Issue #59 supports AssumeRoleWithOIDC/SAML from pinned official DSL. These functions
use authType=Anonymous and the eight-argument doRPCRequest handoff; this differs
from the signed callApi path. The pinned imported OpenApi 0.3.23 main.tea:176-290
supplies Action, Version, Format=json, UTC Timestamp and SignatureNonce in the query,
plus action/version headers. With no request.body, the body is empty. The Anonymous
branch does not obtain credentials or add a signature. Preserve this reviewed wire
contract; do not reuse ACS3 signing or silently accept dynamic/unknown handoffs.

The explicit credentials.AnonymousProvider marker lets an application construct a
client without inventing source keys; nil/typed-nil providers remain invalid. It
never supplies signing credentials. Generated operations choose the reviewed
AuthenticationAnonymousRPC mode; signed operations remain signed and reject an
anonymous marker. A supplied custom/static/cache provider is not retrieved by an
anonymous operation. Configuration is not mutated. This follows the explicit-marker
idea in [AWS AnonymousCredentials](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws#AnonymousCredentials)
while retaining our per-operation protocol and strict missing-provider conventions.

OIDCToken/SAMLAssertion are explicit native request fields. Runtime preparation strips
source authentication headers/query parameters, preserves exact service field values,
and rejects unsupported method/path/body before transport. Policies redact complete
sensitive request/response models in String/GoString; explicit JSON remains raw.
Error/tracing defaults never record raw URLs/bodies/tokens. Retry remains opt-in and
these token-issuing operations remain non-idempotent under the reviewed policy.

Acceptance uses independent offline common/service query, encoded token/assertion,
empty-body/header/provider-isolation, complete response/presence, JSON-v2/error/context,
middleware/tracing/formatting and signed-operation regression contracts. Generate
English comments, paired guides/source indexes and deterministic external Examples.
Successful live OIDC/SAML federation is NOT RUN, outside v0.1.0 required live scope;
IdP/provider/assertion provisioning and federation credential helpers are future work.

## 中文

#59 从固定官方 DSL 支持 AssumeRoleWithOIDC/SAML，使用 Anonymous 及八参数 doRPCRequest，
与签名 callApi 路径不同。固定 OpenApi 0.3.23 main.tea:176-290 的导入实现提供 query 中的
Action/Version/Format=json/UTC Timestamp/SignatureNonce 及 action/version header；
没有 request.body 时发送空体，Anonymous 分支不读凭据/不添加签名。保留这个审核线路
契约，不复用 ACS3 签名，不静默接受动态或未知 handoff。

应用显式声明 credentials.AnonymousProvider，无需伪造来源密钥；nil/typed-nil 仍非法，
该 marker 不提供签名凭据。生成操作显式选择 AuthenticationAnonymousRPC；签名操作
保持签名且拒绝匿名 marker。匿名调用不读取配置的 custom/static/cache provider，不改
共享配置。借鉴上述 AWS 显式 marker 思路，保留本项目逐操作协议及严格缺 provider 规则。

OIDCToken/SAMLAssertion 显式输入；运行时移除来源认证 header/query，保留准确服务参数，
不支持的 method/path/body 在 transport 前拒绝。策略使敏感完整请求/响应 String/GoString
脱敏，显式 JSON 仍原样。默认错误/trace 不记录 URL/body/token。重试仍显式启用，
这两项签发操作按审核策略保持非幂等。

独立离线验收公共/服务 query、token/assertion 编码、空体/header/provider 隔离、完整
响应/存在、JSON v2/错误/取消、middleware/trace/格式及签名回归；同步英文注释、配对
指南/来源和确定性外部 Example。OIDC/SAML 真实成功联邦预先范围外并记 NOT RUN，
IdP/provider/assertion 环境与联邦 credential helper 后续单独建设。
