# 统一分页

[English](pagination.md)

- 历史实现直接提交 main：共享引擎 `61d2581`（issue #7），原生生成适配器 `89e1d07` （issue #28）；issue 编号不是 PR。
- #36 的 `service/` 后端交付完整操作模型和接口。
- #37 新增 ECS DescribeInstances、DescribeInstanceStatus、DescribeImages 及 VPC DescribeVpcs 的稀疏策略原生适配器，复用此引擎。
- 原 `services/` 参考适配器仍可用。
- 见[策略规格](capability-policy.zh-CN.md)。
- 新输入使用可选指针：nil 选择默认值，显式 PageNumber: 0 违反正数下限而失败。
- DescribeImages 默认页一/大小十，上限 100。

- `pagination.Paginator[T]` 使用强类型取页函数、可比较 Cursor 和明确 HasMore，根据服务 token/页元数据决定终止。
- 有 token 时空条目可继续；失败/取消保留游标。
- 遇到重复或循环游标时，默认返回当前页，然后停止， 旧 ErrRepeatedCursor 诊断已弃用；显式 StopOnDuplicateCursor=false 可禁用保护，可能出现无界循环。
- 耗尽返回 ErrNoMorePages。
- 分页器单消费者，不保证并发安全。

- 构造选项改为专属 `OperationPaginatorOptions`：Limit 覆盖原生页大小，StopOnDuplicateToken 默认 true， 复制的 ClientOptions 每页生效。
- `NextPage(ctx, optFns...)` 的服务选项最后执行且只影响本页；nil 选项失败且不前进。
- 迁移时，把构造函数原有的服务选项放入 `ClientOptions`，或在调用 `NextPage` 时传入。下方示例展示前一种方式。

- ECS DescribeInstances 默认 MaxResults=10 的原生 token，显式 PageNumber/PageSize 选择旧页码且禁止混用。
- Token 不使用服务声明无意义的 TotalCount。
- DescribeInstanceStatus 另有纯页码分页器，默认大小 50； VPC DescribeVpcs 默认页一/大小十。
- 页码模式在空结果或到达总数时结束，拒绝不一致元数据并使用除法防溢出。
- 每次获取复制输入、指针和切片，不给纯页码 API 造 NextToken。
- 参考[官方 ECS 分页契约](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeinstances)。

## 可运行命令与示例

```go
pages, err := ecs.NewDescribeInstancesPaginator(client, input,
    func(o *ecs.DescribeInstancesPaginatorOptions) {
        o.Limit = 20
        o.ClientOptions = []func(*ecs.Options){func(call *ecs.Options) { call.Timeout = 5 * time.Second }}
    })
// Handle err, then call pages.NextPage(ctx, func(o *ecs.Options) { ... }).
```
