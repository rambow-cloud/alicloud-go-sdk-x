# 默认配置与原生 Profile

[English](default-configuration.md)

## 快速开始

- 先执行 `aliyun configure --mode OAuth --profile oss-sftp`，完成一次交互登录。

- 使用 `config.LoadDefaultConfig` 读取配置，再把结果交给 `sts.NewFromConfig`。Profile 指阿里云 CLI 保存的一组命名配置。

- 下列程序会发起真实的只读身份查询，只输出 HTTP 状态。包内的 Example 则使用模拟响应，不需要账号或网络。

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/rambow-cloud/alicloud-go-sdk-x/config"
    "github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

func main() {
    ctx := context.Background()
    cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile("oss-sftp"))
    if err != nil { log.Fatal(err) }
    client, err := sts.NewFromConfig(cfg)
    if err != nil { log.Fatal(err) }
    output, err := client.GetCallerIdentity(ctx, nil)
    if err != nil { log.Fatal(err) }
    fmt.Println(output.Metadata.HTTPStatusCode)
}
```

## 配置选项

- `config` 和 `feature/profilecreds` 只依赖标准库。

- 加载选项包括 `WithRegion`、`WithCredentialsProvider`、`WithSharedConfigProfile`、`WithSharedConfigFile`、`WithHTTPClient` 和 `WithCredentialsCacheOptions`。

- 默认读取 `~/.aliyun/config.json`，用户目录通过 `os.UserHomeDir` 获取。

- 支持的 CLI 模式为 OAuth、StsToken、AK、RamRoleArn、ChainableRamRoleArn、OIDC、EcsRamRole、CredentialsURI、External 和 CloudSSO。

- 使用 AK 或以 AK 为来源的角色配置时，必须显式调用 `profilecreds.NewProvider` 并设置 `AllowLongLived: true`。也可以明确注入 StaticProvider、EnvProvider 或自定义凭据提供者。

- 外部来源的限制与原生 CloudSSO 会话规则见[外部临时凭据](external-credentials.zh-CN.md)。完整的 OIDC 环境配置会按 #89 加载可轮换的 token 文件；SAML 断言来源须显式注册，详见[联邦身份凭据](federation-credentials.zh-CN.md)。

- 服务构造函数仍要求提供凭据提供者。加载配置会读取本地 JSON 并建立缓存，但不会通过 HTTP 获取凭据。

## 来源优先级

1. 显式注入的凭据提供者。

2. 显式选择的 Profile。

3. 完整的临时环境凭据：ALIBABA_CLOUD_ACCESS_KEY_ID、ALIBABA_CLOUD_ACCESS_KEY_SECRET、ALIBABA_CLOUD_SECURITY_TOKEN。

4. 完整的 OIDC 环境配置：角色 ARN、OIDC provider ARN 与 token 文件名。

5. URI 环境变量 ALIBABA_CLOUD_CREDENTIALS_URI。

6. ALIBABA_CLOUD_PROFILE 指定的 Profile；未指定时依次使用 CLI 的 `current` 和 `default`。

- 环境变量已设置但不完整时，加载直接失败；长期环境密钥也必须显式启用。无效来源不会被悄悄跳过。

- 地域优先级为 WithRegion → ALIBABA_CLOUD_REGION_ID → ALIBABA_CLOUD_REGION → 所选 Profile。SDK 不自行指定默认地域。

- 默认文件不存在时，最后尝试延迟获取的 ECS IMDSv2；发现开关及显式来源的处理规则见[外部临时凭据](external-credentials.zh-CN.md)。

- 显式文件或 Profile 不存在时返回 `credentials.ErrNotFound`。非法 JSON、重复 Profile 和角色来源循环会报错，错误内容不含敏感信息。配置文件和 OAuth 响应最多读取 1 MiB。

## 凭据续期与文件更新

- STS 凭据仍有效时直接复用。OAuth 访问令牌过期后，通过 POST /v1/token 刷新，再通过 POST /v1/exchange 换取 STS 凭据。

- 并发调用共用一次刷新，并设定超时。取消一个等待者不会影响其他等待者。

- 新的 OAuth 令牌在交换前保存；新的 STS 字段在交换成功后保存。进程重启后和 CLI 都能继续使用更新后的会话。

- 只更新所选 Profile 的认证字段，保留根配置、未知字段和其他 Profile。普通配置仍使用构造时的副本。

- SDK 写入使用受 context 限制的文件锁，并在同目录替换配置文件；检测到外部修改时停止保存。

- 续期期间不要同时运行 CLI 重配，CLI 不使用 SDK 的锁。崩溃留下的锁会导致刷新超时失败，SDK 不会自动抢锁。

- 首次登录仍需交互操作。登录撤销、过期或服务返回 invalid_grant 时，返回可用 errors.Is 判断的 `ErrLoginRequired`。默认错误不包含令牌、响应体或 URL；SDK 不启动浏览器；External Profile 仅在获取凭据时启动配置的进程。

- 角色配置复用生成的 STS 辅助组件和缓存，来源凭据与目标角色分开，输入会复制。会话名默认 alicloud-go-sdk-x；未设置的有效期参数不发送。

## 协议依据与验收

- 协议依据为官方 CLI 提交 `fa14dd7b0359b5be229f9d770a662a86e69c13c1` 的[配置字段](https://github.com/aliyun/aliyun-cli/blob/fa14dd7b0359b5be229f9d770a662a86e69c13c1/config/profile.go)和[刷新、交换流程](https://github.com/aliyun/aliyun-cli/blob/fa14dd7b0359b5be229f9d770a662a86e69c13c1/config/configure.go)。使用对应 CN/INTL 端点及公开 client ID。

- 真实 CN 响应使用 AccessKeyId、AccessKeySecret、SecurityToken、Expiration。固定 CLI 源码使用 camelCase 标签，旧解码器忽略大小写；本 SDK 的 JSON v2 仅兼容这两个完整结构，拒绝混用、重复和歧义。

- [原生真实验收](acceptance/profile-oauth-live.json)：2026-10-08、实现提交 `656ce39dda0b89ae743645ba1328974a937dc780`，3 次身份查询、1 次主动交换、缓存复用、仅认证字段保存、会话重建和锁释放均通过。未启动 CLI 子进程、未新建资源、未披露凭据或身份。

- 真实 refresh-token 轮换和等待 OAuth 自然到期尚未运行（NOT RUN），因为新登录的访问令牌仍有效。离线轮换测试与 [#55 角色续期](live-sts-renewal.zh-CN.md)是独立证据。

- 离线测试覆盖来源优先级、nil/带类型的 nil、AK 显式启用、来源循环、数据复制、角色组合、并发、取消、令牌轮换、持久化及非法 HTTP/JSON/过期响应。两个包均纳入文档检查、vet、测试和 Linux race/Windows CI。

- #68 已取代 #53 的旧禁止发现规则；#60 代理消费者验收已完成，独立人工体验保留为可选项 #76；#61 发布与索引仍未完成。见[消费者任务](sts-consumer-acceptance.zh-CN.md)。
