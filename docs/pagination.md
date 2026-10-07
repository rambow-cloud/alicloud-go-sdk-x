# Unified pagination / 统一分页

## English

Historical implementation was committed directly to main: shared engine `61d2581`
(issue #7), generated native adapters `89e1d07` (issue #28); neither issue is a PR.
The #36 full-DSL `service/` backend provides complete operation models and interfaces.
#37 adds sparse-policy native adapters for ECS DescribeInstances, DescribeInstanceStatus,
DescribeImages and VPC DescribeVpcs, reusing this engine. The earlier `services/`
reference adapters remain available. See the [policy specification](capability-policy.md).
New inputs use optional pointers: nil selects defaults; explicit PageNumber: 0 fails
the reviewed positive bound. DescribeImages defaults to page one/size ten (maximum 100).

`pagination.Paginator[T]` uses a typed fetcher, comparable Cursor and explicit HasMore.
Native token/page adapters decide termination from service metadata. Empty items can
continue with a token. Failed/canceled fetches preserve the cursor. Duplicate/cyclic
continuations return the current page successfully and then stop by default; the former
ErrRepeatedCursor diagnostic is deprecated. Explicit StopOnDuplicateCursor=false opts
out of protection and may allow unbounded cycles. Exhaustion returns ErrNoMorePages.
Paginators have one consumer and are not concurrency safe.

Generated constructors accept dedicated `OperationPaginatorOptions`: Limit overrides
native page size, StopOnDuplicateToken defaults to true, and copied ClientOptions apply
to every fetch. `NextPage(ctx, optFns...)` applies extra service options last for this
fetch only. Nil options fail without advancing. Example migration:

```go
pages, err := ecs.NewDescribeInstancesPaginator(client, input,
    func(o *ecs.DescribeInstancesPaginatorOptions) {
        o.Limit = 20
        o.ClientOptions = []func(*ecs.Options){func(call *ecs.Options) { call.Timeout = 5 * time.Second }}
    })
// Handle err, then call pages.NextPage(ctx, func(o *ecs.Options) { ... }).
```

ECS DescribeInstances defaults to native tokens with MaxResults=10; explicit PageNumber
or PageSize selects legacy pages, and mixing modes fails. Token completion ignores
TotalCount because the service declares it meaningless. DescribeInstanceStatus has its
own page-only paginator with size 50. VPC DescribeVpcs defaults to page one/size ten.
Page profiles stop on empty results or the total boundary, reject inconsistent metadata
and use division to avoid integer overflow. Inputs, pointers and slices are copied for
each fetch. No NextToken is invented for page-only APIs. See the
[official ECS paging contract](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeinstances).

## 中文

历史实现直接提交 main：共享引擎 `61d2581`（issue #7），原生生成适配器 `89e1d07`
（issue #28）；issue 编号不是 PR。#36 的 `service/` 后端交付完整操作模型和接口。
#37 新增 ECS DescribeInstances、DescribeInstanceStatus、DescribeImages 及 VPC
DescribeVpcs 的稀疏策略原生适配器，复用此引擎。原 `services/` 参考适配器仍可用。
见[策略规格](capability-policy.md)。新输入使用可选指针：nil 选择默认值，显式
PageNumber: 0 违反正数下限而失败。DescribeImages 默认页一/大小十，上限 100。

`pagination.Paginator[T]` 使用类型化 fetcher、可比较 Cursor 和明确 HasMore，根据服务 token/页元数据
决定终止。有 token 时空条目可继续；失败/取消保留游标。重复/循环游标默认成功返回当前页后停止，
旧 ErrRepeatedCursor 诊断已弃用；显式 StopOnDuplicateCursor=false 可禁用保护，可能出现无界循环。
耗尽返回 ErrNoMorePages。分页器单消费者，不保证并发安全。

构造选项改为专属 `OperationPaginatorOptions`：Limit 覆盖原生页大小，StopOnDuplicateToken 默认 true，
复制的 ClientOptions 每页生效。`NextPage(ctx, optFns...)` 的服务选项最后执行且只影响本页；nil
选项失败且不前进。迁移示例见英文章节，将原先构造时的服务选项移到 ClientOptions，或放在 NextPage。

ECS DescribeInstances 默认 MaxResults=10 的原生 token，显式 PageNumber/PageSize 选择旧页码且禁止混用。
Token 不使用服务声明无意义的 TotalCount。DescribeInstanceStatus 另有纯页码 paginator，默认大小 50；
VPC DescribeVpcs 默认页一/大小十。页码模式在空结果或到达总数时结束，拒绝不一致元数据并使用除法防溢出。
每次获取复制输入、指针和切片，不给纯页码 API 造 NextToken。
参考[官方 ECS 分页契约](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeinstances)。
