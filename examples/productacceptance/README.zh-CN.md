# 产品消费者验收

[English](README.md)

- 这是独立的 Go 1.27 模块，只导入 SDK 公共包，通过 local replace 使用受评审的检出版本。
- 使用手写的合成响应和小型 mock，不需要云账号、真实凭据或网络连接。
- SDK 操作、模型、分页器和 waiter 来自固定官方 Darabonba 及已审核策略的生成流程。
- `imageIDs` 和 `instanceIDs` 只包含业务字段提取及标准 HasMorePages/NextPage 循环，不手动推进游标。

```powershell
go -C examples/productacceptance test ./...
go -C examples/productacceptance run .
```

## ECS 使用方式

- 通过 config.LoadDefaultConfig 和 ecs.NewFromConfig 构造客户端。原生临时凭据 Profile 或带缓存的 STS provider 可以由多个产品客户端共用。
- 原生页码分页使用 ecs.NewDescribeImagesPaginator。
- ecs.NewDescribeInstancesPaginator 中，MaxResults/NextToken 选择 token 模式，PageSize/PageNumber 选择页码模式；两种参数不能混用。
- 分页器复制输入，允许每页传入操作选项，只能由一个消费者使用。读取失败或取消时，游标不推进。
- 使用 ecs.NewInstanceRunningWaiter，再通过 Wait/WaitForOutput 指定全部目标 InstanceIDs 和正数超时，无需自行编写轮询循环。
- waiter 可以并发复用。缺失目标会继续等待，重复或未知观察会失败；超时可用 errors.Is 识别 waiter.ErrTimeout 和 context.DeadlineExceeded。
- 应用只需 mock 用到的 ecs.DescribeImagesAPI、ecs.DescribeInstancesAPI 或 ecs.DescribeInstanceStatusAPI。
- 用 errors.As 提取 alicloud.APIError/OperationError，保留 errors.Is 对取消的识别。Message 可能含敏感信息，需要显式处理。
- 重试需显式启用，只有审核过的读取策略允许暂时性错误重试；任意写入继续保守处理。
- OTel 使用应用持有的 provider，通过 telemetry/otel 中间件注入。导出属性不包含请求/响应体、凭据或错误消息。

## 范围与证据

- [ECS 验收计划](../../docs/ecs-product-acceptance.zh-CN.md)在执行前确定必需项。
- [VPC 验收计划](../../docs/vpc-product-acceptance.zh-CN.md)单独说明范围和真实调用限制。
- Go 测试耗时只表示自动化执行，不代表独立人工体验或性能。
- 完整生成清单、选定离线契约、历史真实调用的选定字段证据分别记录。
- local replace 用于验收；发布后的消费者应在 #61 完成后选择不可变 SDK 版本。
- 运行时和测试代码使用本项目 MIT 许可证；生成的产品包保留 Apache-2.0 来源及 LICENSE/NOTICE。

## VPC 使用方式

- 通过 vpc.NewFromConfig 使用与 ECS 相同的原生临时凭据 Profile 或带缓存的 STS 配置。
- 使用 vpc.NewDescribeVpcsPaginator 和 HasMorePages/NextPage，保留原生 PageNumber/PageSize 及 Vpcs.Vpc 容器，不添加 NextToken/MaxResults 或 waiter。
- `vpcIDs` 消费者只依赖 vpc.DescribeVpcsAPI，不包含手动页码推进或终止规则。
- 缺失/负数总数及错误的响应页码元数据会失败，且不消耗当前页。空集合会结束遍历，即使过时总数仍非零；非空短页按声明总数继续判断。
- Limit 最大为 50；显式零值、false 和缺失继续区分，嵌套响应容器保留原生结构。
- 分页器只能由一个消费者使用；客户端和 provider 按各自并发契约共用。通过每页操作选项覆盖，不修改客户端配置。
- 无请求模型操作 ListGeographicSubRegions 的离线用例与历史 DescribeVpcs 真实证据分别记录。
