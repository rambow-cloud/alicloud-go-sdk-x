# AWS-style review remediation / AWS 风格 review 修复路径

## English

This path precedes implementation of the review on commit abcf961. Use AWS Go SDK v2
calling conventions with Alibaba OpenAPI operation names, wire fields, token values and
page semantics. Preserve Go 1.27, JSON v2, explicit credentials, no-retry defaults,
thirty-second operation timeout, HTTPS rules and standard-library core dependencies.
This is an independent SDK; neither upstream SDK's source compatibility is promised.

| Order | Issue scope | Acceptance |
| --- | --- | --- |
| 1 | runtime attempt isolation and interrupted response retries | no stale output/successful incomplete short circuit; bounded idempotent EOF/read retries; cancellation/non-idempotent/JSON failures unchanged |
| 2 | typed middleware and service Options | Initialize -> Serialize -> Build once; Finalize/Deserialize per attempt; owned typed input/output; NewFromConfig and concrete service Options; call retry/transport overrides stay isolated |
| 3 | pagination and generator policy collections | current page delivered on repeated continuation, configurable stop; overflow-safe pages; dedicated options and NextPage call options; multiple token-only/page-only/dual adapters without invented wire fields |
| 4 | reusable waiter API | inputs on Wait/WaitForOutput, independent concurrent waits, dedicated options/default acceptor override, all-ID safety and cancellation preserved |

Actual GitHub issues were created before code, ordered #26 -> #27 -> #28 -> #29.
Each issue gets its own branch/commit, tests and equivalent English/Chinese documentation. Regenerate
through the emitter/templates; do not patch generated code. Close only after full local
gates and exact-commit Linux race/Windows CI. Benchmarks, ROA/body, default discovery and
changes to default retry/timeout policies are separate work.

The review reproduced stale decoded output across retries, loss of a page when a token
repeats, signed integer overflow in ECS legacy page traversal, and an idempotent response
read ending in io.ErrUnexpectedEOF without retry. Regression fixtures must address those
triggers directly, alongside generated clients and existing ECS/STS/VPC contracts.

Migration notes must distinguish additive changes from early-v0 paginator/waiter changes.
Keep the existing New(Config) convenience path; NewFromConfig accepts service functional
options with validation. Token strings remain opaque; empty items alone do not end token
pagination. Page-only services retain their native page fields. Dedicated paginator
ClientOptions apply to each fetch and NextPage options override them for that fetch.
Duplicate continuation defaults to safely stopping after returning the current page;
an opt-out is explicit and may permit cycles. Failed/canceled fetches do not advance.
Wait returns error; WaitForOutput returns the successful typed response. Waiter input is
copied per invocation and per poll; overrides never mutate shared waiter configuration.

References: [AWS paginator](https://github.com/aws/aws-sdk-go-v2/blob/main/service/ec2/api_op_DescribeInstances.go),
[AWS waiter](https://github.com/aws/aws-sdk-go-v2/blob/main/service/ec2/api_op_DescribeInstanceStatus.go),
[AWS middleware](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/middleware.html).
Alibaba protocol evidence remains pinned under metadata/ with paired operation guides.

## 中文

本路径写在 abcf961 的 review 修复实现之前。Go 调用范式参考 AWS v2，操作名、线字段、token 值及
页语义遵循阿里云 OpenAPI。保留 Go 1.27、JSON v2、显式凭据、默认不重试、30 秒操作期限、HTTPS
规则和标准库核心。本项目独立，不承诺与两个上游 SDK 源码兼容。

| 顺序 | Issue 范围 | 验收 |
| --- | --- | --- |
| 1 | 尝试隔离及中断响应重试 | 不发布旧输出/不完整短路成功；EOF/读取失败仅幂等有界重试；取消/非幂等/JSON 边界保留 |
| 2 | 类型化 middleware 和服务 Options | Initialize/Serialize/Build 一次，Finalize/Deserialize 每尝试；输入/输出所有权；NewFromConfig、独立服务 Options；调用重试/transport 覆盖隔离 |
| 3 | 分页及生成策略集合 | 重复游标先返回当前页且停止策略可配置；页码防溢出；专属 options、NextPage 调用选项；多个纯 token/纯页码/双模式适配器且不造字段 |
| 4 | 可复用 waiter API | Wait/WaitForOutput 接收输入；并发等待独立；专属选项、默认 acceptor 可覆盖；保留全 ID 安全及取消 |

代码前已创建真实 GitHub issues，依赖顺序为 #26 → #27 → #28 → #29。
每 issue 独立分支/提交、测试及对应双语文档。通过模板再生成，不直接编辑生成文件。
全量本地门禁和同一提交的 Linux race/Windows CI 通过后关闭。
基准、ROA/body、默认凭据发现及默认重试/超时策略调整独立跟踪。

Review 已复现跨重试旧解码输出、重复 token 丢当前页、ECS 旧分页整数溢出及 io.ErrUnexpectedEOF
不重试；回归直接覆盖这些触发条件，并保留生成客户端和 ECS/STS/VPC 契约。

迁移说明区分增量及早期 v0 分页/waiter 修改。保留 New(Config)，NewFromConfig 提供服务 functional
options 及校验。Token 不解析，空条目不能单独结束 token 分页；纯页码保留服务字段。
Paginator ClientOptions 每页生效，NextPage 的当前选项优先。重复游标默认先交付当前页再安全停止，
显式禁用防护可能允许循环；失败/取消不前进。Wait 返回 error，WaitForOutput 返回成功输出。
每次等待及轮询复制输入，覆盖不得修改共享配置。AWS 来源与英文章节相同，阿里云协议证据固定在 metadata/。
