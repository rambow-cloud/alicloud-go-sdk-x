# Unified pagination

[中文](pagination.zh-CN.md)

- Historical implementation was committed directly to main: shared engine `61d2581` (issue #7), generated native adapters `89e1d07` (issue #28); neither issue is a PR.
- The #36 full-DSL `service/` backend provides complete operation models and interfaces. #37 adds sparse-policy native adapters for ECS DescribeInstances, DescribeInstanceStatus, DescribeImages and VPC DescribeVpcs, reusing this engine.
- #81 removes the earlier services/ adapters; use generated service/ clients only.
- See the [policy specification](capability-policy.md).
- New inputs use optional pointers: nil selects defaults; explicit PageNumber: 0 fails the reviewed positive bound.
- DescribeImages defaults to page one/size ten (maximum 100).

- `pagination.Paginator[T]` uses a typed fetcher, comparable Cursor and explicit HasMore.
- Native token/page adapters decide termination from service metadata.
- Empty items can continue with a token.
- Failed/canceled fetches preserve the cursor.
- Duplicate/cyclic continuations return the current page successfully and then stop by default; the former ErrRepeatedCursor diagnostic is deprecated.
- Explicit StopOnDuplicateCursor=false opts out of protection and may allow unbounded cycles.
- Exhaustion returns ErrNoMorePages.
- Paginators have one consumer and are not concurrency safe.

- Generated constructors accept dedicated `OperationPaginatorOptions`: Limit overrides native page size, StopOnDuplicateToken defaults to true, and copied ClientOptions apply to every fetch. `NextPage(ctx, optFns...)` applies extra service options last for this fetch only.
- Nil options fail without advancing.
- Example migration:

```go
pages, err := ecs.NewDescribeInstancesPaginator(client, input,
    func(o *ecs.DescribeInstancesPaginatorOptions) {
        o.Limit = 20
        o.ClientOptions = []func(*ecs.Options){func(call *ecs.Options) { call.Timeout = 5 * time.Second }}
    })
// Handle err, then call pages.NextPage(ctx, func(o *ecs.Options) { ... }).
```

- ECS DescribeInstances defaults to native tokens with MaxResults=10; explicit PageNumber or PageSize selects legacy pages, and mixing modes fails.
- Token completion ignores TotalCount because the service declares it meaningless.
- DescribeInstanceStatus has its own page-only paginator with size 50.
- VPC DescribeVpcs defaults to page one/size ten.
- Page profiles stop on empty results or the total boundary, reject inconsistent metadata and use division to avoid integer overflow.
- Inputs, pointers and slices are copied for each fetch.
- No NextToken is invented for page-only APIs.
- See the [official ECS paging contract](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeinstances).
