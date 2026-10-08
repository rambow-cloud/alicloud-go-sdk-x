# Endpoint resolution

[中文](endpoints.zh-CN.md)

- `endpoint.Resolver` is replaceable; `ResolverFunc` adapts custom functions and `NewRules` copies explicit rules. `BaseEndpoint` takes precedence.
- Origins must use HTTPS, with no credentials, non-root path, query or fragment.
- Unknown rules return `ErrUnsupported`.
- Default public ECS/STS/VPC rules cover only cn-hangzhou, cn-shanghai, cn-beijing, cn-shenzhen and ap-southeast-1.
- Other regions and special partitions need explicit reviewed rules; the SDK never infers their hostnames.
- Resolvers are shared and must be concurrency safe.
- No resolution network calls occur.
- Rules come from the [ECS endpoint table](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-endpoint) and [STS endpoint table](https://help.aliyun.com/zh/ram/developer-reference/api-sts-2015-04-01-endpoint) and [VPC endpoint table](https://help.aliyun.com/zh/vpc/developer-reference/api-vpc-2016-04-28-endpoint).
