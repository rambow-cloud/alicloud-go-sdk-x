# Live STS renewal acceptance / 真实 STS 续期验收

## English

## Problem and evidence

#51/#53 provide offline full-DSL STS provider/cache integration, but AC-05 live issuance and renewal are unverified. The user now authorizes local unsandboxed verification using the existing Aliyun Profile. Current default oss-sftp uses OAuth and has no configured target role; the user now authorizes creation of a dedicated validation role. CLI preflight PASS confirms an Account identity, which cannot AssumeRole directly.

## Scope and dependencies

Depends on merged #49/#51/#53. Plan: docs/live-sts-renewal.md, baseline f4cebddd3660d0d09d2b56e6319f5895c2f4d4a5. Preflight one bounded CLI GetCallerIdentity and use that account Profile only for setup/cleanup. Inject a separate explicit RAM-user source into generated STS. Use the dedicated authorized role with DurationSeconds=900, real server time, existing role provider/cache and generated consumer reads. Provision only a unique RAM user without console login, one temporary AccessKey kept in memory, a policy permitting sts:AssumeRole on one new role, and a role trusting that user with only ecs:DescribeRegions. Clean up all newly owned resources/attachments after the run. Do not modify existing identities/permissions/resources. No implicit SDK Profile/OAuth feature.

## Acceptance criteria

- [ ] Verify the selected source identity/authentication and sufficient source lifetime; stop safely on missing role or invalid login.
- [ ] Perform real AssumeRole and validate nonblank keys/token/future expiry, role identity and structured HTTP metadata without exposing credentials or account/resource identifiers.
- [ ] Prove cache reuse and a separately labeled forced Invalidate refresh with actual issuance counts and in-memory change comparisons.
- [ ] Wait for genuine server-side expiration, then prove automatic cache renewal and successful role-authenticated generated-client reads with no virtual time/manual invalidation.
- [ ] Publish paired sanitized evidence with version/commit/timing/status/limits. Authorization failure, missing prerequisites or skipped required cases do not establish PASS or Beta.

## Verification

Keep disposable Go 1.27/JSON-v2 harness and reports under ignored .git/. Bound network requests/context and CLI retries; progress waits at most 60 seconds. Use real clock, explicit source provider, generated service/sts, stscreds and Cache, with the generated AssumeRole response identity plus generated ECS DescribeRegions (GetCallerIdentity is CLI preflight only; it is not emitted by the pinned STS DSL); do not broaden permissions on failure. Record STS versus ECS authorization separately. Existing offline CI remains the baseline; no production code change is planned.

## 中文

依赖已合并 #49/#51/#53，用户授权沙箱外本地验证真实 STS 续期。先按双语计划预检 oss-sftp 身份/登录和来源期限；用户已授权创建专用角色；CLI 预检确认来源为主账号，无法直接 AssumeRole，因此需要无控制台登录的专用 RAM 用户、一枚仅留内存的临时 AccessKey、只可扮演该角色的策略，以及仅信任该用户、只可 ecs:DescribeRegions 的角色/策略。测试后清理本次专用对象和授权。使用 900 秒真实凭据，验证生成 STS→provider/cache→生成消费者的签发、复用、强制刷新和自然到期后的自动续期，全部区分记录。脚本/本地报告仅放忽略 .git/，公开脱敏状态、次数、版本、时间、限制；不输出凭据/账号/资源标识，只创建授权的本次专用角色/用户/最小策略，不修改既有对象/权限/信任/资源，不新增 SDK 原生 Profile/OAuth 功能。缺失前提或授权失败不算通过/Beta，等待间隔最多 60 秒，无生产代码变更计划。

管理主账号 Profile 仅用于创建/清理，生成 STS 注入独立、显式的专用 RAM 用户来源。

### 对应验收与验证

- 验证所选管理来源身份/认证与足够有效期，缺失角色或登录失效时安全停止。
- 真实 AssumeRole 校验非空密钥/token、未来期限、角色身份、结构化 HTTP 元数据；不公开凭据或账号/资源标识。
- 用真实签发次数和内存变化比较证明缓存复用与单独标明的 Invalidate 强制刷新。
- 等待服务端真实期限后，无虚拟时钟/手动失效，验证自动续期和生成客户端读取成功。
- 配对发布脱敏版本/提交/时间/状态/限制；失败或跳过必需项不代表 PASS/Beta。

一次性 Go 1.27/JSON v2 脚本和报告放忽略 .git/，网络/context/CLI 重试有界，等待间隔最多
60 秒。使用真实时钟、显式来源、生成 STS/provider/cache、AssumeRole 返回身份和生成 ECS；
GetCallerIdentity 仅为 CLI 预检，不是固定 DSL 生成操作。权限失败不扩大权限，区分 STS/ECS，
既有离线 CI 为基线，无生产代码变更。
