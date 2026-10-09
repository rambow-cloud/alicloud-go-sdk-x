# ECS usage

[中文](ecs.zh-CN.md)

- Import `service/ecs`; the old `services/ecs` package is removed in #81.
- The pinned full-DSL backend generates 380 RPC operations after #83 and complete reachable models. This is not all-action live coverage.
- Use `NewFromConfig`, context-first operations and small operation interfaces for mocks.
- Optional scalar pointers preserve absence, zero and false. Outputs retain native response containers and include Metadata.
- DescribeInstances uses native tokens by default; explicit page fields select page mode. Do not mix both modes.
- DescribeInstances, DescribeInstanceStatus and DescribeImages have reviewed paginators. InstanceRunningWaiter supports reusable Wait/WaitForOutput.
- Clients share the runtime, providers/cache, reviewed opt-in retry, middleware, structured errors and optional tracing.
- See the [generated guide](products/ecs.md), [capability policy](capability-policy.md), [consumer acceptance](ecs-product-acceptance.md) and [migration](service-consolidation.md).
