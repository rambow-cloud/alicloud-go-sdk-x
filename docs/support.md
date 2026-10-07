# Supported foundation / 基础支持范围

## English

All entries have implementations, offline behavior tests, public Go documentation and external executable examples. APIs are early v0 contracts. This matrix records code coverage, not real-account acceptance or full parity with AWS SDK v2.

| Capability | Package / API | Scope and limits |
| --- | --- | --- |
| Unified paginator | pagination.Paginator[T]; ECS DescribeInstancesPaginator | tokens/page numbers; cycle protection; single consumer |
| Unified waiter | waiter.Waiter[T]; ECS InstanceRunningWaiter | bounded total time; explicit acceptors; 1..50 distinct reference IDs |
| Retry/backoff | retry.Standard | opt-in; jitter, Retry-After, budget; idempotent/replayable only |
| Mock interfaces | HTTPClient, Provider, Resolver, Retryer, ECS/STS operation APIs | small handwritten fakes |
| Credential provider | credentials static/env/Chain/Cache | explicit order; bounded shared refresh; no implicit file/process/metadata discovery |
| STS helper | services/sts; feature/stscreds | AssumeRole, expiration, separate source identity, cache composition |
| Endpoint resolver | endpoint.Resolver/Rules | ECS/STS public cn-hangzhou, cn-shanghai, cn-beijing, cn-shenzhen, ap-southeast-1 |
| Middleware | middleware.Stack | Initialize/Build once; Finalize/Deserialize per attempt; ordered hooks |
| OpenTelemetry | telemetry/otel | injected provider; operation/attempt spans; W3C propagation; no raw secrets |
| Structured errors | alicloud.OperationError/APIError/Metadata | errors.Is/As; safe default cause/message formatting |
| Testing helpers | sdktest | offline script, RoundTripper adapter, virtual clock |

| Service version | Generated operation | Selected output fields |
| --- | --- | --- |
| ECS 2014-05-26 | DescribeRegions | ID, localized name, endpoint, availability |
| ECS 2014-05-26 | DescribeInstances | ID/name/region/zone/status; token/page metadata |
| ECS 2014-05-26 | DescribeInstanceStatus | ID/status; page metadata |
| STS 2015-04-01 | AssumeRole | keys/token/expiration, assumed user, source identity |

RPC uses POST `/` with reviewed query encoding and ACS3. ROA path encoding is tested in the signer, but no ROA product client is delivered. Other regions require reviewed custom HTTPS endpoints. Other operations, complete models, OSS/SLS signing, streaming and automatic credential discovery are outside scope. The [first generator profile](generator.md) produces these clients and reviewed paginator/waiter adapters from pinned metadata and overlays; prose-only validators remain handwritten. Benchmarks #20 and broader generation remain separate. Release/indexing steps are in releasing.md.

## 中文

每项均有实现、离线行为测试、公共 Go 文档和外部可运行示例。API 属于早期 v0 契约。矩阵记录代码覆盖，不代表真实账号验收或与 AWS SDK v2 完全等价。

| 能力 | 包 / API | 范围与限制 |
| --- | --- | --- |
| 统一 paginator | pagination.Paginator[T]；ECS DescribeInstancesPaginator | token/页码，循环防护，单消费者 |
| 统一 waiter | waiter.Waiter[T]；ECS InstanceRunningWaiter | 总时长有界，明确 acceptor，参考支持 1..50 个不同 ID |
| retry/backoff | retry.Standard | 显式启用，jitter、Retry-After、预算，仅幂等/可重放 |
| mock 接口 | HTTPClient、Provider、Resolver、Retryer、ECS/STS 操作 API | 小型手写 fake |
| 凭据 provider | credentials static/env/Chain/Cache | 明确顺序，共享有界刷新，无隐式文件/进程/metadata 发现 |
| STS helper | services/sts；feature/stscreds | AssumeRole、过期时间、独立来源身份、cache 组合 |
| endpoint resolver | endpoint.Resolver/Rules | ECS/STS 公网 cn-hangzhou、cn-shanghai、cn-beijing、cn-shenzhen、ap-southeast-1 |
| middleware | middleware.Stack | Initialize/Build 一次，Finalize/Deserialize 每尝试一次，有序 hook |
| OpenTelemetry | telemetry/otel | 注入 provider、操作/尝试 span、W3C 传播、无原始秘密 |
| 结构化错误 | alicloud.OperationError/APIError/Metadata | errors.Is/As，默认隐藏 cause/message 敏感文本 |
| testing helper | sdktest | 离线脚本、RoundTripper 适配、虚拟时钟 |

| 服务版本 | 生成操作 | 选定输出字段 |
| --- | --- | --- |
| ECS 2014-05-26 | DescribeRegions | ID、本地化名称、端点、可用性 |
| ECS 2014-05-26 | DescribeInstances | ID/名称/地域/可用区/状态；token/页元数据 |
| ECS 2014-05-26 | DescribeInstanceStatus | ID/状态；页元数据 |
| STS 2015-04-01 | AssumeRole | 密钥/token/过期、扮演用户、来源身份 |

RPC 使用 POST `/`、核实的 query 编码和 ACS3。签名器测试 ROA 路径编码，但未交付 ROA 产品客户端。其他地域需核实的自定义 HTTPS 端点。其他操作、完整模型、OSS/SLS 签名、流式、自动凭据发现不在范围内。[首版生成器](generator.md) 由固定元数据和 overlay 生成这些客户端及审核分页/waiter 适配器；仅说明中存在的校验保持手写。基准 #20 和更多生成独立跟踪。发布/索引步骤见 releasing.md。
