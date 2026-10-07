# Development path / 开发路线

[English](#english) | [中文](#中文)

## English

Authoritative route revised at the user's direction on 2026-10-07: complete official
product DSL -> official Darabonba semantic parser -> normalized operation/model/binding
IR -> our Go backend -> existing runtime. [product-generator-roadmap.md](product-generator-roadmap.md)
and AGENTS.md govern this route and supersede conflicting older metadata-first,
per-operation snapshot/overlay, field-selection and handwritten-doc prerequisites.
The older stages and issue references below describe historical acceptance only.

Execute source normalization -> complete operation/model discovery and IR -> batch Go
emission -> sparse capability policy -> documentation automation/profile expansion.
Every stage includes paired docs, relevant tests, Examples and pkg.go.dev acceptance;
documentation automation at the last stage does not postpone earlier documentation.
Optional pinned canonical metadata enriches/cross-checks DSL after representation
normalization. New discovery cannot require metadata/overlay entries for each API.

First milestone: normalize real source representations and publish complete pinned
ECS inventory/coverage; then batch generate supported RPC operations. #31/PR #32 are
the five-operation compatibility bridge, not a product-wide generator acceptance gate.
Overlay supplies reviewed compatibility exceptions, pagination/waiters, idempotency,
sensitive fields and special validators, rather than re-declaring each model/field.
Unsupported operations have explicit reasons; selected unsupported behavior fails
before writes. Normalize Filter indexed bindings and itemName response wrappers before
classifying conflicts. Browser, CLI-local validation and actual HTTP evidence are distinct.

Document -> establish/update actual issue dependencies -> one reviewable branch per
issue -> implement, validate and record -> linked PR. Keep the roadmap parent open
until all stages meet acceptance, preserve accepted runtime/AWS calling conventions,
Go 1.27/JSON v2 and standard-library core. Benchmarks remain separate.

The goal is a unified Go runtime, validated by a small handwritten ECS/STS reference,
followed by product generation. Go 1.27 and encoding/json/v2 are required. Core runtime
imports remain standard-library-only; OpenTelemetry is an optional integration.

| Stage | Deliverables | Gate |
| --- | --- | --- |
| A: specification | bilingual docs, English issues, acyclic dependencies, language policy | documented before code |
| B: request foundation | staged middleware, endpoint resolver, ACS3 signing, HTTP execution, structured errors, testing helpers | offline request/response and concurrency contracts |
| C: resilience and identity | explicit credential chain, expiry-aware cache, bounded retries, STS AssumeRole helper | cancellation, freshness, replayability, idempotency |
| D: reference behavior | handwritten ECS reads, unified pagination, bounded waiters, optional OTel | typed usage and cross-capability integration |
| E: foundation acceptance | eleven capabilities, runnable examples, paired guides, Linux race and Windows CI | implementations and tests, not interface presence |
| F: generation | frozen metadata to typed clients/codecs/docs/policies | #19 passed; follow #21 -> #22 -> #23 under #8 |

Dependency order: testing/middleware/endpoints/errors and signing -> HTTP runtime ->
ECS/STS clients -> paginator/waiter/AssumeRole integration. Credential chain/cache and
retry policy can be implemented independently before runtime integration. OTel depends
on middleware and request metadata. Signing's standalone acceptance must not depend on
HTTP integration; integration belongs to the HTTP issue, removing the previous cycle.

All eleven requested areas are mandatory: paginator, waiter, retry/backoff, mock interfaces,
credential provider, STS helper, endpoint resolver, middleware, OTel, structured errors,
testing helpers. Public extension contracts are shared; product-specific rules remain typed.

Initial protocol is ACS3-HMAC-SHA256 for Alibaba Cloud OpenAPI. Initial reference coverage:
ECS DescribeRegions, DescribeInstances (selected response fields), DescribeInstanceStatus;
STS AssumeRole. This is not full ECS/STS, OSS/SLS data-plane, or live-cloud acceptance.
Use explicit endpoint rules; unsupported regions fail rather than guessing domains.

Preserve the initial no-retry default. Standard retry is opt-in, bounded, cancellation-aware,
with jitter and Retry-After, and applies only to explicitly idempotent/replayable operations.
Waiter total expiry is distinct from per-request retry. Credential refresh is bounded and
shared without allowing one canceled waiter to cancel other callers. No implicit process or metadata credentials.

Every issue includes package docs, deterministic external Examples and paired behavior guides.
GitHub is the source of execution status. #1/#2 are historical bootstrap completion;
#3-#7/#9 are refined foundation tasks; #8 is generation only. Benchmark work is separate.
New foundation issues are listed in the issue index and GitHub milestone; record actual numbers, never assume them.

Generator architecture, supported profile, provenance and acceptance are defined in
[generator.md](generator.md). The foundation gate passed on commit 28684e4 before #21.
The next real-product gate is #24 -> #25; see [generator-expansion.md](generator-expansion.md).
After that acceptance, resolve the review under #26 -> #27 -> #28 -> #29 before wider
profiles: [AWS-style remediation](aws-style-remediation.md).

The frontend refactor #31 uses official product DSL and the official parser; its
ordered delivery and conflict verification are defined in [darabonba-migration.md](darabonba-migration.md).

Sources: [ACS3](https://help.aliyun.com/zh/sdk/product-overview/v3-request-structure-and-signature),
[metadata](https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/),
[AWS middleware](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/middleware.html),
[AWS testing](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/unit-testing.html).

## 中文

用户于 2026-10-07 确认的权威路线：完整官方产品 DSL → 官方 Darabonba 语义 parser →
规范化操作/模型/绑定 IR → 本项目 Go 后端 → 已有 runtime。
[product-generator-roadmap.md](product-generator-roadmap.md) 和 AGENTS.md 规定新路线，
优先于冲突的旧元数据优先、逐操作 snapshot/overlay、字段选择、手写说明前置要求。
下文旧阶段及 issue 引用仅作为历史验收记录。

按来源规范化 → 完整操作/模型发现及 IR → 批量 Go 输出 → 少量能力策略 → 文档自动化/
协议扩展推进。每阶段同步双语文档、相关测试、Example 及 pkg.go.dev，最后的文档
自动化不表示前面可以推迟文档。固定 canonical 元数据在规范化后可选补充/交叉验证，
新扫描不以每 API 的元数据/overlay 为前提。

首个里程碑是规范化真实表示并交付完整固定 ECS 的 inventory/覆盖报告，再批量生成
支持的 RPC。#31/PR #32 是五操作兼容桥，不是产品全量验收。Overlay 仅补审核的兼容
例外、分页/waiter、幂等、敏感字段及特殊校验，不重述每个模型/字段。未支持操作明确
列原因，选中不支持行为写前失败；先规范化 Filter 索引绑定及 itemName 响应包装再判
冲突，浏览器、CLI 本地校验、真实 HTTP 证据分别记录。

文档 → 建/更新真实 issue 依赖 → 每项独立分支 → 实现/验证/记录 → 关联 PR。路线父
任务在所有阶段验收前保持打开，保留已验收 runtime/AWS 范式、Go 1.27/JSON v2、
标准库核心；基准独立。

目标是先完成统一 Go runtime，用少量手写 ECS/STS 参考操作验证，然后建设产品生成器。
要求 Go 1.27、直接使用 encoding/json/v2；核心运行时导入保持标准库，OpenTelemetry 为可选集成。

| 阶段 | 交付 | 门槛 |
| --- | --- | --- |
| A：规格 | 双语文档、英文 issue、无环依赖与语言规则 | 写在代码之前 |
| B：请求基础 | 分阶段 middleware、endpoint、ACS3、HTTP、结构化错误、测试辅助 | 离线请求/响应与并发契约 |
| C：可靠性与身份 | 显式凭据链、过期缓存、有界重试、STS helper | 取消、新鲜度、可重放与幂等性 |
| D：参考行为 | 手写 ECS、统一分页、waiter、可选 OTel | 类型化用法与跨能力集成 |
| E：基础验收 | 十一项能力、示例、双语指南、Linux race 与 Windows CI | 实现及测试，而非仅有接口 |
| F：生成器 | 固定元数据生成类型化客户端、编码、文档与规则 | #19 已通过，按 #8 下 #21 → #22 → #23 执行 |

依赖顺序：测试/middleware/endpoint/错误及签名 → HTTP runtime → ECS/STS → 分页/waiter/AssumeRole 集成。
凭据链/缓存与重试策略可先独立实现，再接 runtime；OTel 依赖 middleware 和请求元数据。
签名单独验收不依赖 HTTP，集成验收归 HTTP issue，解除之前的循环。

十一项均必须交付：分页、waiter、重试/退避、mock 接口、凭据 provider、STS helper、endpoint、middleware、
OTel、结构化错误、测试辅助。共享公共扩展契约，产品特殊规则采用具体类型。

首版协议为 OpenAPI ACS3-HMAC-SHA256。参考操作：ECS DescribeRegions、DescribeInstances（选定响应字段）、
DescribeInstanceStatus，以及 STS AssumeRole。范围不包含全量 ECS/STS、OSS/SLS 数据面或真实云验收。
Endpoint 使用明确规则，不支持的地域报错，不猜测域名。

保留默认不重试；显式启用的标准策略有界、可取消、带 jitter/Retry-After，仅适用于明确幂等且可重放的操作。
Waiter 总期限与单次重试分开；凭据刷新有界并合并并发，一个等待者取消不影响其他调用者。
不隐式执行外部进程或读取 metadata 凭据。

每项 issue 包含包文档、确定性外部 Example、对应双语指南。GitHub 是执行状态来源。
#1/#2 保留历史完成记录，#3-#7/#9 细化为基础任务，#8 只负责生成器；基准独立跟踪。
新增 issue 的真实编号在索引与里程碑中登记，不预设编号。协议及设计参考来源与英文章节相同。

生成器架构、支持范围、来源及验收见 [generator.md](generator.md)；#21 开始前，基础门槛已在 28684e4 通过。
下一真实产品门槛为 #24 → #25，见 [generator-expansion.md](generator-expansion.md)。
验收后，扩展更多 profile 前按 #26 → #27 → #28 → #29 修复 review，见
[AWS 风格修复路径](aws-style-remediation.md)。
后续前端重构使用官方产品 DSL 与官方 parser，交付顺序及偏差验证见
[Darabonba 迁移路径](darabonba-migration.md)。
