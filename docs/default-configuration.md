# Default configuration and native profiles / 默认配置与原生 Profile

## English

The user's 2026-10-08 correction makes default configuration loading and native
Alibaba Cloud CLI Profile/OAuth support required foundation capabilities before
v0.1.0 publication. It supersedes the earlier blanket no-discovery scope in #53,
without weakening provider-only service configuration or deliberate long-lived
key registration. Preserve the accepted STS/generated/runtime baseline.

### Route and public contract

Implement a standard-library-only `config` package exposing
`LoadDefaultConfig(ctx, optFns...)`, `WithRegion`, `WithCredentialsProvider`,
`WithSharedConfigProfile`, `WithSharedConfigFile` and `WithHTTPClient`. It returns
the existing alicloud.Config and supports all generated NewFromConfig constructors.
Direct service construction still rejects nil providers and does not discover sources.

After CLI login, this complete local-use program loads the named profile and reads
identity without printing it. It requires an authorized account/network; the package
Examples instead use temporary fictional files and injected offline HTTP.

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

Add `feature/profilecreds` for native `~/.aliyun/config.json` reading, named profile
selection and credential retrieval. Support OAuth, StsToken, AK, RamRoleArn and
ChainableRamRoleArn; AK or AK-based role sources require an explicitly constructed
profile provider with AllowLongLived enabled. Unsupported CLI modes fail with a typed,
bounded error rather than pretending to cover CloudSSO/process/URI/metadata sources.
OIDC/SAML remain supported native operations, not automatic token-file discovery here.

Credential precedence: explicit provider; explicitly selected profile; temporary
environment credentials; profile selected by ALIBABA_CLOUD_PROFILE, then CLI current,
then default. Temporary environment credentials require the complete ID/secret/token
triple. A present incomplete or long-lived environment source is an error, never
silently skipped. Explicit EnvProvider/StaticProvider remain supported overrides.
Region precedence: WithRegion, ALIBABA_CLOUD_REGION_ID, ALIBABA_CLOUD_REGION, selected
profile region. There is no invented default region. File override is explicit;
the default filename uses os.UserHomeDir. Missing configuration returns ErrNotFound;
malformed/duplicate JSON, duplicate profiles, invalid selected modes and source cycles
fail with sanitized errors. Bound configuration and credential HTTP reads.

Construct providers without retrieving credentials or making network requests.
Returned providers are cached, bounded and concurrency safe; explicit providers
retain their own contract. Separate source identity for role assumption, copy inputs,
and reuse the generated STS helper/cache instead of adding a translator/refresh loop.
Role session name defaults to alicloud-go-sdk-x; an unconfigured role duration stays absent.

### OAuth protocol and session ownership

