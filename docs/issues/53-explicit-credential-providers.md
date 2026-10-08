# Explicit credential providers / 显式凭据来源

## English

## Problem and evidence

The user requires STS-first usage and long-lived access keys to be secondary, available only through explicitly declared providers. Config already exposes only CredentialsProvider, and sources are explicit; however the README first-call path uses long-lived keys directly, and root NewClient, NewChain and NewCache accept typed-nil provider pointers. A nil ProviderFunc also passes root construction. These values can panic during signing/refresh rather than fail before requests.

AWS Go SDK v2 documents provider injection and an explicit StaticCredentialsProvider for hard-coded keys: https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html . AWS permits custom providers and automatic discovery in its default loader; this SDK intentionally has no implicit discovery. This scope does not impose an STS-only runtime or reject deliberate custom providers.

## Scope and dependencies

Depends on merged #49 and #51. Follow the updated docs/development-path.md and AC-01/05/11, UX-01/04. Make STS provider/cache primary in the README and a complete deterministic external Example. Document static/env providers as explicit secondary paths. Reject nil/typed-nil sources at runtime, chain and cache construction, including operation configuration replacement; do not retrieve credentials during construction. Preserve StaticProvider, ProviderFunc and exact Alibaba wire names.

No new public credentials wrapper, generator changes, dependencies, implicit Profile/environment discovery, live calls or STS-only enforcement.

## Acceptance criteria

- [ ] Config/service Options contain no bare access-key/secret/token fields; a populated environment never substitutes for an omitted provider.
- [ ] Runtime and generated NewFromConfig reject nil/typed-nil/nil-function sources before HTTP or retrieval; operation replacement preserves output and OperationError wrapping on rejection.
- [ ] Chain/cache reject typed-nil sources; explicit static/env/custom and STS providers remain usable, context/error/redaction contracts remain intact.
- [ ] A runnable external Example and equivalent README snippets compose explicit source -> generated STS -> role provider/cache -> generated ECS and verify role signing with scripted HTTP.
- [ ] Paired docs, AGENTS.md and acceptance scope state STS-first guidance, explicit long-lived opt-in and deliberate differences from AWS, without claiming live/Beta acceptance.
- [ ] Formatting, bilingual/doccheck, vet, full Go tests/Examples and Linux race/Windows CI pass; no generated outputs change.

## Verification plan

Use synthetic environment keys, rejecting providers that panic if called during construction, scripted transport call counts/signing assertions, reflection checks of credential configuration shape, existing rotation/concurrency contracts and runnable Example output. Execute local Go gates once; CI covers Linux race, Windows and unchanged generation/frontend checks.

## 中文

按用户要求优先展示 STS provider/cache，长期 AK/SK 必须显式注册 StaticProvider，环境来源必须显式注册 EnvProvider，保留自定义 provider。当前入口已只有 provider，但 README 默认展示长期密钥，且 typed-nil 来源可能在请求/刷新时崩溃。依赖已合并 #49/#51，先更新路线，再在独立分支补校验、运行时/生成客户端/操作覆盖测试、完整离线 STS→ECS Example、配对文档和约束。无新依赖/生成变更/隐式发现/真实调用，不施加 STS-only 限制；通过本地门禁和 Linux race/Windows CI后仅验收本范围，不宣称 Beta。

### 对应验收与验证

- Config/服务 Options 不含裸 access-key/secret/token 字段；完整环境凭据不能替代缺失的 provider。
- runtime/生成 NewFromConfig 在 HTTP/读取前拒绝 nil、typed-nil、nil 函数；操作替换失败保留输出和 OperationError。
- chain/cache 拒绝 typed-nil 来源；显式 static/env/custom/STS 来源可用，保留取消、错误和脱敏契约。
- 可执行外部 Example 与等价 README 组合显式来源→生成 STS→role provider/cache→生成 ECS，并检查角色签名。
- 配对文档、AGENTS 和验收范围明确 STS 优先、长期密钥显式选择及与 AWS 的差异，不宣称真实/Beta。
- 格式、双语/doccheck、vet、全 Go 测试/Example 和 Linux race/Windows CI 通过，生成输出不变。

验证使用合成环境密钥、构造期间若读取即 panic 的来源、脚本 HTTP 次数/签名断言、配置字段反射、
既有轮换/并发契约和确定 Example 输出。本地 Go 门禁执行一次，CI 验证 race、Windows 及未改动的
生成/前端门禁。修改只在独立 issue 分支实施，无真实云调用。
