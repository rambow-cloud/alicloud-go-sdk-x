# Project working agreements / 项目约束

[English](#english) | [中文](#中文)

## English

# Project working agreements

## Product goal

Build an independent Alibaba Cloud SDK for Go with idiomatic APIs, small dependencies,
predictable request behavior, and documentation available on pkg.go.dev. This project
is not an official Alibaba Cloud SDK. Read `docs/research.md` and `docs/design.md`
before changing the public API.

Follow docs/product-acceptance.md for candidate Beta/release scope, AC/UX criterion
IDs and required evidence. Record PASS/FAIL/SKIP/NOT RUN per case; skipped required
cases keep the gate open. Generated/compiled counts, local PoCs and closed implementation
issues do not establish Beta/release acceptance. Keep Smithy isolated until a separately
approved route decision; follow the full-DSL STS composition -> consumer/official-v2
comparison -> authorized live gaps -> source-update rehearsal -> release sequence.

The user-approved 2026-10-08 first release is v0.1.0 STS, governed by
docs/sts-v0.1.0.md, milestone v0.1.0, Project 3 and parent #57. This scoped
experimental release supersedes conflicting broader Beta scheduling; it does not
claim ECS/VPC/STS Beta. Execute #58 requestless GetCallerIdentity -> #59 anonymous
OIDC/SAML RPC -> #60 four-action acceptance/consumer/source-update rehearsal -> #61
release/indexing. #58/#59 are independent prerequisites for #60. Keep full official
Darabonba/parser -> IR -> our Go backend/runtime; signed credential rules remain
strict, anonymous auth is explicit per operation. Preserve #51/#53/#55 evidence and
disclose OIDC/SAML live federation as NOT RUN/outside this release's required live
scope. Maintain milestone membership, native issue hierarchy/dependencies, one
status label and matching Project Status; do not claim automatic synchronization.
Keep parent/milestone open until required release/indexing evidence is delivered.

## Issue-driven development

- Every code or behavior change starts with an issue containing the problem, evidence,
  scope, acceptance criteria, and documentation requirements. Use GitHub issues in
  `rambow-cloud/alicloud-go-sdk-x`; when offline, create a draft under `docs/issues/`
  and synchronize it before opening a PR.
- Keep one independently reviewable issue per branch: `issue/<number>-<slug>`.
- Before implementing, establish how the acceptance criteria will be verified.
- PRs link the issue using `Closes #<number>` for complete work, or `Refs #<number>`
  for partial work, and include behavior, validation, and docs.
- A change is complete only when code, relevant checks, examples, and package docs
  satisfy the issue. Do not close roadmap issues just because an interface exists.
- Never invent issue numbers, completed checks, compatibility, or service coverage.

## Go API rules

- Follow AWS Go SDK v2 calling conventions for service Options, NewFromConfig,
  operation functional options, paginators and reusable Wait/WaitForOutput. Preserve
  Alibaba OpenAPI operation names, exact wire fields and native pagination semantics;
  never invent NextToken for a page-only API. Document intentional defaults and v0
  migration differences rather than claiming upstream source compatibility.

- Every blocking operation accepts `context.Context` as its first argument and
  preserves cancellation/deadline errors for `errors.Is`.
- Use standard `net/http`, injectable HTTP clients, and `time.Duration` for timeouts.
- Require Go 1.27 or newer. Import `encoding/json/v2` directly for JSON; do not add
  legacy `encoding/json`, third-party JSON libraries, or an experiment flag.
- Keep client configuration private after construction; do not mutate shared state
  or caller requests. Document concurrency contracts of extension interfaces.
- Use concrete request/response types and `errors.As` for service errors. Pointers
  represent meaningful absence; do not require helpers for every ordinary value.
- Retry only when the operation's idempotency and error policy allow it. Bound total
  attempts and elapsed time, respect context during backoff, and never retry arbitrary
  writes by default.
- Add dependencies only with issue justification. Keep the runtime core standard
  library only unless a concrete requirement prevents it.
- Keep signing, encoding, and retry machinery internal until real use requires an
  extension contract. Do not implement cloud operations without protocol evidence.
- Never log credentials, authorization headers, or raw request/response bodies by
  default. Do not commit real credentials or make live cloud calls in unit tests.
- Make renewable STS role providers with credentials.Cache the primary application
  guidance. Config/service Options accept providers only, never bare AK/SK/token
  fields. Long-lived keys and environment sources require explicit StaticProvider
  or EnvProvider registration; preserve deliberate custom providers. Do not add
  implicit discovery or enforce an STS-only runtime. Reject nil/typed-nil providers
  before requests without retrieving credentials during construction.

## pkg.go.dev definition of done

- Every public package has a `doc.go` overview explaining usage and limitations.
- Every exported declaration, field, and interface method has a Go doc comment
  beginning with its name. Explain zero values, defaults, errors, cancellation,
  ownership, and concurrency when relevant. Use native Go comments, not `@param`.
- Every public package has at least one runnable external `Example` with deterministic
  output and no account/network requirement. Add operation examples as operations land.
- Keep `LICENSE`, canonical module imports, and runnable README snippets current.
- Run `go run ./internal/cmd/doccheck`, `go vet ./...`, and `go test ./...` once for
  a meaningful change. CI runs the race detector on Linux. Check formatting once.
- `docs/releasing.md` describes indexing; local docs passing does not mean the package
  has already been published or indexed.

## Validation and communication

- Avoid redundant validation and repeated retries. Repeat only after a meaningful
  state change or a failure whose reason is newly understood.
- If a result can be verified by opening it in a browser, tell the user the exact URL
  and what to inspect instead of checking it using curl or similar local requests.
- For DNS, server-side CloudFormation status and outputs are the source of truth.
  Do not validate DNS locally; provide user verification steps only if needed.
- Do not spawn sub-agents unless the user explicitly requests delegation.

## Language and development sequence

- The user-approved 2026-10-07 route in docs/development-path.md and
  docs/product-generator-roadmap.md is authoritative and overrides conflicting older
  metadata-first, per-operation snapshot/overlay and handwritten-doc prerequisites.
  Historical acceptance remains recorded; superseded plans must not constrain new work.
- The initial five-stage route and #44 review fix are integrated into main as recorded
  in docs/generator-integration.md (2026-10-08). New product/protocol/capability expansion
  starts with a separate issue and acceptance scope; preserve this accepted baseline.
- Execute source normalization -> complete DSL operation/model discovery and IR -> batch
  Go emission -> sparse capability policy -> documentation automation/profile expansion.
  Complete official DSL and the official semantic parser are primary; canonical metadata
  is optional enrichment/cross-check input after representation normalization.
- Discover operations/reachable models automatically without per-API metadata, field/model
  selection or handwritten-doc overlays. Overlays supply reviewed policy and compatibility
  exceptions. #31 is a five-operation bridge, not product-wide acceptance or a permanent
  per-API authoring workflow. Preserve runtime/AWS conventions, Go 1.27 and JSON v2.
- Publish deterministic coverage/reasons; discovered, lowered, emitted, compiled and live
  coverage are distinct. Selected unsupported behavior fails before writes. Normalize
  indexed inputs/itemName wrappers before source conflict decisions.
- Full-DSL products use `service/<product>`; `services/<product>` retains the bounded
  compatibility/reference bridge. Follow docs/batch-go-emission.md for explicit migration.
  Product changes also run `sdkgen product-check`; old/new generators must preserve each
  other's owned outputs. Product capabilities use optional source-bound policies under
  policies/ (#37); resolve exact wire paths against complete IR, fail invalid policy
  before writes, and report unlisted actions as unreviewed. Never infer retry safety or
  paginator/waiter behavior from operation names or token-shaped fields. Follow
  docs/capability-policy.md; caller inputs stay owned copies and retry remains opt-in.
- Establish bilingual route docs, then actual issue dependencies, then stage implementation
  on separate issue branches. Each stage includes docs/tests/Examples/pkg.go.dev acceptance;
  the documentation automation stage does not postpone earlier documentation.
- Follow docs/product-documentation.md for #38 documentation: reuse official parser
  prose with source coordinates and Apache notices; generate English Go comments and
  paired usage/contracts/source indexes. Report missing language/prose coverage, never
  infer runtime policy or promote upstream account/resource examples into Go tests.

- All human-facing Markdown documentation must contain equivalent English and Chinese
  sections, or explicitly linked language counterparts, updated in the same change.
- GitHub issue and PR titles/bodies, identifiers, diagnostics, and code comments are
  English-primary. Chinese explanations may supplement them; never replace the English
  problem, scope, dependencies, acceptance criteria, or verification evidence.
- Go package and symbol documentation remains English-primary for pkg.go.dev. Provide
  equivalent Chinese usage/behavior guidance in the paired package guides.
- Complete the eleven shared foundation capabilities and their handwritten ECS/STS
  contract tests before starting generator implementation. Metadata research is allowed
  earlier; interfaces alone do not satisfy the foundation acceptance gate.
- Read docs/development-path.md before selecting an issue. Document the path, create or
  update the issue, then implement. Keep the issue dependency graph acyclic.
- Every capability includes implementation, behavior tests, executable examples, paired
  documentation, and recorded verification. Keep generator and benchmark issues separate.
- Generator changes also run `go run ./internal/cmd/sdkgen check`. Edit pinned metadata,
  overlays, templates or handwritten validators under an issue; do not edit generated
  files directly. Review schema drift and rerun generation before public API checks.
- Production generation uses pinned official Darabonba sources/imports and the official
  semantic parser projection. Run the frontend check/tests with Node 22 before Go gates.
  Never silently resolve DSL/metadata conflicts; maintain the paired decision document
  and machine-readable approvals. Explorer browser evidence and CLI evidence are distinct.
- Preserve third-party source, README and license notice bytes under sources/darabonba
  and sources/openapi-meta;
  their upstream documents are source artifacts. All project-authored source/tool guides
  and decisions remain bilingual; upstream notices are not relabeled under our MIT license.

## 中文

### 产品目标

构建独立的阿里云 Go SDK，提供 Go 原生 API、小依赖核心、可预测请求行为和 pkg.go.dev 文档。
本项目不是阿里云官方 SDK。修改公共 API 前阅读调研、设计和开发路径。

候选 Beta/发布范围、AC/UX 标准编号和证据遵循 docs/product-acceptance.md，逐项记录
PASS/FAIL/SKIP/NOT RUN，必需项跳过则门槛未通过。生成/编译数量、本地 PoC 或实现 issue
关闭不代表 Beta/发布验收。Smithy 保持隔离，改变路线需独立批准的决策；继续按完整 DSL STS
组合→消费者/官方 v2 对比→授权真实缺口→来源升级演练→发布执行。

用户于 2026-10-08 指定首版为 v0.1.0 STS，遵循 docs/sts-v0.1.0.md、同名 milestone、
Project 3 和父 issue #57。限定实验版本优先于冲突旧 Beta 排期，不宣称 ECS/VPC/STS
整体 Beta。先 #58 无请求 GetCallerIdentity→#59 匿名 OIDC/SAML RPC→#60 四操作验收/
消费者/来源升级→#61 发布/索引；#58/#59 独立且均为 #60 前提。保留官方 Darabonba/parser
→IR→本项目 Go 后端/runtime；签名凭据规则严格，匿名认证逐操作显式。保留 #51/#53/#55
证据，OIDC/SAML 真实联邦预先列为首版必需真实范围外并记 NOT RUN。维护 milestone、
原生子项/依赖、唯一状态 label 及对应 Project Status，不宣称自动同步；发布/索引证据
齐全前父项/milestone 保持打开。

### Issue 驱动

- 每项代码或行为修改先建立 issue，包含问题、证据、范围、验收和文档要求；使用本仓库 GitHub issues。
  离线时先在 docs/issues/ 建立草稿，并在打开 PR 前同步。
- 每个可独立评审的 issue 使用 issue/<number>-<slug> 分支，实现前明确验证方式。
- 完整交付 PR 使用 Closes #N，部分交付使用 Refs #N，说明行为、验证和文档。
- 代码、检查、示例和文档同时满足验收才算完成；不能仅因接口存在就关闭路线图 issue。
- 不编造 issue 编号、检查结果、兼容性或服务覆盖。

### Go API

- 服务 Options、NewFromConfig、操作 functional options、分页及可复用 Wait/WaitForOutput
  采用 AWS Go SDK v2 调用范式；操作名、准确线字段和原生分页语义保留阿里云 OpenAPI 规则，
  不为纯页码 API 造 NextToken。明确默认策略与 v0 迁移差异，不宣称上游源码兼容。

- 阻塞操作第一参数为 context.Context，errors.Is 能识别取消和超时。
- 使用标准 net/http、可注入 HTTP 客户端与 time.Duration；最低 Go 1.27，JSON 直接使用 encoding/json/v2，
  不引入旧版 JSON、第三方替代或实验开关。
- 构造后配置保持私有，不修改共享状态或调用者请求；扩展接口明确并发与所有权契约。
- 产品请求/响应使用具体类型；errors.As 提取服务错误；指针只表达有意义的缺失，不强制普通值使用辅助函数。
- 重试需满足幂等性和错误策略，限制尝试次数及总时间，退避可取消；任意写操作默认不重试。
- 新依赖须有 issue 依据；核心运行时保持标准库依赖，除非具体需求证明无法做到。
- 签名、编码及重试实现保持内部，实际需要用户扩展时公开稳定契约；云操作必须有协议依据。
- 默认不记录凭据、授权头或原始请求/响应体；不提交真实凭据，单元测试不访问真实云资源。
- 应用指南优先采用可刷新的 STS role provider 与 credentials.Cache；Config/服务 Options
  只接受 provider，不增加裸 AK/SK/token 字段。长期密钥及环境来源必须显式注册
  StaticProvider/EnvProvider，保留有意注入的自定义 provider；不隐式发现来源，不强制运行时
  仅接受 STS。请求前拒绝 nil/typed-nil provider，构造期间不读取凭据。

### pkg.go.dev 完成标准

- 每个公共包提供 doc.go 概述，说明用途和限制；每个导出声明、字段和接口方法有以名称开头的原生 Go 注释。
- 相关注释说明零值、默认值、错误、取消、所有权和并发，避免 @param 风格。
- 每个公共包至少一个输出确定、无需网络/云账号的外部 Example；操作实现时同步添加示例。
- 维护标准 LICENSE、规范 module 导入和可运行 README 片段。
- 对有意义的变更各运行一次文档检查、vet、Go 测试和格式检查；Linux CI 使用 race。
- 发布与索引见 releasing 文档；本地文档检查通过不代表网页已经发布或索引。

### 验证与沟通

- 不重复验证或盲目重试；只有状态改变或失败原因已理解时重新执行。
- 网页结果告知用户准确 URL 和检查项，避免用 curl 等本地请求重复验证。
- DNS 以服务端 CloudFormation 状态与输出为准，不做本地 DNS 验证；需要时提供用户验证步骤。
- 用户未明确要求时不启动子代理。

### 语言与开发顺序

- 用户于 2026-10-07 确认的 docs/development-path.md 和 docs/product-generator-roadmap.md
  为权威路线，优先于冲突旧元数据优先、逐操作 snapshot/overlay、手写说明前置要求；
  保留历史验收，不用已被替代的计划约束新工作。
- 初始五阶段路线及 #44 评审修复已集成 main，证据见 docs/generator-integration.md
  （2026-10-08）。更多产品/协议/能力先建独立 issue 和验收范围，保留此已接受基线。
- 按来源规范化 → 完整 DSL 操作/模型发现与 IR → 批量 Go 输出 → 少量能力策略 →
  文档自动化/协议扩展执行。完整官方 DSL/官方语义 parser 为主，canonical 元数据
  规范化后可选补充/交叉验证。
- 自动发现操作/可达模型，不以逐 API 元数据、字段/模型选择、手写说明 overlay 为前提。
  Overlay 只补审核策略和兼容例外；#31 为五操作兼容桥，不是产品全量验收或永久逐 API
  手写流程。保留 runtime/AWS 范式、Go 1.27、JSON v2。
- 输出确定性覆盖/原因，区分发现/降低/输出/编译/真实验收；选中不支持行为写前失败。
  先规范化索引输入/itemName 包装，再判断来源冲突。
- 完整 DSL 产品使用 `service/<product>`，原 `services/<product>` 保留有界兼容/参考桥；
  明确迁移见 docs/batch-go-emission.md。产品变更还需运行 `sdkgen product-check`，新旧
  生成器相互保留各自产物。产品能力使用 policies/ 下可选且绑定来源的策略（#37），
  按完整 IR 解析准确线路径，无效策略写前失败，未列操作报告未审核。不从操作名称
  或形似 token 的字段推断重试安全/分页/waiter。遵循 docs/capability-policy.md，
  输入保持独占副本，重试仍显式启用。
- 先落实双语路线，再建真实 issue 依赖，然后各 issue 分支实现阶段；每阶段同步文档/
  测试/Example/pkg.go.dev，文档自动化阶段不表示前面可推迟文档。
- #38 文档遵循 docs/product-documentation.md：复用官方 parser 说明，保留来源坐标及
  Apache 通知；输出英文 Go 注释和对应双语使用/契约/来源索引。明确缺失语言/说明
  覆盖，不推断 runtime 策略、不将上游账号/资源示例提升为 Go 测试。

- 所有面向人的 Markdown 文档必须有语义对应的英文与中文章节，或明确互链的语言版本，并在同一变更更新。
- GitHub issue/PR 标题正文、标识符、诊断和代码注释以英文为主；中文可补充，但不能替代英文的问题、
  范围、依赖、验收或验证证据。
- pkg.go.dev 的包和符号文档以英文为主；配对的使用指南提供对应的中文行为说明。
- 十一项公共基础能力及手写 ECS/STS 契约测试验收后才能开发 generator；之前可以调研元数据，接口存在不等于验收完成。
- 选取 issue 前阅读开发路径；先写路径、再建或更新 issue、然后实现；依赖图保持无环。
- 每项能力包含实现、行为测试、可执行示例、双语文档与验证记录；生成器与基准分开跟踪。
- 生成器变更还需运行 `go run ./internal/cmd/sdkgen check`。在 issue 下修改固定元数据、overlay、
  模板或手写 validator，不直接修改生成文件；评审 schema 漂移并重新生成后执行公共 API 检查。
- 生产生成使用固定官方 Darabonba 源码、导入模块及官方语义 parser 投影；Go 门禁前执行 Node 22
  前端检查/测试。不静默解决 DSL/元数据冲突，维护双语决策及机器可读审核记录；Explorer 浏览器
  与 CLI 证据分别记录。
- sources/darabonba 和 sources/openapi-meta 下的第三方源码、README、许可证通知
  保留上游原始字节，作为源码产物；
  本项目编写的来源/工具指南及决策仍需双语，不把上游通知重新标为本项目 MIT。
