# 完整 DSL STS 凭据

[English](sts-credentials.md)

- [#51](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/51) 实现 [AC-05 / UX-04](product-acceptance.zh-CN.md) 离线组合。
- NewAssumeRoleProviderFromClient 接受 service/sts 窄接口和完整原生输入/options，直接组合 credentials.Cache 与生成客户端，无需响应转换。
- 旧构造器继续保留 services/sts 参考桥调用方式。
- [可复用性评审 #72](sts-reuse-review.zh-CN.md) 将共享校验规则与请求模型分离，两个构造器采用同一套返回凭据校验。

### 生成与手写组件的边界

- `service/sts` 的操作、模型和协议绑定来自官方 Darabonba DSL，经官方语义解析器和完整 IR 生成。
- 这里的手写适配器把生成的响应转换为 `credentials.Credentials`，不另外实现签名或 HTTP 协议。
- 共享内部角色校验不依赖旧版请求类型；原生输入直接传给生成的 API。
- 缓存和原生 Profile 的角色配置复用同一适配器。来源发现使用 [LoadDefaultConfig](default-configuration.zh-CN.md)，适配器自身不执行发现。
- 旧构造器的公开签名仍需导入 `services/sts`；移除签名需要单独安排迁移。

### 组合

- 这是应用凭据的主要使用路径。
- README 与外部 ExampleAssumeRoleProvider 提供完整离线 STS→ECS 程序，包含角色签名断言与缓存复用。
- 长期来源为可选项，必须显式调用 credentials.NewStaticProvider；环境来源必须显式声明 credentials.EnvProvider{}，仍支持自定义来源。
- Config/Options 不接受裸密钥、不发现回退、不在构造期间读取凭据；发送请求前拒绝 nil/带类型的 nil 凭据提供者。

- 下方代码是函数体片段，应用传入 ctx 与独立 source 凭据提供者，导入根包 alicloud、credentials、 feature/stscreds、service/sts、service/ecs。
- 按 STS client→角色凭据提供者→cache→生成 ECS client 组合，仅授权真实使用时替换示例 ARN。
- 外部 ExampleNewAssumeRoleProviderFromClient 采用脚本 HTTP/合成凭据，无账号/网络；消费者集成测试覆盖两个生成客户端及过期轮换。

### 契约

- 构造/每次读取复制指针和 options 注册，调用方/API/切片修改不会延续。
- 两个构造器都会复制每次调用的选项切片；自定义 API 修改切片，也不会影响后续调用。
- API/callback 共享且需并发安全，构造期间不修改输入、callback 不保留对象。
- Provider 支持并发，不自动缓存，格式化脱敏，nil/零值及非法构造参数（含带类型的 nil API/nil callback）返回错误。

- 只复用既有审核辅助组件的角色/会话、identity、policy JSON、时长校验；发送完整 DSL 快照， 不转换旧请求/响应。
- Nil duration 缺省，显式小于 900（含零）失败，其他存在语义/int64 保留。
- 授权、角色最大时长及其他规则由服务决定，不修改生成操作校验/策略。

- Nil 响应/credentials 失败，空白密钥/token 返回 ErrMissingCredentials，缺失/空/过期时间为 ErrExpired，非法时间诊断不带值。
- 按 RFC3339 解析 UTC，要求未来时间，Source 为 sts.AssumeRole。
- 保留 errors.Is/As 和调用前后取消/超时，不记录凭据/原始 body。

- STS 使用独立 source 以避免递归，来源不变。
- 通过 credentials.NewCache 按[缓存契约](credential-cache.zh-CN.md) 限制刷新时间、合并并发请求，并在到期前刷新，某个等待者取消不取消共享刷新。
- AssumeRole 配置 Standard 仍不重试。
- #51 不加入 Profile/OAuth 发现、新依赖、生成文件修改，也未执行真实角色/刷新或云写调用；独立 UX-04 和 Beta/发布仍待验收。

- 独立[真实续期记录 #55](live-sts-renewal.zh-CN.md)现证明限定真实签发/复用、强制 Invalidate、等待真正 900 秒到期后的自动续期、生成 ECS 读取及全部临时 IAM 清理。
- 不替代独立 UX-04 或整体 Beta/发布验收，也不证明真实后台/并发刷新或原生 Profile/OAuth 续期。

### 证据

- 本地门禁为 doccheck、vet、全 Go 测试/Example、product-check、格式；Go 前执行 Node 前端契约。
- 测试覆盖输入/响应、缺省/int64、并发快照、结构化/取消、cache 取消隔离、STS 来源签名→角色 cache→生成 ECS 签名、虚拟时钟过期轮换、503 不重试。
- Linux race/Windows 由 CI 独立记录， 等待 CI 不算通过；未来合成时间不证明真实刷新。

## 可运行命令与示例

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
