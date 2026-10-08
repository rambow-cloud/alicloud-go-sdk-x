# Supported foundation / 基础支持范围

## English

The `service/` full-DSL backend emits 283 ECS, 296 VPC and four STS actions with complete
models and small operation interfaces; see [batch emission](batch-go-emission.md) and
[product coverage](products/ecs.coverage.json). The matrix below records accepted shared
foundation and the earlier `services/` reference adapters. #37 adds four `service/`
paginators (including DescribeImages), InstanceRunningWaiter, five retry-safe reads,
an AllocateDedicatedHosts token helper/validator and now sixteen sensitive STS models (#59).
Nine actions have reviewed policy; the remaining 574 emitted actions remain unreviewed.
See [policy](capability-policy.md); emission does not imply live acceptance.

All entries have implementations, offline behavior tests, public Go documentation and external executable examples. APIs are early v0 contracts. This matrix records code coverage, not real-account acceptance or full parity with AWS SDK v2.

Selected generated ECS/VPC reads have also passed the explicit local-profile comparison
in [live validation #30](live-validation.md). That run ended each paginator on page one;
waiter/AssumeRole and Explorer browser checks are separately documented as skipped/not run.

| Capability          | Package / API                                                                                             | Scope and limits                                                                                                                                     |
| ------------------- | --------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| Unified paginator   | pagination.Paginator[T]; ECS DescribeInstances/DescribeInstanceStatusPaginator; VPC DescribeVpcsPaginator | native tokens/page numbers; dedicated/per-page options; current page delivered before cycle stop; single consumer                                    |
| Unified waiter      | waiter.Waiter[T]; ECS InstanceRunningWaiter                                                               | reusable Wait/WaitForOutput; concurrent input isolation; dedicated options/acceptor; bounded time; 1..50 distinct IDs                                |
| Retry/backoff       | retry.Standard                                                                                            | opt-in; jitter, Retry-After, budget; idempotent/replayable only                                                                                      |
| Mock interfaces     | HTTPClient, Provider, Resolver, Retryer, ECS/STS/VPC operation APIs                                       | small handwritten fakes                                                                                                                              |
| Credential provider | credentials static/env/Chain/Cache; config.LoadDefaultConfig and feature/profilecreds                     | explicit overrides, temporary environment and native CLI Profile/OAuth; bounded renewal/session persistence; no automatic process/metadata discovery |
| STS helper          | services/sts; feature/stscreds                                                                            | AssumeRole, expiration, separate source identity, cache composition                                                                                  |
| Endpoint resolver   | endpoint.Resolver/Rules                                                                                   | ECS/STS/VPC public cn-hangzhou, cn-shanghai, cn-beijing, cn-shenzhen, ap-southeast-1                                                                 |
| Middleware          | middleware.Stack                                                                                          | Initialize/Build once; Finalize/Deserialize per attempt; ordered hooks                                                                               |
| OpenTelemetry       | telemetry/otel                                                                                            | injected provider; operation/attempt spans; W3C propagation; no raw secrets                                                                          |
| Structured errors   | alicloud.OperationError/APIError/Metadata                                                                 | errors.Is/As; safe default cause/message formatting                                                                                                  |
| Testing helpers     | sdktest                                                                                                   | offline script, RoundTripper adapter, virtual clock                                                                                                  |

| Service version | Generated operation    | Selected output fields                                                         |
| --------------- | ---------------------- | ------------------------------------------------------------------------------ |
| ECS 2014-05-26  | DescribeRegions        | ID, localized name, endpoint, availability                                     |
| ECS 2014-05-26  | DescribeInstances      | ID/name/region/zone/status; token/page metadata                                |
| ECS 2014-05-26  | DescribeInstanceStatus | ID/status; page metadata                                                       |
| STS 2015-04-01  | AssumeRole             | keys/token/expiration, assumed user, source identity                           |
| VPC 2016-04-28  | DescribeVpcs           | IDs/CIDRs/bools/int64 owner, tags/IPv6 blocks/vSwitch ID subset; page metadata |

RPC uses POST `/` with reviewed query encoding and ACS3. ROA path encoding is tested in the signer, but no ROA product client is delivered. Other regions require reviewed custom HTTPS endpoints. Other operations, complete models, OSS/SLS signing, streaming and automatic process/metadata credential discovery are outside this reference scope; native default Profile/OAuth loading is supported under #68. The [first generator profile](generator.md) produces these clients and reviewed paginator/waiter adapters from pinned metadata and overlays; prose-only validators remain handwritten. The [VPC guide](vpc.md) documents scalar presence, tag constraints and page-only traversal. Benchmarks #20 and broader generation remain separate. Release/indexing steps are in releasing.md.

## 中文

完整 DSL 后端在 `service/` 输出 ECS 283、VPC 296、STS 4 个操作、完整模型及操作小
接口，见[批量输出](batch-go-emission.md)和[覆盖报告](products/ecs.coverage.json)。下表
记录已验收公共基础及原 `services/` 参考适配器。#37 在 `service/` 新增四个分页器
（含 DescribeImages）、InstanceRunningWaiter、五项安全读重试、AllocateDedicatedHosts
token helper/validator 与现有十六个 STS 敏感模型（#59）。九项操作已审核，其余 574 个已输出操作
仍未审核。见[策略](capability-policy.md)；已输出不等于真实验收。

每项均有实现、离线行为测试、公共 Go 文档和外部可运行示例。API 属于早期 v0 契约。矩阵记录代码覆盖，不代表真实账号验收或与 AWS SDK v2 完全等价。

选定生成 ECS/VPC 读取还通过了[真实验证 #30](live-validation.md) 的显式本地 Profile 对比。
该次分页均第一页结束；waiter/AssumeRole 及 Explorer 浏览器检查分别明确标记为跳过/未运行。

| 能力              | 包 / API                                                                                                  | 范围与限制                                                                               |
| ----------------- | --------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| 统一 paginator    | pagination.Paginator[T]；ECS DescribeInstances/DescribeInstanceStatusPaginator；VPC DescribeVpcsPaginator | 原生 token/页码、专属/每页选项、交付当前页后停止循环、单消费者                           |
| 统一 waiter       | waiter.Waiter[T]；ECS InstanceRunningWaiter                                                               | 可复用 Wait/WaitForOutput、并发输入隔离、专属选项/acceptor、总时长有界、1..50 个不同 ID  |
| retry/backoff     | retry.Standard                                                                                            | 显式启用，jitter、Retry-After、预算，仅幂等/可重放                                       |
| mock 接口         | HTTPClient、Provider、Resolver、Retryer、ECS/STS/VPC 操作 API                                             | 小型手写 fake                                                                            |
| 凭据 provider     | credentials static/env/Chain/Cache；config.LoadDefaultConfig、feature/profilecreds                        | 显式覆盖、临时环境及原生 CLI Profile/OAuth，有界续期/会话持久化，不自动发现进程/metadata |
| STS helper        | services/sts；feature/stscreds                                                                            | AssumeRole、过期时间、独立来源身份、cache 组合                                           |
| endpoint resolver | endpoint.Resolver/Rules                                                                                   | ECS/STS/VPC 公网 cn-hangzhou、cn-shanghai、cn-beijing、cn-shenzhen、ap-southeast-1       |
| middleware        | middleware.Stack                                                                                          | Initialize/Build 一次，Finalize/Deserialize 每尝试一次，有序 hook                        |
| OpenTelemetry     | telemetry/otel                                                                                            | 注入 provider、操作/尝试 span、W3C 传播、无原始秘密                                      |
| 结构化错误        | alicloud.OperationError/APIError/Metadata                                                                 | errors.Is/As，默认隐藏 cause/message 敏感文本                                            |
| testing helper    | sdktest                                                                                                   | 离线脚本、RoundTripper 适配、虚拟时钟                                                    |

| 服务版本       | 生成操作               | 选定输出字段                                                      |
| -------------- | ---------------------- | ----------------------------------------------------------------- |
| ECS 2014-05-26 | DescribeRegions        | ID、本地化名称、端点、可用性                                      |
| ECS 2014-05-26 | DescribeInstances      | ID/名称/地域/可用区/状态；token/页元数据                          |
| ECS 2014-05-26 | DescribeInstanceStatus | ID/状态；页元数据                                                 |
| STS 2015-04-01 | AssumeRole             | 密钥/token/过期、扮演用户、来源身份                               |
| VPC 2016-04-28 | DescribeVpcs           | ID/CIDR/布尔/int64 owner、tag/IPv6 块/vSwitch ID 子集；分页元数据 |

RPC 使用 POST `/`、核实的 query 编码和 ACS3。签名器测试 ROA 路径编码，但未交付 ROA 产品客户端。其他地域需核实的自定义 HTTPS 端点。其他操作、完整模型、OSS/SLS 签名、流式及自动进程/metadata 凭据发现不在参考范围内；#68 支持默认原生 Profile/OAuth 加载。[首版生成器](generator.md) 由固定元数据和 overlay 生成这些客户端及审核分页/waiter 适配器；仅说明中存在的校验保持手写。[VPC 指南](vpc.md) 说明标量存在语义、tag 约束与纯页码遍历。基准 #20 和更多生成独立跟踪。发布/索引步骤见 releasing.md。
