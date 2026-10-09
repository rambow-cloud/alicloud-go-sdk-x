# VPC usage

[中文](vpc.zh-CN.md)

- Import `service/vpc`; the old `services/vpc` package is removed in #81.
- The pinned full-DSL backend generates 296 supported RPC operations and complete reachable models. This is not all-action live coverage.
- Use `NewFromConfig`, context-first operations, copied inputs and small mock interfaces.
- Optional fields use pointers; response models retain native Vpcs.Vpc and other containers. Owner IDs retain int64 width.
- DescribeVpcsPaginator uses native page numbers, defaults to page one/size ten, and rejects sizes above fifty. Failed fetches preserve its cursor.
- No token cursor or VPC waiter is invented. Retry requires opt-in and a reviewed operation policy.
- Use shared Profile/STS providers, credentials.Cache, middleware, structured errors and optional tracing.
- See the [generated guide](products/vpc.md), [capability policy](capability-policy.md), [external consumer](../examples/productacceptance/README.md), [acceptance](vpc-product-acceptance.md) and [migration](service-consolidation.md).
