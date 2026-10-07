# Runtime-first development path / 基础优先开发路径

[English](#english) | [中文](#中文)

## English

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
