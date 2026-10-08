# Credentials / 凭据

[English](#english) | [中文](#中文)

## English

### STS-first application configuration

Prefer a renewable `stscreds.AssumeRoleProvider` wrapped in `credentials.Cache` for
application clients. The full generated-client composition is in [the STS guide](sts-credentials.md)
and the runnable README/ExampleAssumeRoleProvider. Configure a separate source provider
for STS to avoid recursive retrieval. Cache coalesces refresh; a copied STS token in
StaticProvider never renews itself.

`alicloud.Config` and service Options expose only `CredentialsProvider`, never bare
AccessKey/KeySecret/SecurityToken fields. Explicit provider injection is required even
when environment credentials are populated. Nil/typed-nil providers, including a nil
ProviderFunc, fail during runtime/chain/cache construction before HTTP; construction
does not retrieve credentials. Per-operation configuration replacement applies the
same validation without changing the client.

| Source                    | Explicit declaration                                                                               | Intended use                                                                   |
| ------------------------- | -------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| Renewable role            | `stscreds.NewAssumeRoleProviderFromClient(...)`, then `credentials.NewCache(...)`                  | Primary application path; generated STS -> role provider -> generated consumer |
| Long-lived AK/SK          | `credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: id, AccessKeySecret: secret})` | Secondary deliberate opt-in; for example an authorized source for STS          |
| Static temporary snapshot | `NewStaticProvider` with SecurityToken and ExpiresAt                                               | Already-issued temporary credentials; expiration checked, never renewed        |
| Environment               | `credentials.EnvProvider{}`                                                                        | Explicit environment lookup, never auto-registered                             |
| Custom provider           | A `credentials.Provider` implementation or `ProviderFunc`                                          | Application-managed source; preserve context/concurrency/error contracts       |

All deliberate custom providers remain supported, including ones returning long-lived
keys. STS-first is guidance and a provider-only configuration rule, not an STS-only
runtime. AWS Go SDK v2 likewise accepts providers and documents an explicit
[StaticCredentialsProvider](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html).
Unlike AWS LoadDefaultConfig, this SDK has no implicit default source discovery. Do not
claim a default credential chain or native Profile/OAuth support.

### Provider contracts

`Provider` retrieves copied credential snapshots with a context and must be concurrency
safe. NewStaticProvider copies explicit keys. A zero/nil StaticProvider returns
ErrMissingCredentials; an already-canceled context takes precedence. Formatting
redacts values, but exported fields remain sensitive and must not be logged directly.

EnvProvider reads ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET and optional
ALIBABA_CLOUD_SECURITY_TOKEN each time. All three absent returns ErrNotFound; any
incomplete configuration returns ErrMissingCredentials. NewChain copies an explicit
ordered provider list, skips only ErrNotFound and stops on other failures or invalid
snapshots. There is no implicit file, process, metadata or role discovery. Use
ProviderFunc for a custom source and preserve cancellation errors. See [cache contracts](credential-cache.md).

Issue [#53](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/53) verifies the
provider-only rule, no environment fallback, constructor/per-call rejection and
offline role signing/cache reuse. Synthetic Examples and CI do not establish live
role renewal or overall Beta acceptance. Separate [live STS evidence #55](live-sts-renewal.md)
now records real issuance/reuse, forced refresh, natural expiry renewal and cleanup
with a dedicated minimal-permission role; it does not add native Profile/OAuth support.

## 中文

### 优先采用 STS 的应用配置

应用客户端优先采用可刷新的 stscreds.AssumeRoleProvider，并包装 credentials.Cache。
完整生成客户端组合见[STS 指南](sts-credentials.md)及可执行 README/ExampleAssumeRoleProvider。
STS 配置独立来源以避免递归；Cache 合并刷新，复制到 StaticProvider 的 STS token 不会自动续期。

alicloud.Config 和服务 Options 只提供 CredentialsProvider，不提供裸 AccessKey/KeySecret/
SecurityToken 字段。即便环境凭据已经存在，也必须显式注入 provider。Nil/typed-nil provider
（含 nil ProviderFunc）在运行时/chain/cache 构造期间失败，不访问 HTTP、不读取凭据。
操作级配置替换执行同样校验，且不修改客户端。

| 来源            | 显式声明                                                                                         | 使用定位                                        |
| --------------- | ------------------------------------------------------------------------------------------------ | ----------------------------------------------- |
| 可刷新角色      | stscreds.NewAssumeRoleProviderFromClient(...)，随后 credentials.NewCache(...)                    | 应用主要路径：生成 STS→role provider→生成消费者 |
| 长期 AK/SK      | credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: id, AccessKeySecret: secret}) | 次要、明确选择；例如 STS 的授权来源             |
| 静态临时快照    | NewStaticProvider 加 SecurityToken 和 ExpiresAt                                                  | 已签发临时凭据，检查过期但不刷新                |
| 环境            | credentials.EnvProvider{}                                                                        | 显式读取环境，不自动注册                        |
| 自定义 provider | credentials.Provider 实现或 ProviderFunc                                                         | 应用管理来源，遵守 context/并发/错误契约        |

所有有意注入的自定义 provider 均可用，也可返回长期密钥。STS 优先是使用指南和 provider-only
配置规则，不是运行时强制仅接受 STS。AWS Go SDK v2 同样接受 provider，并记录显式
[StaticCredentialsProvider](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html)。
本 SDK 不采用 AWS LoadDefaultConfig 的隐式默认来源发现，不宣称默认凭据链或原生 Profile/OAuth 支持。

### Provider 契约

Provider 使用 context 返回复制的凭据快照，必须并发安全；NewStaticProvider 复制显式密钥。
零值/nil StaticProvider 返回 ErrMissingCredentials，已取消的 context 优先。格式化隐藏值，
但导出字段仍敏感，不得直接记录。

EnvProvider 每次读取 ALIBABA_CLOUD_ACCESS_KEY_ID、ALIBABA_CLOUD_ACCESS_KEY_SECRET 和可选
ALIBABA_CLOUD_SECURITY_TOKEN。三项都不存在返回 ErrNotFound，不完整配置返回 ErrMissingCredentials。
NewChain 复制显式有序列表，只跳过 ErrNotFound，其他失败/无效快照立即终止。不会隐式发现
文件、进程、metadata 或角色；ProviderFunc 自定义来源需保留取消错误。见[缓存契约](credential-cache.md)。

[#53](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/53) 验证 provider-only 配置、无环境回退、
构造/操作覆盖拒绝，以及离线角色签名/缓存复用。合成 Example 和 CI 不证明真实角色续期或整体 Beta 验收。
独立的[真实 STS 证据 #55](live-sts-renewal.md)现记录最小权限专用角色的真实签发/复用、强制刷新、
自然到期续期和清理，不新增原生 Profile/OAuth 支持。

## Anonymous RPC / 匿名 RPC

### English

Explicit `credentials.AnonymousProvider{}` is a marker for reviewed anonymous STS OIDC/SAML operations. Those operations never retrieve even a configured source provider. Signed operations reject the marker; nil/typed-nil remain invalid everywhere. No implicit key discovery is added. See [protocol contracts](sts-anonymous-rpc.md).

### 中文

显式 `credentials.AnonymousProvider{}` 是已审核 STS OIDC/SAML 匿名操作的标记，这些操作不读取任何已配置来源 provider。签名操作拒绝该标记；所有位置的 nil/typed-nil 仍无效，不引入隐式密钥发现。见[协议契约](sts-anonymous-rpc.md)。
