# ECS 使用指南

[English](ecs.md)

- 导入 `service/ecs`；#81 已移除旧 `services/ecs` 包。
- 固定的完整 DSL 流程在 #83 后生成 380 个 RPC 操作及完整可达模型。这不表示全部操作经过真实调用验证。
- 使用 `NewFromConfig`、以 context 为首个参数的操作方法，以及用于 mock 的小型操作接口。
- 可选标量指针区分缺失、零值和 false。响应保留原生容器，并包含 Metadata。
- DescribeInstances 默认使用原生 token；显式传入页码字段时使用页码模式，两种模式不能混用。
- DescribeInstances、DescribeInstanceStatus 和 DescribeImages 提供审核过的分页器；InstanceRunningWaiter 支持重复调用 Wait/WaitForOutput。
- 客户端复用公共运行时、凭据 provider/cache、显式启用的审核重试、中间件、结构化错误和可选遥测。
- 详见[生成指南](products/ecs.zh-CN.md)、[能力策略](capability-policy.zh-CN.md)、[消费者验收](ecs-product-acceptance.zh-CN.md)和[迁移说明](service-consolidation.zh-CN.md)。
