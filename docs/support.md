# Supported foundation

[中文](support.zh-CN.md)

- The `service/` full-DSL backend emits 283 ECS, 296 VPC and four STS actions with complete models and small operation interfaces; see [batch emission](batch-go-emission.md) and [product coverage](products/ecs.coverage.json).
- The matrix below records accepted shared foundation and the earlier `services/` reference adapters. #37 adds four `service/` paginators (including DescribeImages), InstanceRunningWaiter, five retry-safe reads, an AllocateDedicatedHosts token helper/validator and now sixteen sensitive STS models (#59).
- Nine actions have reviewed policy; the remaining 574 emitted actions remain unreviewed.
- See [policy](capability-policy.md); emission does not imply live acceptance.

- All entries have implementations, offline behavior tests, public Go documentation and external executable examples.
- APIs are early v0 contracts.
- This matrix records code coverage, not real-account acceptance or full parity with AWS SDK v2.

- Selected generated ECS/VPC reads have also passed the explicit local-profile comparison in [live validation #30](live-validation.md).
- That run ended each paginator on page one; waiter/AssumeRole and Explorer browser checks are separately documented as skipped/not run.

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

- RPC uses POST `/` with reviewed query encoding and ACS3.
- ROA path encoding is tested in the signer, but no ROA product client is delivered.
- Other regions require reviewed custom HTTPS endpoints.
- Other operations, complete models, OSS/SLS signing, streaming and automatic process/metadata credential discovery are outside this reference scope; native default Profile/OAuth loading is supported under #68.
- The [first generator profile](generator.md) produces these clients and reviewed paginator/waiter adapters from pinned metadata and overlays; prose-only validators remain handwritten.
- The [VPC guide](vpc.md) documents scalar presence, tag constraints and page-only traversal.
- Benchmarks #20 and broader generation remain separate.
- Release/indexing steps are in releasing.md.
