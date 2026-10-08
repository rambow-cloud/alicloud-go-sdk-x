# ECS 参考客户端

[English](ecs.md)

- `ecs.New(alicloud.Config)` 支持 2014-05-26 版本三个生成只读操作：DescribeRegions、DescribeInstances、DescribeInstanceStatus。
- 输入使用普通 Go 值；nil 输入表示默认值。
- 输出展平服务容器并携带传输 Metadata。
- 单操作接口使用相同 context/options 签名，消费者可实现小型 fake。
- 地域覆盖一致替换线协议 RegionId。
- InstanceIDs 在 DescribeInstances 编码为 JSON 字符串，在 DescribeInstanceStatus 编码为 InstanceId.N 索引参数。
- DescribeInstances 可用 token 或旧页码参数，但不得混用；响应仅选择 ID/名称/地域/可用区/状态字段。
- 状态请求最多每页 50 个，ID 列表最多 100 个。
- 不表示完整响应或 ECS API 覆盖。
- 忽略未知字段。
- 验证和运行时错误保留 OperationError 与 context 底层错误。
- 测试使用文档线协议测试数据，绝不调用云账号。
- 协议来源：[地域](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeregions)、[实例](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeinstances)、[状态](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeinstancestatus)。

- 生成字段说明：[ECS](generated/ecs.md)；设计：[生成器](generator.zh-CN.md)。

- NewFromConfig 接收服务 Options，Client.Options() 返回快照。
- DescribeInstances 与 DescribeInstanceStatus 提供专属分页器选项及每页 NextPage 选项，后者保留原生页码。
- InstanceRunningWaiter 可复用，每次 Wait/WaitForOutput 提供输入；见[分页](pagination.zh-CN.md)、 [状态等待器](waiters.zh-CN.md) 及[迁移](aws-style-remediation.zh-CN.md)。
