## English

### Problem and evidence

Pinned STS main.tea:294/426 use authType=Anonymous and doRPCRequest. Coverage reports DSL_PROTOCOL_PROFILE for AssumeRoleWithOIDC/SAML. Current signed runtime must not be bypassed by globally permitting missing providers.

### Scope and dependencies

Member of #57; accepted #35/#36/#37/#38 baseline. Independent of requestless operation work. Normalize constant doRPCRequest handoff; represent explicit operation authentication in IR and safely execute reviewed anonymous RPC over HTTPS. Generate both complete operations/models/interfaces/docs. Tokens/assertions are supplied explicitly; no implicit discovery or new federation provider helper. Preserve signed operations' missing-provider rejection and conservative no-retry defaults.

### Acceptance

- [ ] Both operations lower/emit/compile with exact native fields, complete responses and optional presence semantics.
- [ ] Anonymous calls omit signature/authorization/source security token and never retrieve a credential provider; signed AssumeRole/GetCallerIdentity still require explicit providers.
- [ ] Independent wire/body/response/error/context/ownership/negative protocol fixtures verify the official helper contract; unknown/dynamic auth/handoff fails before writes.
- [ ] Default formatting/errors/middleware/traces protect OIDC tokens, SAML assertions and returned credentials.
- [ ] English docs, paired guides and deterministic offline Examples disclose that successful live federation is NOT RUN, outside v0.1.0 required live scope.

### Verification

Node 22 contracts before both generation checks, doccheck, formatting, vet, tests/Examples, Linux race and Windows CI. Establish exact protocol evidence from pinned official imports before implementing; no unit-test cloud/IdP access.

## 中文

官方 OIDC/SAML 的 Anonymous/doRPCRequest 造成 DSL_PROTOCOL_PROFILE。归属 #57，基于已接受生成器，可独立于无请求操作开发。规范化官方固定 handoff、显式表示逐操作认证，并用 HTTPS 执行匿名 RPC；完整生成两操作，不全局放开签名操作的缺 provider 规则，不新增隐式发现/联邦 helper。验收准确字段/存在/响应、无签名和来源 token/不读 provider、独立协议/错误/取消/所有权/负面契约、非法行为写前失败以及 token/assertion/响应凭据默认隐藏。同步英文 docs、双语指南、离线 Example；真实联邦预先声明范围外并记 NOT RUN。按上述 Node/Go/CI 门禁执行，先核官方导入协议，不在测试接云/IdP。
