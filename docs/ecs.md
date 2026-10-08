# ECS reference client

[中文](ecs.zh-CN.md)

- `ecs.New(alicloud.Config)` supports three generated read operations in version 2014-05-26: DescribeRegions, DescribeInstances and DescribeInstanceStatus.
- Inputs use ordinary Go values; nil input means defaults.
- Outputs flatten service containers and carry transport Metadata.
- Individual operation interfaces accept the same context/options signature, so consumers can implement small fakes.
- Region overrides replace wire RegionId consistently.
- InstanceIDs use a JSON string for DescribeInstances and indexed InstanceId.N for DescribeInstanceStatus.
- DescribeInstances offers token or legacy page parameters, never both; selected response fields are ID/name/region/zone/status.
- Status requests support page sizes up to 50, ID lists up to 100.
- No full response or ECS API coverage is claimed.
- Unknown fields are ignored.
- Validation and runtime errors preserve OperationError and context causes.
- Tests use documented wire fixtures and never call cloud accounts.
- Protocol sources: [regions](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeregions), [instances](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeinstances), [status](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeinstancestatus).

- Generated field guidance: [ECS](generated/ecs.md); design: [generator](generator.md).

- `NewFromConfig` accepts service Options and `Client.Options()` returns a snapshot.
- DescribeInstances and DescribeInstanceStatus have dedicated paginator options and per-page NextPage options; the latter keeps native page numbers.
- InstanceRunningWaiter is reusable with Wait/WaitForOutput inputs supplied per invocation.
- See [pagination](pagination.md), [waiters](waiters.md) and [migration](aws-style-remediation.md).
