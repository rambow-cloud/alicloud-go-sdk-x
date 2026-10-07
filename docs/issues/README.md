# Issue index / Issue 索引

## English

GitHub owns current state; this table records dependencies. Existing #3 HTTP depends on #4/#11/#12/#13/#15; #5 ECS on #3; #7 paginator integration on #5; #9 cache follows #18; #8 generator is blocked by #19. #4 signing and #6 retry are independently testable.

| Issue | Capability | Dependencies |
| --- | --- | --- |
| #10 | [Docs]: Define the runtime-first development path and bilingual documentation policy | None |
| #11 | [Feature]: Implement a shared staged middleware pipeline | None |
| #12 | [Feature]: Implement replaceable endpoint resolution with explicit rules | None |
| #13 | [Feature]: Provide structured operation errors and response metadata | None |
| #14 | [Feature]: Implement a unified bounded waiter engine and ECS running waiter | #5 |
| #15 | [Feature]: Provide small mock interfaces and deterministic testing helpers | None |
| #16 | [Feature]: Add a typed STS AssumeRole client and credential provider helper | #3, #9 |
| #17 | [Feature]: Add optional OpenTelemetry operation and attempt instrumentation | #3, #11, #13 |
| #18 | [Feature]: Add an explicit composable credential chain | None |
| #19 | [Maintenance]: Validate the unified runtime foundation before enabling generator development | #3, #4, #5, #6, #7, #9, #10, #11, #12, #13, #14, #15, #16, #17, #18 |
| #20 | [Maintenance]: Establish reproducible runtime and dependency comparison benchmarks | #19 |
| #21 | Pinned metadata importer and validated IR | #19 |
| #22 | Generated clients, models and paginator/waiter adapters | #21 |
| #23 | Deterministic regeneration and CI acceptance | #22 |
| #24 | Scalar presence, object lists, local references and page-only generation | #23 |
| #25 | Generated VPC DescribeVpcs client and paginator | #24 |
| #26 | Isolated attempt output and interrupted response retries | #25 |
| #27 | Typed middleware and concrete service Options | #26 |
| #28 | Native paginator options and multiple policy profiles | #27 |
| #29 | Reusable Wait/WaitForOutput and waiter options | #28 |

## 中文

Review 修复按 #26 → #27 → #28 → #29 执行，详见 [修复路径](../aws-style-remediation.md)。
对应草稿记录每尝试隔离、类型化 pipeline、分页策略集合和可复用 waiter 验收。

状态以 GitHub 为准，本表记录依赖。已有 #3 HTTP 依赖 #4/#11/#12/#13/#15，#5 ECS 依赖 #3，#7 分页集成依赖 #5，#9 缓存在 #18 后完成，#8 generator 被 #19 阻塞。#4 签名及 #6 重试可独立测试。

| Issue | 能力 | 依赖 |
| --- | --- | --- |
| #10 | 交付语义对应的中英文开发路径与项目文档。 | None |
| #11 | 提供 Initialize、Serialize、Build、Finalize、Deserialize 阶段，以及具名 middleware 和函数适配器。 | None |
| #12 | 提供 context-aware resolver 和函数适配器；显式 endpoint 优先。 | None |
| #13 | 保留 APIError，增加可 Unwrap 的操作错误。 | None |
| #14 | 提供类型化轮询及成功、重试、失败 acceptor。 | #5 |
| #15 | 提供操作级小接口和可 mock 的分页/waiter 契约。 | None |
| #16 | 使用共同 runtime 实现官方 STS AssumeRole RPC。 | #3, #9 |
| #17 | 提供注入 TracerProvider 的可选 middleware 适配器。 | #3, #11, #13 |
| #18 | 区分凭据来源不存在与来源不完整/无效。 | None |
| #19 | 用手写 ECS/STS 参考验证十一项能力。 | #3, #4, #5, #6, #7, #9, #10, #11, #12, #13, #14, #15, #16, #17, #18 |
| #20 | 固定 Go、OS、架构和官方 SDK 版本。 | #19 |
| #21 | 固定元数据导入及校验 IR。 | #19 |
| #22 | 生成客户端、模型及分页/waiter 适配器。 | #21 |
| #23 | 确定性再生成与 CI 验收。 | #22 |
| #24 | 标量存在语义、对象数组、本地引用及纯页码生成。 | #23 |
| #25 | 生成 VPC DescribeVpcs 客户端及分页。 | #24 |
| #26 | 尝试输出隔离与中断响应重试。 | #25 |
| #27 | 类型化 middleware 和独立服务 Options。 | #26 |
| #28 | 原生分页选项及多策略 profile。 | #27 |
| #29 | 可复用 Wait/WaitForOutput 和 waiter 选项。 | #28 |
