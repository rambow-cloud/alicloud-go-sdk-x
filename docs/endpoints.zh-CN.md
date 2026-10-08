# 端点解析

[English](endpoints.md)

- `endpoint.Resolver` 可替换；`ResolverFunc` 适配自定义函数，`NewRules` 复制显式规则。
- `BaseEndpoint` 优先。
- 地址必须是 HTTPS，不能包含凭据、非根路径、查询或片段。
- 未知规则返回 `ErrUnsupported`。
- 默认 ECS/STS/VPC 公网规则仅覆盖 cn-hangzhou、cn-shanghai、cn-beijing、cn-shenzhen、ap-southeast-1。
- 其他地域和特殊分区需要明确核实的规则；SDK 不推测主机名。
- 解析器共享，必须并发安全。
- 解析不访问网络。
- 依据为 [ECS 端点表](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-endpoint)和 [STS 端点表](https://help.aliyun.com/zh/ram/developer-reference/api-sts-2015-04-01-endpoint)和 [VPC 端点表](https://help.aliyun.com/zh/vpc/developer-reference/api-vpc-2016-04-28-endpoint)。
