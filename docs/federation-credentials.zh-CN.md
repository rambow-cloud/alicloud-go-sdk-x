# 可续期的联邦身份凭据

[English](federation-credentials.md)

- 对应 issue：#89。复用固定官方 Darabonba 源码生成的 `service/sts` OIDC 与 SAML API。
- `feature/stscreds.NewAssumeRoleWithOIDCProvider` 和 `NewAssumeRoleWithSAMLProvider` 将这些 API 适配为 `credentials.Provider`。
- 签发客户端配置 `credentials.AnonymousProvider{}`；应用客户端使用签发返回的临时凭据。
- 使用 `credentials.Cache` 合并并限制刷新请求。辅助 provider 本身不缓存，也不重试凭据签发。

## Token 所有权与续期

- 提供支持并发的 `TokenProvider`，每次刷新返回当前身份材料。
- `TokenProviderFunc` 接收支持 context 的函数。自定义来源错误保留其标识；错误是否泄露敏感信息由自定义来源负责。
- `NewFileTokenProvider` 只保存文件名，不在构造时读文件。获取 token 时读取最多 1 MiB 的普通文件，去除首尾空白。
- 外部 IdP 轮换 token 时应原子替换文件。SDK 不写入文件，也不记录其中内容。
- OIDC 使用原始 token；SAML 使用完整 SAML 响应的 Base64 编码。SDK 不解码身份材料。
- 辅助 provider 输入中的 `OIDCToken`/`SAMLAssertion` 保持 nil；同时嵌入 token 并指定来源会被拒绝。
- 构造时复制输入和选项注册列表。共享 API、来源及回调须支持并发；回调不能保留请求或选项对象。
- 到期、提前刷新或 `Invalidate` 后，缓存通过来源重新取得身份材料。清空缓存不会撤销已经签发的云端凭据。

## 校验规则

- 必须提供角色与身份提供商 ARN；OIDC 还需提供会话名。
- 显式指定少于 900 秒的有效期会被拒绝；nil 保留服务默认值。最长有效期由服务端决定。
- 已审核的 token 长度范围：OIDC 4～20,000 字节，SAML 4～100,000 字节。辅助 provider 不在本地验证 JWT 签名或 XML 断言。
- 响应必须包含非空的 AK、密钥、SecurityToken 及未来的 RFC3339 到期时间。无效响应不会进入缓存。
- 取消与 API 错误可通过 `errors.Is`/`errors.As` 判断。
- 文件及 provider 的默认格式不显示路径或秘密；文件错误不包含原始操作系统错误内容。

## 默认 OIDC 发现

- 配置 `ALIBABA_CLOUD_ROLE_ARN`、`ALIBABA_CLOUD_OIDC_PROVIDER_ARN` 和 `ALIBABA_CLOUD_OIDC_TOKEN_FILE`。
- 可选的 `ALIBABA_CLOUD_ROLE_SESSION_NAME` 默认为 `alicloud-go-sdk-x`。
- 优先级：显式 provider、显式 Profile、环境临时密钥及 token、OIDC 环境配置、自动选择的原生 Profile。
- 已选中的 OIDC 来源配置不完整时直接报错，不静默回退。
- 加载配置时不发 HTTP 请求，也不读取 token 文件。签发使用公共地址 `https://sts.aliyuncs.com`；应用地域仍按原有选项、环境与 Profile 的顺序解析。
- 长期密钥仍须显式启用。SAML 来源须显式注册，SDK 不自行猜测断言来源。

## 示例与证据

- `go test ./feature/stscreds ./config` 验证输出确定的 OIDC/SAML 示例、token 轮换、缓存合并、匿名请求、输入所有权、无效结果、取消及来源优先级。
- 这些测试不需要外部身份提供商或云资源。
- 真实联邦续期由 #94 跟踪；离线通过不表示真实联邦验收完成。
- 协议依据：固定的 `sources/darabonba/products/sts/main.tea`、[官方 OIDC 操作](https://www.alibabacloud.com/help/en/ram/developer-reference/api-sts-2015-04-01-assumerolewithoidc)及[官方 CLI 环境变量](https://www.alibabacloud.com/help/en/cli/environment-variables)。
