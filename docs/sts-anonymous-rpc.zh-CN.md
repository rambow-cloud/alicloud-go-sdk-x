# 匿名 STS RPC

[English](sts-anonymous-rpc.md)

- #59 从固定官方 DSL 支持 AssumeRoleWithOIDC/SAML，使用 Anonymous 及八参数 doRPCRequest， 与签名 callApi 路径不同。
- 固定 OpenApi 0.3.23 main.tea:176-290 的导入实现提供 query 中的 Action/Version/Format=json/UTC Timestamp/SignatureNonce 及 action/version header； 没有 request.body 时发送空体，Anonymous 分支不读凭据/不添加签名。
- 保留这个审核线路契约，不复用 ACS3 签名，不静默接受动态或未知 handoff。

- 应用显式声明 credentials.AnonymousProvider，无需伪造来源密钥；nil/带类型的 nil 仍非法， 该 marker 不提供签名凭据。
- 生成操作显式选择 AuthenticationAnonymousRPC；签名操作保持签名且拒绝匿名 marker。
- 匿名调用不读取配置的 custom/static/cache 凭据提供者，不改共享配置。
- 借鉴上述 AWS 显式 marker 思路，保留本项目逐操作协议及严格缺凭据提供者规则。

- OIDCToken/SAMLAssertion 显式输入；运行时移除来源认证 header/query，保留准确服务参数， 不支持的 method/path/body 在 HTTP 传输实现前拒绝。
- 策略使敏感完整请求/响应 String/GoString 脱敏，显式 JSON 仍原样。
- 默认错误/trace 不记录 URL/body/token。
- 重试仍显式启用， 这两项签发操作按审核策略保持非幂等。

- 独立离线验收公共/服务 query、令牌或身份断言编码、空体/header/凭据提供者隔离、完整响应/存在、JSON v2/错误/取消、中间件/trace/格式及签名回归；同步英文注释、配对指南/来源和确定性外部 Example。
- OIDC/SAML 真实成功联邦预先范围外并记 NOT RUN， IdP/凭据提供者/身份断言 环境与联邦 credential 辅助组件后续单独建设。

## 参考资料

- [AWS AnonymousCredentials](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws#AnonymousCredentials)
