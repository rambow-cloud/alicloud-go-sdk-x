# ECS/VPC 能力扩展

[English](product-capability-expansion.md)

- 对应 #88，沿用官方 Darabonba、完整 IR、公共生成后端与运行时。
- 新增 21 个原生分页器：ECS 9 个，VPC 12 个；总计 ECS 12 个、VPC 13 个。
- 策略审核覆盖 ECS 14/380、VPC 13/403 个操作；其余操作仍为未审核，不能声称全部能力已覆盖。
- 重试仍须显式启用，只对本表已审核的读取允许重放；写操作不因名称或 token 字段获得重试权限。

## 新增分页器

- 以下规则逐项审核固定 DSL 的操作说明及分页字段说明；不是运行时按字段名猜测。
- 页码接口保留 PageNumber/PageSize/TotalCount，token 接口保留原生 NextToken 及限制字段。
- 新增 ECS token 分页器拒绝已废弃的页码参数，避免混用；直接调用原始操作仍可使用这些字段。
- ListNatIps/ListNatIpCidrs 的 MaxResults 保留字符串类型；ListPrefixLists/ListVpcGatewayEndpoints 保留 int64；DescribeRouteEntryList 使用原生 MaxResult。
- 字符串限制值按十进制有界解析，无效内容不进入诊断信息。

| 服务 | 操作 | 模式 | 集合路径 | 默认 / 上限 | 依据 |
| --- | --- | --- | --- | --- | --- |
| ecs | DescribeDisks | tokens | Disks.Disk | 10 / 500 | [DSL:21017](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L21017) |
| ecs | DescribeSnapshots | tokens | Snapshots.Snapshot | 10 / 100 | [DSL:33550](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L33550) |
| ecs | DescribeSecurityGroups | tokens | SecurityGroups.SecurityGroup | 10 / 100 | [DSL:32545](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L32545) |
| ecs | DescribeNetworkInterfaces | tokens | NetworkInterfaceSets.NetworkInterfaceSet | 10 / 500 | [DSL:28797](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L28797) |
| ecs | DescribeKeyPairs | pages | KeyPairs.KeyPair | 10 / 50 | [DSL:27254](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L27254) |
| ecs | DescribeDeploymentSets | pages | DeploymentSets.DeploymentSet | 10 / 50 | [DSL:19970](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L19970) |
| ecs | DescribeAutoSnapshotPolicyEx | pages | AutoSnapshotPolicies.AutoSnapshotPolicy | 10 / 100 | [DSL:17479](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L17479) |
| ecs | DescribeCommands | tokens | Commands.Command | 10 / 50 | [DSL:19029](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L19029) |
| ecs | DescribeInvocations | tokens | Invocations.Invocation | 10 / 50 | [DSL:27070](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L27070) |
| vpc | DescribeVSwitches | pages | VSwitches.VSwitch | 10 / 50 | [DSL:29740](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L29740) |
| vpc | DescribeRouteTables | pages | RouteTables.RouteTable | 10 / 50 | [DSL:27792](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L27792) |
| vpc | DescribeRouteEntryList | tokens | RouteEntrys.RouteEntry | 10 / 100 | [DSL:27434](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L27434) |
| vpc | DescribeNatGateways | pages | NatGateways.NatGateway | 10 / 50 | [DSL:26304](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L26304) |
| vpc | DescribeEipAddresses | pages | EipAddresses.EipAddress | 10 / 100 | [DSL:22455](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L22455) |
| vpc | DescribeVpnGateways | pages | VpnGateways.VpnGateway | 10 / 50 | [DSL:32230](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L32230) |
| vpc | DescribeVpnConnections | pages | VpnConnections.VpnConnection | 10 / 50 | [DSL:31601](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L31601) |
| vpc | DescribeCustomerGateways | pages | CustomerGateways.CustomerGateway | 10 / 50 | [DSL:22158](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L22158) |
| vpc | ListPrefixLists | tokens | PrefixLists | 20 / 100 | [DSL:37669](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L37669) |
| vpc | ListNatIps | tokens | NatIps | 20 / 100 | [DSL:37387](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L37387) |
| vpc | ListNatIpCidrs | tokens | NatIpCidrs | 20 / 100 | [DSL:37224](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L37224) |
| vpc | ListVpcGatewayEndpoints | tokens | Endpoints | 20 / 100 | [DSL:39509](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L39509) |

## 等待与验证

- InstanceRunningWaiter 保留；新增 InstanceStoppedWaiter，只有全部指定实例均为 Stopped 才成功。Pending/Starting/Running/Stopping 继续等待。
- 新增 VpcAvailableWaiter、VSwitchAvailableWaiter 和 NatGatewayAvailableWaiter，使用原生单 ID 过滤条件，限制第一页一条结果；空结果继续等待，未知状态或重复目标失败。
- VPC 和交换机的 Pending 继续等待，Available 成功。NAT 的 Creating/Modifying/Converting 继续等待，Available 成功，Deleting 失败。
- 本范围不添加删除成功判定；空结果可能来自传播延迟，不能直接当作资源已删除。
- Wait/WaitForOutput 复用公共的有界轮询、取消、虚拟时间和输入复制实现。
- 生成所有新增适配器的离线 Example；独立测试验证真实线字段、失败后游标不前进、空 token 页继续、单 ID 过滤、全部实例状态及超时。
- 真实状态迁移仍由 #79 跟踪；本批不创建或修改云资源。剩余分页策略及删除等生命周期能力仍在 #88 下继续审核。

## 补充依据

- [VPC states](https://www.alibabacloud.com/help/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs)
- [vSwitch states](https://www.alibabacloud.com/help/en/vpc/developer-reference/api-vpc-2016-04-28-describevswitches)
- [ECS lifecycle](https://www.alibabacloud.com/help/en/ecs/user-guide/overview-52)
- [Native MaxResult](https://www.alibabacloud.com/help/en/vpc/developer-reference/api-vpc-2016-04-28-describerouteentrylist)
- NAT 状态取自固定 DSL 的 DescribeNatGatewaysRequest.Status 说明。

## 本地检查

- Node 22 前端检查及 84 项测试、sdkgen check、doccheck、vet、完整 Go 测试、格式及文档语言检查通过。
- Linux race 与 Windows 以当前 PR 的 CI 结果为准；真实云验收不在本地测试结果中。
