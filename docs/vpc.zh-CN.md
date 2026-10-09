# VPC 使用指南

[English](vpc.md)

- 导入 `service/vpc`；#81 已移除旧 `services/vpc` 包。
- 固定的完整 DSL 流程生成 296 个支持的 RPC 操作及完整可达模型。这不表示全部操作经过真实调用验证。
- 使用 `NewFromConfig`、以 context 为首个参数的操作方法、独立的输入副本和小型 mock 接口。
- 可选字段使用指针；响应保留 Vpcs.Vpc 等原生容器，Owner ID 保留 int64 宽度。
- DescribeVpcsPaginator 使用原生页码，默认从第 1 页开始、每页 10 条，拒绝超过 50 的页大小；请求失败时保留当前游标。
- 不添加服务不存在的 token 游标或 VPC waiter。重试需显式启用，并符合审核过的操作策略。
- 复用 Profile/STS provider、credentials.Cache、中间件、结构化错误和可选遥测。
- 详见[生成指南](products/vpc.zh-CN.md)、[能力策略](capability-policy.zh-CN.md)、[外部消费者](../examples/productacceptance/README.zh-CN.md)、[验收](vpc-product-acceptance.zh-CN.md)和[迁移说明](service-consolidation.zh-CN.md)。
