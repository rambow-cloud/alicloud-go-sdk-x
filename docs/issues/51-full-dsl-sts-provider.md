# Full-DSL STS credential provider composition

## English

### Affected areas

credentials, sts, testing

### Dependencies

#49 defines product AC-05/UX-04. Foundation/cache and full-DSL generation are already accepted; documentation PR will be the explicit stack base.

### Problem and evidence

feature/stscreds.NewAssumeRoleProvider imports services/sts and its typed operation interface/options. A service/sts.Client cannot satisfy this older interface, even though it provides the complete generated AssumeRole API. Consumers must manually translate the new response into credentials.Provider, defeating AC-05/UX-04.

### Scope

Add NewAssumeRoleProviderFromClient for service/sts.AssumeRoleAPI, keeping the existing constructor and returned AssumeRoleProvider contract. Copy input pointers and option registrations at construction and each retrieval. Validate required role/session fields and existing reviewed helper rules without changing generated operation validation. Convert the exact full-DSL response to a credential snapshot, with valid nonempty keys/token and future RFC3339 expiration. Reuse credentials.Cache; preserve source credentials, structured errors and cancellation; sanitize invalid expiration diagnostics. Do not edit generated files, add dependencies or infer retry safety.

### Acceptance criteria

- [ ] New full-DSL constructor, small mock seam, redacted formatting and zero/nil safety; old constructor/examples preserved.
- [ ] Pointer presence/optional fields and duration width retained; constructor/API/callback mutations do not persist; concurrent retrieval supported with safe extensions.
- [ ] Reject nil/typed-nil API, nil options, invalid required/helper fields, missing/blank keys/token and missing/malformed/expired expiration; diagnostics omit raw values.
- [ ] errors.Is/As preserve API failures and pre/post-call cancellation. Canceled callers do not cancel shared cache refresh.
- [ ] Offline generated STS -> provider -> cache -> generated ECS request verifies original source signing, assumed-role signing, native fields and expiry rotation. Default/non-idempotent token issuance remains non-retrying.
- [ ] Runnable external Example, doc.go/Go docs and equivalent English/Chinese provider guide and acceptance evidence.
- [ ] doccheck, formatting, vet, full Go tests and product-check pass; Linux race/Windows CI recorded separately.

### Limits

This issue establishes offline composition. Live role/refresh, native Profile/OAuth loading, independent UX testing, Beta completion and publication remain open acceptance cases. No real role or cloud write is needed.

## 中文

### 领域、依赖及问题

领域为 credentials/sts/testing，依赖 #49 的 AC-05/UX-04；基础/cache/完整 DSL 已接受，文档 PR 为明确堆叠基分支。现有 helper 接口来自 services/sts，新 service/sts 客户端不能直接接入，用户需手写凭据转换。

### 范围与验收

新增 NewAssumeRoleProviderFromClient 接受完整 DSL 窄接口，保留旧构造器及返回类型。构造/调用复制输入指针与 options，验证角色/会话及已有审核 helper 规则，不改变生成操作。转换准确响应，校验密钥/token 和未来 RFC3339 过期时间，复用 cache，保留来源、结构化错误、取消与安全诊断。不改生成文件、不加依赖、不猜重试安全。

- [ ] 新构造器、小 mock、格式化脱敏、零值/nil 安全，旧 API/Example 保留。
- [ ] 保留指针存在语义、可选字段和时长宽度，构造/API/callback 修改不延续，并发扩展安全。
- [ ] 拒绝 nil/typed-nil、nil options、无效 helper 输入、缺失/空白密钥/token、缺失/非法/过期时间，错误不带值。
- [ ] 保留 errors.Is/As 和调用前后取消；取消某个 cache 等待者不取消公共刷新。
- [ ] 离线生成 STS→provider→cache→生成 ECS 验证来源/角色签名、线字段、过期轮换；发 token 保持不重试。
- [ ] 可运行外部 Example、Go 文档/doc.go、中英指南/验收证据。
- [ ] 文档/格式/vet/全 Go 测试/product-check 通过，Linux race/Windows CI 另外记录。

仅证明离线组合，真实角色/刷新、原生 Profile/OAuth、独立用户体验、Beta/发布仍未验收，不需真实角色或云写入。
