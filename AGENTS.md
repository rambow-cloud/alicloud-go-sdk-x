# Project working agreements / 项目约束

[English](#english) | [中文](#中文)

## English

# Project working agreements

## Product goal

Build an independent Alibaba Cloud SDK for Go with idiomatic APIs, small dependencies,
predictable request behavior, and documentation available on pkg.go.dev. This project
is not an official Alibaba Cloud SDK. Read `docs/research.md` and `docs/design.md`
before changing the public API.

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

## 中文

### 产品目标

构建独立的阿里云 Go SDK，提供 Go 原生 API、小依赖核心、可预测请求行为和 pkg.go.dev 文档。
本项目不是阿里云官方 SDK。修改公共 API 前阅读调研、设计和开发路径。

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

- 所有面向人的 Markdown 文档必须有语义对应的英文与中文章节，或明确互链的语言版本，并在同一变更更新。
- GitHub issue/PR 标题正文、标识符、诊断和代码注释以英文为主；中文可补充，但不能替代英文的问题、
  范围、依赖、验收或验证证据。
- pkg.go.dev 的包和符号文档以英文为主；配对的使用指南提供对应的中文行为说明。
- 十一项公共基础能力及手写 ECS/STS 契约测试验收后才能开发 generator；之前可以调研元数据，接口存在不等于验收完成。
- 选取 issue 前阅读开发路径；先写路径、再建或更新 issue、然后实现；依赖图保持无环。
- 每项能力包含实现、行为测试、可执行示例、双语文档与验证记录；生成器与基准分开跟踪。
- 生成器变更还需运行 `go run ./internal/cmd/sdkgen check`。在 issue 下修改固定元数据、overlay、
  模板或手写 validator，不直接修改生成文件；评审 schema 漂移并重新生成后执行公共 API 检查。
