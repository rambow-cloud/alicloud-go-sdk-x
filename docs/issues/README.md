# Issue index / Issue 索引

## English

The first release is [v0.1.0 STS](../sts-v0.1.0.md), tracked by
[milestone v0.1.0](https://github.com/rambow-cloud/alicloud-go-sdk-x/milestone/4)
and [Project 3](https://github.com/orgs/rambow-cloud/projects/3). Native child issues
belong to #57; parent membership is not a blocking dependency.

| Issue | v0.1.0 work                                                                 | Dependencies                                              |
| ----- | --------------------------------------------------------------------------- | --------------------------------------------------------- |
| #57   | [Scoped STS release parent](57-sts-v010-roadmap.md)                         | Aggregates #58-#61; remains open through release/indexing |
| #58   | [Requestless GetCallerIdentity](58-requestless-sts-generation.md)           | Accepted generator baseline                               |
| #59   | [Anonymous OIDC/SAML RPC](59-anonymous-sts-generation.md)                   | Accepted generator baseline; independent of #58           |
| #60   | [Four-operation acceptance and source rehearsal](60-sts-v010-acceptance.md) | #58, #59, accepted #51/#53/#55                            |
| #61   | [Release and pkg.go.dev](61-sts-v010-release.md)                            | #60                                                       |

GitHub owns current state; this table records historical and current dependencies, not
current blocking state. The authoritative full-DSL route is tracked by #33, with
#34 -> #35 -> #36 -> #37 -> #38. #31 remains the compatibility bridge. Older accepted
foundation/generator issues are history, not new per-operation prerequisites.

| Issue | Capability                                                                                   | Dependencies                                                        |
| ----- | -------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| #10   | [Docs]: Define the runtime-first development path and bilingual documentation policy         | None                                                                |
| #11   | [Feature]: Implement a shared staged middleware pipeline                                     | None                                                                |
| #12   | [Feature]: Implement replaceable endpoint resolution with explicit rules                     | None                                                                |
| #13   | [Feature]: Provide structured operation errors and response metadata                         | None                                                                |
| #14   | [Feature]: Implement a unified bounded waiter engine and ECS running waiter                  | #5                                                                  |
| #15   | [Feature]: Provide small mock interfaces and deterministic testing helpers                   | None                                                                |
| #16   | [Feature]: Add a typed STS AssumeRole client and credential provider helper                  | #3, #9                                                              |
| #17   | [Feature]: Add optional OpenTelemetry operation and attempt instrumentation                  | #3, #11, #13                                                        |
| #18   | [Feature]: Add an explicit composable credential chain                                       | None                                                                |
| #19   | [Maintenance]: Validate the unified runtime foundation before enabling generator development | #3, #4, #5, #6, #7, #9, #10, #11, #12, #13, #14, #15, #16, #17, #18 |
| #20   | [Maintenance]: Establish reproducible runtime and dependency comparison benchmarks           | #19                                                                 |
| #21   | Pinned metadata importer and validated IR                                                    | #19                                                                 |
| #22   | Generated clients, models and paginator/waiter adapters                                      | #21                                                                 |
| #23   | Deterministic regeneration and CI acceptance                                                 | #22                                                                 |
| #24   | Scalar presence, object lists, local references and page-only generation                     | #23                                                                 |
| #25   | Generated VPC DescribeVpcs client and paginator                                              | #24                                                                 |
| #26   | Isolated attempt output and interrupted response retries                                     | #25                                                                 |
| #27   | Typed middleware and concrete service Options                                                | #26                                                                 |
| #28   | Native paginator options and multiple policy profiles                                        | #27                                                                 |
| #29   | Reusable Wait/WaitForOutput and waiter options                                               | #28                                                                 |
| #30   | Live local-profile reads and Explorer CLI comparison                                         | #29                                                                 |
| #31   | Official Darabonba semantic frontend and reviewed DSL/metadata decisions                     | #29, #30                                                            |
| #33   | Full-DSL product generator roadmap (parent)                                                  | Stages #34-#38                                                      |
| #34   | Source representation normalization                                                          | #31                                                                 |
| #35   | Complete DSL discovery, reachable IR and coverage                                            | #34                                                                 |
| #36   | Batch Go emission from product IR                                                            | #35                                                                 |
| #37   | Sparse capability policies                                                                   | #36                                                                 |
| #38   | Licensed bilingual/pkg.go.dev documentation automation                                       | #37                                                                 |
| #44   | Reject aliased pagination and waiter policy roles                                            | #38                                                                 |
| #49   | Scoped product Beta/release criteria and developer experience definition                     | #33, #44; #47 evidence tracked separately                           |
| #51   | Full-DSL STS client -> provider/cache -> generated consumer composition                      | #49; accepted foundation and generator                              |
| #53   | STS-first application guidance and explicit credential-provider configuration                | #49, #51                                                            |
| #55   | Dedicated-role live STS issuance, cache reuse, forced refresh and real expiry renewal        | #49, #51, #53                                                       |

## 中文

首版为 [v0.1.0 STS](../sts-v0.1.0.md)，由 [同名 milestone](https://github.com/rambow-cloud/alicloud-go-sdk-x/milestone/4)
和 [Project 3](https://github.com/orgs/rambow-cloud/projects/3) 跟踪，原生子项归属 #57，
父项归属不作为阻塞依赖。

| Issue | v0.1.0 工作                                                  | 依赖                              |
| ----- | ------------------------------------------------------------ | --------------------------------- |
| #57   | [限定 STS 发布父项](57-sts-v010-roadmap.md)                  | 汇总 #58-#61，发布/索引前保持打开 |
| #58   | [无请求 GetCallerIdentity](58-requestless-sts-generation.md) | 已接受生成器基线                  |
| #59   | [匿名 OIDC/SAML RPC](59-anonymous-sts-generation.md)         | 已接受生成器基线，独立于 #58      |
| #60   | [四操作验收与来源升级](60-sts-v010-acceptance.md)            | #58、#59、已完成 #51/#53/#55      |
| #61   | [发布与 pkg.go.dev](61-sts-v010-release.md)                  | #60                               |

Review 修复按 #26 → #27 → #28 → #29 执行，详见 [修复路径](../aws-style-remediation.md)。
对应草稿记录每尝试隔离、类型化 pipeline、分页策略集合和可复用 waiter 验收。

状态以 GitHub 为准，本表记录历史/当前依赖，不表示当前阻塞状态。权威完整 DSL
路线由 #33 跟踪，按 #34 → #35 → #36 → #37 → #38 执行；#31 保留为兼容桥，已验收
基础/生成任务为历史，不作为新增逐操作前置要求。

| Issue | 能力                                                                                              | 依赖                                                                |
| ----- | ------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| #10   | 交付语义对应的中英文开发路径与项目文档。                                                          | None                                                                |
| #11   | 提供 Initialize、Serialize、Build、Finalize、Deserialize 阶段，以及具名 middleware 和函数适配器。 | None                                                                |
| #12   | 提供 context-aware resolver 和函数适配器；显式 endpoint 优先。                                    | None                                                                |
| #13   | 保留 APIError，增加可 Unwrap 的操作错误。                                                         | None                                                                |
| #14   | 提供类型化轮询及成功、重试、失败 acceptor。                                                       | #5                                                                  |
| #15   | 提供操作级小接口和可 mock 的分页/waiter 契约。                                                    | None                                                                |
| #16   | 使用共同 runtime 实现官方 STS AssumeRole RPC。                                                    | #3, #9                                                              |
| #17   | 提供注入 TracerProvider 的可选 middleware 适配器。                                                | #3, #11, #13                                                        |
| #18   | 区分凭据来源不存在与来源不完整/无效。                                                             | None                                                                |
| #19   | 用手写 ECS/STS 参考验证十一项能力。                                                               | #3, #4, #5, #6, #7, #9, #10, #11, #12, #13, #14, #15, #16, #17, #18 |
| #20   | 固定 Go、OS、架构和官方 SDK 版本。                                                                | #19                                                                 |
| #21   | 固定元数据导入及校验 IR。                                                                         | #19                                                                 |
| #22   | 生成客户端、模型及分页/waiter 适配器。                                                            | #21                                                                 |
| #23   | 确定性再生成与 CI 验收。                                                                          | #22                                                                 |
| #24   | 标量存在语义、对象数组、本地引用及纯页码生成。                                                    | #23                                                                 |
| #25   | 生成 VPC DescribeVpcs 客户端及分页。                                                              | #24                                                                 |
| #26   | 尝试输出隔离与中断响应重试。                                                                      | #25                                                                 |
| #27   | 类型化 middleware 和独立服务 Options。                                                            | #26                                                                 |
| #28   | 原生分页选项及多策略 profile。                                                                    | #27                                                                 |
| #29   | 可复用 Wait/WaitForOutput 和 waiter 选项。                                                        | #28                                                                 |
| #30   | 本地 Profile 真实读取及 Explorer CLI 对比。                                                       | #29                                                                 |
| #31   | 官方 Darabonba 语义前端与审核后的 DSL/元数据决策。                                                | #29, #30                                                            |
| #33   | 完整 DSL 产品生成路线父任务。                                                                     | 阶段 #34-#38                                                        |
| #34   | 来源表示规范化。                                                                                  | #31                                                                 |
| #35   | 完整 DSL 发现、可达 IR 及覆盖。                                                                   | #34                                                                 |
| #36   | 产品 IR 批量 Go 输出。                                                                            | #35                                                                 |
| #37   | 少量能力策略。                                                                                    | #36                                                                 |
| #38   | 授权双语/pkg.go.dev 文档自动化。                                                                  | #37                                                                 |
| #44   | 拒绝分页与 waiter 策略的同模型角色重叠。                                                          | #38                                                                 |
| #49   | 限定产品 Beta/发布标准及开发者体验定义。                                                          | #33、#44；#47 证据独立跟踪                                          |
| #51   | 完整 DSL STS client→provider/cache→生成消费者组合。                                               | #49；已接受基础和生成器                                             |
| #53   | STS 优先应用指南与显式凭据 provider 配置。                                                        | #49, #51                                                            |
| #55   | 专用角色真实 STS 签发、缓存复用、强制刷新和真实到期续期。                                         | #49, #51, #53                                                       |
