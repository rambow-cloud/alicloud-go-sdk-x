# 外部临时凭据

[English](external-credentials.md)

- 对应 issue #90；运行时仍只依赖标准库。
- 构造时仅校验配置，不发 HTTP 请求、不启动进程，也不读取 token。
- 可通过 `config.WithCredentialsProvider` 显式注册，也可以使用下表中的原生 Profile 模式。
- 直接使用 Provider 时，用 `credentials.Cache` 包装；默认配置加载和 Profile 已自带缓存。
- 并发调用共用一次有超时的刷新；取消一个等待者不会影响其他等待者。

## 支持的来源

| Provider | 原生 Profile | 获取与续期方式 |
| --- | --- | --- |
| `externalcreds.URIProvider` | CredentialsURI：`credentials_uri` | 对配置的 HTTP(S) 凭据服务发起 GET，要求完整 STS 字段和未来到期时间。 |
| `externalcreds.ProcessProvider` | External：`process_command` | 无 shell 地执行复制后的参数列表，接受 CLI 原生 StsToken 输出结构。 |
| `externalcreds.ECSMetadataProvider` | EcsRamRole：可选 `ram_role_name` | 固定实例元数据地址；先获取 IMDSv2 token，必要时发现唯一角色，再获取 STS。 |
| `stscreds.AssumeRoleWithOIDCProvider` | OIDC：`ram_role_arn`、`oidc_provider_arn`、`oidc_token_file` | 每次刷新重新读取 token 文件，调用生成的匿名 STS API。 |
| `externalcreds.CloudSSOProvider` | CloudSSO：CLI 原生登录、账号和访问配置字段 | 复用有效的原生 STS，或用当前门户访问令牌换取凭据。 |

## 默认发现

- 顺序为：显式 Provider → 显式 Profile → 临时环境凭据 → OIDC 环境配置 → URI 环境配置 → 自动选择的 Profile → 默认文件不存在时使用 ECS IMDSv2。
- URI 环境变量为 `ALIBABA_CLOUD_CREDENTIALS_URI`；元数据角色可通过 `ALIBABA_CLOUD_ECS_METADATA` 指定。
- `ALIBABA_CLOUD_DISABLE_EXTERNAL_PROCESS=true` 或 `1` 会拒绝自动发现的 URI/External 来源。
- `ALIBABA_CLOUD_ECS_METADATA_DISABLED=true` 或 `1` 会拒绝自动发现的实例元数据来源。
- 显式注册 Provider 表示主动启用，不读取上述发现开关。
- 显式文件或 Profile 不存在、所选来源无效、环境配置不完整时直接报错，不继续尝试实例元数据。
- 加载配置不会探测本机元数据，也不会自行补默认地域；元数据请求在实际获取凭据时才发生。

## 边界与数据归属

- 外部获取默认总超时为 5 秒，响应或标准输出上限为 1 MiB；显式 Provider 可调整，上限为 16 MiB。
- 元数据请求共用一次超时；申请的会话 token TTL 为 21,600 秒。
- 不降级到 IMDSv1。实例元数据使用复制后的原生 HTTP transport，并关闭代理。
- 传入的 `http.Client` 会复制，并禁用重定向；自定义客户端需遵守 context、禁止重定向，不得通过代理转发元数据 token。
- 默认错误不含 URL、进程参数、标准输出、标准错误或响应体；可用 `errors.Is` 判断取消和超时。
- 标量配置和进程参数会复制；token 提供者、HTTP transport 由调用方共享，必须支持并发。
- URI 和可执行程序必须可信。为兼容本地或私网凭据服务允许 HTTP；远程服务应使用 HTTPS。
- 进程无标准输入，标准错误会丢弃，标准输出受限。超时会终止直接子进程，继承输出管道的等待也有上限；SDK 不管理整个进程树。
- `ParseCommand` 支持引号参数和 Windows 路径，不解释 shell 语法，也不展开环境变量。
- 进程必须输出 CLI 原生 `StsToken` Profile，`sts_expiration` 使用 Unix 秒；不允许输出引导递归 Profile 或进程发现。只有显式设置 `ProcessOptions.AllowLongLived` 或 Profile `Options.AllowLongLived` 才接受 AK，默认拒绝。

## CloudSSO 会话

- 首次通过 CLI 的 CloudSSO 配置流程登录；SDK 不启动浏览器。
- 使用原生字段：`cloud_sso_sign_in_url`、`cloud_sso_account_id`、`cloud_sso_access_config`、`access_token`、`cloud_sso_access_token_expire`。
- 向门户 POST `/cloud-credentials`，使用 bearer 访问令牌，JSON 包含 `AccountId` 和 `AccessConfigurationId`。
- 接受完整的 `CloudCredential` 包装或旧版顶层凭据结构；拒绝歧义包装、失败状态码以及不一致的到期字段。
- Profile 续期会重读登录状态；账号、门户或访问配置改变时返回 `ErrConfigurationChanged`。
- 原生 STS 仍有效时可以继续复用，即使登录令牌已到期；需要续期时若登录令牌过期，返回 `ErrLoginRequired`，要求重新登录 CLI。
- 新 STS 只保存在内存缓存中；SDK 不写入或刷新 CloudSSO 配置及登录令牌。OAuth 的持久化规则单独保留。

## 验证与协议来源

- 离线测试覆盖准确的请求顺序和请求头、不降级 IMDSv1、代理隔离、延迟获取、缓存复用、token 重读、CloudSSO 配置变化及登录过期、输出限制、重定向、取消、参数复制和 AK 显式启用。
- 包内提供确定性 Example，注入模拟凭据服务，不需要网络或账号。
- 本次没有使用真实 IMDS、URI、External 或 CloudSSO 账号；真实联邦身份验收另见 #94。
- 原生字段与 CloudSSO 门户协议来自 CLI 提交 `fa14dd7b0359b5be229f9d770a662a86e69c13c1` 的[配置定义](https://github.com/aliyun/aliyun-cli/blob/fa14dd7b0359b5be229f9d770a662a86e69c13c1/config/profile.go)和[CloudSSO 刷新流程](https://github.com/aliyun/aliyun-cli/blob/fa14dd7b0359b5be229f9d770a662a86e69c13c1/cloudsso/refresh.go)。
- URI/IMDS 交叉验证来源为 credentials-go v1.4.5 的 [URI 实现](https://github.com/aliyun/credentials-go/blob/v1.4.5/credentials/providers/uri.go)和 [ECS 实现](https://github.com/aliyun/credentials-go/blob/v1.4.5/credentials/providers/ecs_ram_role.go)。
- 另见[默认配置](default-configuration.zh-CN.md)与[联邦身份凭据](federation-credentials.zh-CN.md)。
