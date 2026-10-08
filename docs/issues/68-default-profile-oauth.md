## English

### Problem and evidence

The user rejects the existing no-default-discovery/no-native-Profile/OAuth limitation.
Manual CLI snapshots are insufficient for a modern local developer experience.
The earlier #53 deliberate AK/SK opt-in constraint was overextended into a blanket
discovery exclusion. Authoritative revised route: docs/default-configuration.md.
Pinned official CLI fa14dd7b0359b5be229f9d770a662a86e69c13c1 documents the native
profile fields, CN/INTL client IDs, OAuth refresh and STS exchange protocols.

### Scope and dependencies

Member of #57 and milestone v0.1.0; prerequisite for updated #60 consumer acceptance
and #61 publication. Preserve accepted #51/#53/#55/#58/#59 behavior and historical
evidence. Introduce stdlib-only config.LoadDefaultConfig and feature/profilecreds,
native CLI JSON profiles, OAuth token refresh/STS exchange, StsToken, explicit-opt-in
AK/RamRoleArn, ChainableRamRoleArn, copied configuration and shared bounded caches.
Explicit provider -> explicit profile -> complete temporary environment -> selected
default profile. Region precedence and sanitized errors are documented before code.
No generator/source edits, new dependency, implicit long-lived keys, subprocesses,
browser launch, unrelated CLI configuration writes, arbitrary credential URIs or IAM/IdP creation.
Other CLI modes return explicit unsupported-mode errors and are not claimed.

### Acceptance

- [ ] LoadDefaultConfig(ctx, options...) returns existing Config consumable by generated STS/ECS; service constructors retain strict provider-only configuration.
- [ ] Provider/profile/environment/region precedence, copied state, missing vs invalid sources, typed-nil, strict bounded JSON and duplicate-profile/cycle failures are deterministic and tested.
- [ ] Default loading supports OAuth/StsToken and rejects implicit long-lived keys; explicitly constructed profile/EnvProvider/StaticProvider overrides remain available. Role profiles use generated AssumeRole/helper/cache without translation or refresh loops.
- [ ] OAuth reuses valid STS snapshots, refreshes expired access tokens and exchanges native bearer credentials; rotated sessions persist atomically before exchange and survive reconstruction, unknown/sibling configuration is retained, bounded file locks serialize SDK processes, detected CLI edits fail safely; concurrent callers share bounded renewal, cancellation remains recognizable and default errors/formatting reveal no secrets. Initial login/revocation and shared-file ownership are documented.
- [ ] New public packages have complete English Go comments/doc.go and deterministic external Examples; paired guides and AGENTS/development/acceptance/release routes supersede stale exclusions together.
- [ ] doccheck/vet/tests/format and exact-head Linux race/Windows CI pass; an authorized read-only native local-profile identity check records sanitized results. Record actual live refresh separately from offline tests and prior role-renewal evidence; missing live credentials remain an explicit gate rather than invented PASS.

### Verification

Offline temporary profile files and injected HTTP, no accounts/network in tests.
Source-bound independent token/exchange/role fixtures, concurrency/cancellation,
negative status/body/expiry/security tests, SDK documentation gates and existing CI.
Read-only local identity and OAuth exchange are within prior local-profile authorization;
never print/cache real credentials in committed artifacts. Keep #60/#61 open until
new behavior and independent human acceptance/publication/indexing gates are met.

## 中文

用户要求消除默认加载与原生 Profile/OAuth 缺口；手动 CLI 快照不足以形成现代开发体验。
#53 长期 AK/SK 显式启用不应扩大为全面禁止发现，路线以 docs/default-configuration.md
为准，固定官方 CLI 提交提供字段及 CN/INTL 刷新/交换证据。

归属 #57/v0.1.0，为更新后 #60/#61 前提，保留既有基础/STS/真实续期证据。新增标准库
config.LoadDefaultConfig、feature/profilecreds、原生 CLI Profile、OAuth 刷新与 STS
交换、StsToken、显式启用 AK/RamRoleArn、ChainableRamRoleArn，明确来源/地域优先级、
配置副本与共享有界缓存。不改生成/来源、不新增依赖、长期密钥不默认启用、不启动 CLI/
浏览器、不改 CLI 无关配置、不自动访问任意 URI、不建 IAM/IdP；其他模式明确报未支持。

逐项验收对应英文清单：返回生成客户端可用 Config；优先级/副本/缺失无效/typed-nil/
严格有界 JSON/重复与循环；默认 OAuth/StsToken 和显式密钥 provider；角色复用生成 helper/
cache；OAuth STS 复用、access token 刷新、交换、轮换、交换前原子持久化及重建后复用、
未知/其他配置保留、SDK 文件锁及外部修改拒绝、并发/取消/有界/脱敏；
英文 pkg.go.dev/离线 Example、双语指南和 AGENTS/开发/验收/发布路线共同修正；文档/
vet/tests/格式及准确提交 Linux race/Windows，授权本地原生 Profile 身份只读记录。
真实 OAuth 刷新与 fixture/#55 证据分开，缺凭据记未完成而不虚构通过。

验证用临时虚构配置和注入 HTTP，无真实账号/网络单测；沿用现有 CI。已有授权覆盖本地
Profile 只读身份/OAuth 交换，但不输出或提交真实凭据；独立人员及发布/索引未满足前
#60/#61 保持开放。

### Live protocol decision / 真实协议决策

CN live exchange returns PascalCase fields while the pinned official CLI uses camelCase tags and legacy case-folding JSON. JSON v2 decoding now accepts only the two explicit wire variants and rejects ambiguous, mixed, duplicate or unreviewed casing. The authorized native run passed generated identity reads, forced exchange, persistence/reconstruction and unrelated-config preservation at 656ce39dda0b89ae743645ba1328974a937dc780. Actual live refresh-token rotation remains NOT RUN because the new login access token was valid. See docs/acceptance/profile-oauth-live.json.

中文：真实 CN exchange 为 PascalCase，固定 CLI 标签为 camelCase，旧 JSON 自动忽略大小写；JSON v2 仅兼容两个明确结构，拒绝歧义/混用/重复/未审核大小写。上述实现提交真实原生身份读取、主动交换、持久化/重建及无关配置保留通过；刚登录 token 有效，真实 refresh-token 轮换仍未跑，见脱敏记录。