Research pinned official CLI commit `fa14dd7b0359b5be229f9d770a662a86e69c13c1`:
[profile fields](https://github.com/aliyun/aliyun-cli/blob/fa14dd7b0359b5be229f9d770a662a86e69c13c1/config/profile.go),
[refresh/exchange](https://github.com/aliyun/aliyun-cli/blob/fa14dd7b0359b5be229f9d770a662a86e69c13c1/config/configure.go).
Use the CLI's CN/INTL OAuth endpoints and public client IDs. Reuse unexpired STS
credentials; when renewal is needed, refresh an expired OAuth access token with
POST /v1/token (form grant_type=refresh_token), then exchange its bearer token with
empty POST /v1/exchange for an expiring STS snapshot. Honor context, shared refresh
deadline, exact native fields, token rotation and strict JSON v2 validation.

Persist renewed OAuth/STS fields atomically in the selected native CLI profile so
rotating refresh tokens survive process restart and remain usable by the CLI. Serialize
SDK writers with a bounded same-file lock, reload credential fields under that lock,
preserve unknown/root/sibling configuration, and reject detected external edits before
replacement. Persist token rotation before the potentially failing exchange. No
process/browser is spawned. Initial interactive login remains
`aliyun configure --mode OAuth --profile <name>`. A revoked/expired login or server
invalid_grant returns errors.Is-compatible ErrLoginRequired with no token/body/URL
values. Configuration settings remain construction-time snapshots; credential renewal
reloads only the selected OAuth session fields. CLI writers do not honor the SDK lock;
concurrent CLI reconfiguration is unsupported and detected edits stop persistence.
Crash-left lock files cause a bounded error rather than unsafe lock stealing. Other
CLI modes need their own scoped issues, not silent claims.

### Acceptance and release impact

Live inspection after reauthentication found the CN exchange returns PascalCase
`AccessKeyId`, `AccessKeySecret`, `SecurityToken`, `Expiration`, while the pinned CLI
uses camelCase JSON tags. Its legacy JSON decoder folds case; JSON v2 is exact by
default. Accept these two explicit complete schemas, reject mixed or ambiguous
variants, and preserve strict JSON v2 decoding. This is a reviewed protocol
compatibility exception, not general case-insensitive decoding.

Implement route/docs and real issue before code. Require deterministic external
Examples and tests for precedence, explicit AK opt-in, missing/incomplete config,
typed-nil overrides, profile cycles, copied state, role composition, OAuth reuse,
refresh/exchange/token rotation, shared concurrent refresh/canceled caller isolation,
HTTP timeout/status/invalid JSON/expired responses and default secret redaction.
Run doccheck/vet/tests and formatting once for the meaningful change; Linux race and
Windows CI include both new packages. No dependency or generator-source changes.

An authorized local native Profile -> generated GetCallerIdentity read may establish
live integration without printing account identity or credentials. Live token refresh
must be recorded separately from offline fixtures and existing #55 role renewal;
never claim it passed without actual execution. No IAM/IdP resources are needed.
Revisit #60's human handoff against this new behavior before #61 publication; retain
historical accepted evidence and keep release/indexing gates open until satisfied.

Native live verification passed on 2026-10-08 against implementation
`656ce39dda0b89ae743645ba1328974a937dc780`: three generated identity reads, one
forced OAuth exchange, cache reuse, persisted-session reconstruction, authentication-
fields-only updates and released file lock. See [the sanitized report](acceptance/profile-oauth-live.json).
No CLI subprocess or cloud resource creation was involved. The fresh login access
token was still valid, so actual live refresh-token rotation and natural OAuth expiry
waiting remain NOT RUN; offline rotation fixtures and prior #55 role expiry are
separate evidence.

## 中文

用户于 2026-10-08 修正路线：默认配置加载和阿里云 CLI Profile/OAuth 原生支持是首版
v0.1.0 发布前必需基础能力，优先于 #53 的全面禁止发现范围，但仍保留服务配置只接受
provider、长期密钥显式启用和已接受 STS/生成/runtime 基线。

### 路线与公共契约

新增只依赖标准库的 `config` 包，提供 LoadDefaultConfig(ctx, optFns...) 及上述
WithRegion/WithCredentialsProvider/WithSharedConfigProfile/WithSharedConfigFile/
WithHTTPClient，返回既有 alicloud.Config，所有生成 NewFromConfig 可直接使用。
直接构造服务仍拒绝 nil provider，不自行发现来源。

新增 feature/profilecreds 原生读取 ~/.aliyun/config.json、选择命名 Profile 并获取
凭据，支持 OAuth、StsToken、AK、RamRoleArn、ChainableRamRoleArn。AK 或使用 AK 的
角色来源必须显式构造 AllowLongLived provider；未支持的 CloudSSO/process/URI/metadata
模式返回有界类型错误，不宣称覆盖；OIDC/SAML 已有原生操作，本范围不自动发现 token 文件。

凭据优先级：显式 provider → 显式选中的 Profile → 完整环境临时凭据 →
ALIBABA_CLOUD_PROFILE → CLI current → default。环境临时来源必须同时有 ID/secret/token，
已配置但不完整或仅有长期密钥返回错误，不悄悄跳过。仍可显式注入 EnvProvider/StaticProvider。
地区依次取 WithRegion、ALIBABA_CLOUD_REGION_ID、ALIBABA_CLOUD_REGION、选中 Profile，
不虚构默认地域。可显式覆盖文件，默认通过 os.UserHomeDir 定位；缺配置为 ErrNotFound，
非法/重复 JSON、重复 Profile、非法选中模式、来源循环均脱敏失败；限制文件与 HTTP 读取大小。

构造期间不读凭据、不联网，返回缓存、有界、并发安全的 provider；自定义来源遵守自身契约。
角色使用独立来源，复制输入，复用生成 STS helper/cache，不增加响应翻译或应用刷新循环。

英文完整本地使用程序在 CLI 登录后按指定 Profile 加载，再读取身份，只输出 HTTP 状态。
它需要已授权账号/网络；pkg.go.dev 的 Example 使用临时虚构文件和注入 HTTP，离线运行。
默认不提供地区时按前述优先级解析；角色会话名缺省为 alicloud-go-sdk-x，时长未配则省略。

### OAuth 协议与会话所有权

依据上述固定官方 CLI 提交的字段及刷新/交换实现，使用 CN/INTL endpoint 和公开 client ID。
复用未到期 STS；续期时通过 POST /v1/token 的 refresh_token 表单刷新已过期 OAuth access
token，再用 bearer token 的空 POST /v1/exchange 换取有明确过期时间的 STS。支持 context、
共享刷新 deadline、准确字段、token 轮换和严格 JSON v2。

更新后的 OAuth/STS 认证字段原子保存至选中 CLI Profile，使轮换 refresh token 在进程重启后
仍可用且与 CLI 共享。SDK 用有界同文件锁串行写入，锁内重读认证字段，保留根/未知/其他
Profile 配置，替换前发现外部修改则失败；token 刷新成功即持久化，不等可能失败的 exchange。
不启动进程或浏览器。
初次交互登录仍由 `aliyun configure --mode OAuth --profile <name>` 完成；登录失效、撤销
或 invalid_grant 返回可 errors.Is 判断的 ErrLoginRequired，不带 token/body/URL。
配置设置仍为构造快照，认证续期只重读选中 OAuth 会话。CLI 不遵守 SDK 锁，不支持同时
运行 CLI 重配，检测到修改时停止写入；崩溃残留锁有界报错，不自动抢锁。其他模式后续
单独 issue，不静默宣称已实现。

### 验收与发布影响

重新登录后的真实 CN exchange 返回 PascalCase 的 AccessKeyId、AccessKeySecret、
SecurityToken、Expiration，而固定 CLI 源码使用 camelCase 标签。其旧 JSON 解码器
自动忽略大小写，JSON v2 默认精确匹配。仅支持这两个明确且完整的结构，拒绝混用或
歧义，保留严格 JSON v2；这是已审核的协议兼容例外，不开启通用大小写忽略。

先落路线/文档与真实 issue，再开发。同步离线外部 Example 和行为测试，覆盖优先级、AK
显式启用、缺失/不完整/typed-nil、来源循环、输入副本、角色组合、OAuth 复用/刷新/交换/
轮换、并发共享与等待者取消隔离、HTTP deadline/status/非法 JSON/过期响应及默认脱敏。
相关文档/vet/tests/格式各运行一次，Linux race/Windows CI 覆盖新包；不新增依赖、不改生成来源。

既有授权允许本地原生 Profile→生成 GetCallerIdentity 只读验证，证据不带真实身份或凭据。
真实 OAuth 刷新、离线 fixture 与 #55 角色续期分别记录，未实际执行不能宣称通过，不需要
创建 IAM/IdP。#61 发布前让 #60 独立人员基于新增行为验收，保留历史证据，门禁满足前
发布/索引仍保持开放。

2026-10-08 原生真实验收已通过，实现提交为上述 SHA：3 次生成身份读取、1 次主动 OAuth
交换、缓存复用、持久会话重建、仅更新认证字段及锁释放，见[脱敏报告](acceptance/profile-oauth-live.json)。
未启动 CLI、未创建云资源。新登录 access token 尚未过期，因此真实 refresh-token 轮换
和等待自然 OAuth 到期仍记 NOT RUN；离线轮换 fixture 和 #55 角色到期为独立证据。
