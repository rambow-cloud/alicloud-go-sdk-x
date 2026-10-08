# Full-DSL STS credentials / 完整 DSL STS 凭据

[English](#english) | [中文](#中文)

## English

Issue [#51](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/51) implements the
offline composition part of [AC-05 / UX-04](product-acceptance.md).
`stscreds.NewAssumeRoleProviderFromClient` accepts `service/sts.AssumeRoleAPI` and
complete native input/options. The returned provider composes with credentials.Cache
and generated clients without a response translator. The existing
NewAssumeRoleProvider keeps its services/sts reference-bridge convention.

### Composition

This is the primary application credential path. The README and external
ExampleAssumeRoleProvider provide a complete offline STS-to-ECS program, including
role signing assertions and cache reuse. A long-lived source is optional and must
be explicitly constructed with credentials.NewStaticProvider; an environment source
must be explicitly declared as credentials.EnvProvider{}. Custom sources remain
supported. Config/Options never accept bare keys, discover a fallback or retrieve
credentials during construction; nil/typed-nil providers fail before requests.

```go
// Function body: the application supplies ctx and a separate source provider.
api, err := sts.NewFromConfig(alicloud.Config{
    Region: "cn-hangzhou", CredentialsProvider: source,
})
if err != nil { return err }
roleARN, session := "acs:ram::123456789012:role/example", "application"
provider, err := stscreds.NewAssumeRoleProviderFromClient(api, sts.AssumeRoleInput{
    RoleARN: &roleARN, RoleSessionName: &session,
})
if err != nil { return err }
cache, err := credentials.NewCache(provider, credentials.CacheOptions{})
if err != nil { return err }
client, err := ecs.NewFromConfig(alicloud.Config{
    Region: "cn-hangzhou", CredentialsProvider: cache,
})
if err != nil { return err }
_, err = client.DescribeRegions(ctx, nil)
return err
```

Import the root module as alicloud, credentials, feature/stscreds, service/sts and
service/ecs. Replace the placeholder ARN only for authorized live use. Runnable
external ExampleNewAssumeRoleProviderFromClient uses scripted HTTP and synthetic
credentials with no account/network; the generated-consumer integration test covers
both generated clients and expiry-driven rotation.

### Contracts

Construction and each retrieval copy input pointers and option registrations.
Caller/API changes and option-slice mutations do not persist. API/callback objects
remain shared and must be concurrency safe; do not mutate during construction or
retain callback objects. The provider is concurrency safe, has no automatic cache,
redacts formatting and returns errors for nil/zero provider or invalid constructor
arguments (including typed-nil API and nil callbacks).

Reuse existing reviewed helper role/session, identity, policy JSON and duration
validation only. Send the full-DSL snapshot without a legacy request/response
conversion. Nil duration stays absent; explicit duration below 900 including zero
fails. Other optional presence and int64 width are retained. Authorization, role
maximum lifetime and additional rules remain server decisions; generated STS
operation validation/policy is unchanged.

Nil output/credential containers fail. Keys/token must be nonblank, otherwise
ErrMissingCredentials. Missing/empty/expired expiration returns ErrExpired; malformed
expiration has a bounded diagnostic without its value. Parse RFC3339 to UTC, require
future expiry and return Source=sts.AssumeRole. API errors retain errors.Is/As identity;
check cancellation/deadlines before and after the call. Never log credentials/raw bodies.

Use separate source credentials for STS to avoid recursive retrieval. The source is
not altered. Wrap with credentials.NewCache for bounded/shared early refresh per
[cache contracts](credential-cache.md); canceling a caller does not cancel the shared
refresh. AssumeRole remains non-retrying even with Standard. No Profile/OAuth discovery,
new dependency, generated-file edit, live role execution or cloud write is included.
#51 itself did not exercise live role/refresh. Independent UX-04 and overall
Beta/release acceptance remain open.

The separate [live renewal record #55](live-sts-renewal.md) now establishes scoped real
issuance/reuse, forced Invalidate and automatic renewal after genuine 900-second
expiration with generated ECS reads and full temporary-IAM cleanup. It does not
replace independent UX-04 or overall Beta/release acceptance, and does not establish
live background/concurrent refresh or native Profile/OAuth renewal.

### Evidence

Local gates: doccheck, vet, full Go tests/Examples, product-check and formatting; Node
frontend contracts precede Go checks. Tests cover input/response boundaries, omission/
int64, concurrent snapshots, structured/cancellation errors, cache cancellation
isolation, signed source STS -> role cache -> signed generated ECS requests, virtual
clock expiry rotation and 503 non-retry. Linux race/Windows evidence is recorded in CI
separately; pending CI is not PASS. Future synthetic times do not prove live renewal.

## 中文

[#51](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/51) 实现
[AC-05 / UX-04](product-acceptance.md) 离线组合。NewAssumeRoleProviderFromClient 接受
service/sts 窄接口和完整原生输入/options，直接组合 credentials.Cache 与生成客户端，无需
响应转换。旧构造器继续保留 services/sts 参考桥调用方式。

### 组合

这是应用凭据的主要使用路径。README 与外部 ExampleAssumeRoleProvider 提供完整离线
STS→ECS 程序，包含角色签名断言与缓存复用。长期来源为可选项，必须显式调用
credentials.NewStaticProvider；环境来源必须显式声明 credentials.EnvProvider{}，仍支持
自定义来源。Config/Options 不接受裸密钥、不发现回退、不在构造期间读取凭据；发送请求前
拒绝 nil/typed-nil provider。

英文代码为函数体，应用传入 ctx 与独立 source provider，导入根包 alicloud、credentials、
feature/stscreds、service/sts、service/ecs。按 STS client→role provider→cache→生成 ECS
client 组合，仅授权真实使用时替换示例 ARN。外部 ExampleNewAssumeRoleProviderFromClient
采用脚本 HTTP/合成凭据，无账号/网络；消费者集成测试覆盖两个生成客户端及过期轮换。

### 契约

构造/每次读取复制指针和 options 注册，调用方/API/切片修改不会延续。API/callback 共享且需
并发安全，构造期间不修改输入、callback 不保留对象。Provider 支持并发，不自动缓存，格式化
脱敏，nil/零值及非法构造参数（含 typed-nil API/nil callback）返回错误。

只复用既有审核 helper 的角色/会话、identity、policy JSON、时长校验；发送完整 DSL 快照，
不转换旧请求/响应。Nil duration 缺省，显式小于 900（含零）失败，其他存在语义/int64 保留。
授权、角色最大时长及其他规则由服务决定，不修改生成操作校验/策略。

Nil 响应/credentials 失败，空白密钥/token 返回 ErrMissingCredentials，缺失/空/过期时间为
ErrExpired，非法时间诊断不带值。按 RFC3339 解析 UTC，要求未来时间，Source 为 sts.AssumeRole。
保留 errors.Is/As 和调用前后取消/超时，不记录凭据/原始 body。

STS 使用独立 source 以避免递归，来源不变。通过 credentials.NewCache 按[缓存契约](credential-cache.md)
有界/合并/提前刷新，某个等待者取消不取消共享刷新。AssumeRole 配置 Standard 仍不重试。
#51 不加入 Profile/OAuth 发现、新依赖、生成文件修改，也未执行真实角色/刷新或云写调用；独立
UX-04 和 Beta/发布仍待验收。

独立[真实续期记录 #55](live-sts-renewal.md)现证明限定真实签发/复用、强制 Invalidate、等待
真正 900 秒到期后的自动续期、生成 ECS 读取及全部临时 IAM 清理。不替代独立 UX-04 或整体
Beta/发布验收，也不证明真实后台/并发刷新或原生 Profile/OAuth 续期。

### 证据

本地门禁为 doccheck、vet、全 Go 测试/Example、product-check、格式；Go 前执行 Node 前端契约。
测试覆盖输入/响应、缺省/int64、并发快照、结构化/取消、cache 取消隔离、STS 来源签名→角色
cache→生成 ECS 签名、虚拟时钟过期轮换、503 不重试。Linux race/Windows 由 CI 独立记录，
等待 CI 不算通过；未来合成时间不证明真实刷新。
