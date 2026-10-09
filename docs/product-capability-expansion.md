# ECS/VPC capability expansion

[中文](product-capability-expansion.zh-CN.md)

- Issue #88; preserve official Darabonba, complete IR, shared backend and runtime.
- Add 21 native paginators: 9 ECS and 12 VPC. Totals: ECS 12 and VPC 13.
- Policy review covers ECS 14/380 and VPC 13/403 actions. Others remain unreviewed; complete adapter coverage is not claimed.
- Retry remains opt-in for these reviewed reads. Names or token fields do not make writes retryable.

## Added paginators

- Review pinned operation and pagination prose individually. Do not infer runtime policy from field names.
- Page APIs keep PageNumber/PageSize/TotalCount. Token APIs keep native NextToken and limit fields.
- New ECS token adapters reject deprecated page inputs. Raw operation calls still accept those fields.
- ListNatIps/ListNatIpCidrs keep string MaxResults; ListPrefixLists/ListVpcGatewayEndpoints keep int64. DescribeRouteEntryList uses native MaxResult.
- Parse decimal-string limits with bounds and safe diagnostics.

| Product | Action | Mode | Collection path | Default / maximum | Evidence |
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

## Waiters and verification

- Preserve InstanceRunningWaiter. Add InstanceStoppedWaiter: all requested IDs must be Stopped; Pending/Starting/Running/Stopping retry.
- Add VpcAvailableWaiter, VSwitchAvailableWaiter and NatGatewayAvailableWaiter. Filter a native scalar ID and force page one/size one. Missing results retry; unknown states or duplicate target rows fail.
- VPC/vSwitch Pending retries and Available succeeds. NAT Creating/Modifying/Converting retry; Available succeeds; Deleting fails.
- Deletion acceptance is outside this batch. Empty results can reflect propagation delay and do not establish deletion.
- Reuse bounded Wait/WaitForOutput, cancellation, virtual clocks and owned requests.
- Generate offline Examples for every added adapter. Independent fixtures check native wire fields, failed-cursor retention, empty token continuation, single-ID filtering, all-ID states and timeout.
- Live transitions remain #79. No resources are created or changed. Remaining pagination policy and deletion/lifecycle review stays open under #88.

## Additional evidence

- [VPC states](https://www.alibabacloud.com/help/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs)
- [vSwitch states](https://www.alibabacloud.com/help/en/vpc/developer-reference/api-vpc-2016-04-28-describevswitches)
- [ECS lifecycle](https://www.alibabacloud.com/help/en/ecs/user-guide/overview-52)
- [Native MaxResult](https://www.alibabacloud.com/help/en/vpc/developer-reference/api-vpc-2016-04-28-describerouteentrylist)
- NAT states use pinned DescribeNatGatewaysRequest.Status prose.

## Local checks

- Node 22 frontend check/84 tests, sdkgen check, doccheck, vet, root Go tests, formatting and documentation language checks PASS.
- Linux race and Windows use final PR CI; local tests do not establish live acceptance.
