# VPC usage / VPC 用法

## English

`services/vpc` provides the generated `DescribeVpcs` client, selected models and
`DescribeVpcsPaginator` for API 2016-04-28. It uses the shared runtime for ACS3 signing,
credentials, opt-in retries, middleware, structured errors and optional tracing. This is
one reviewed operation, with selected fields. See the [generated API guide](generated/vpc.md)
and [expansion design](generator-expansion.md).

Create a client with `vpc.New(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: provider})`.
`RegionID` overrides the configured region for that call. Default public endpoints cover
cn-hangzhou, cn-shanghai, cn-beijing, cn-shenzhen and ap-southeast-1. Other regions require
an explicit reviewed HTTPS endpoint or resolver. Clients may be shared concurrently;
callers must not mutate inputs concurrently. Inputs, slices and pointer members are copied
before middleware and separately for every paginator request.

Optional boolean filters use `*bool`: `IPv6Enabled: nil` means no IPv6 filter,
`new(false)` selects disabled and `new(true)` selects enabled. A nil `IsDefault` delegates
omission/default behavior to the service. `DryRun: new(true)` requests preflight;
the service's successful preflight code `DryRunOperation` is returned as `alicloud.APIError`,
wrapped by `OperationError`; inspect it with `errors.As`.

Tag filters encode as `Tag.N.Key` and `Tag.N.Value`. At most twenty are allowed. Keys
must be nonempty; keys and supplied values accept at most 128 Unicode characters,
reject invalid UTF-8, prefixes `aliyun`/`acs:` and substrings `http://`/`https://`.
`Value: nil` omits the value; `Value: new("")` sends an explicit empty value. These
prose-only rules remain in a handwritten validator. `VPCID` accepts up to twenty
comma-separated IDs. Negative owner IDs and page parameters are rejected.

The paginator uses page numbers only, starting at page one with size ten when zero.
Sizes above fifty are rejected. It stops on an empty page or when the total is reached;
service failures, cancellation, nil output and inconsistent page metadata preserve the
current cursor. Returned page size zero falls back to the requested size. Traverse with
one consumer and check `HasMorePages` before `NextPage(ctx)`; exhaustion returns
`pagination.ErrNoMorePages`. Changing data during traversal can change the service totals;
the paginator does not promise a snapshot or deduplication.

This offline usage is exercised by external package Examples:

```go
transport := sdktest.NewTransport(sdktest.Step{
    Body: `{"PageNumber":1,"PageSize":10,"TotalCount":1,"Vpcs":{"Vpc":[{"VpcId":"vpc-example"}]}}`,
})
client, err := vpc.New(alicloud.Config{
    Region: "cn-hangzhou", CredentialsProvider: provider,
    HTTPClient: &http.Client{Transport: transport},
})
if err != nil { panic(err) }
pages, err := vpc.NewDescribeVpcsPaginator(client, &vpc.DescribeVpcsInput{
    IPv6Enabled: new(false),
    Tags: []vpc.TagFilter{{Key: "environment", Value: new("")}},
})
if err != nil { panic(err) }
for pages.HasMorePages() {
    page, err := pages.NextPage(context.Background())
    if err != nil { panic(err) }
    for _, item := range page.VPCs { fmt.Println(item.VPCID) }
}
```

Supply a provider such as `credentials.NewStaticProvider` with placeholder credentials
and import context, fmt, net/http and the root, credentials, sdktest and services/vpc
packages, as in the [complete README example](../README.md). Output: `vpc-example`.
Real calls need authorized credentials; remove the scripted transport. Tests require
neither an account nor network access.

Response models flatten `Tags.Tag`, `Ipv6CidrBlocks.Ipv6CidrBlock` and
`VSwitchIds.VSwitchId`; generated JSON v2 methods restore those containers on marshal
and publish a decoded model only on success. Unknown fields are ignored and future
status strings are retained. Owner IDs remain `int64`. The service-returned vSwitch ID
subset is not a complete vSwitch enumeration; use its dedicated API when needed.

Sources: [DescribeVpcs contract](https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs),
[VPC endpoint table](https://help.aliyun.com/zh/vpc/developer-reference/api-vpc-2016-04-28-endpoint).
Protocol metadata and hashes are pinned in metadata/vpc/manifest.json; regeneration is
offline. No VPC waiter, complete product coverage, live-cloud acceptance or published
pkg.go.dev version is claimed.

## 中文

`services/vpc` 提供 VPC 2016-04-28 的生成 `DescribeVpcs` 客户端、选定模型和
`DescribeVpcsPaginator`。ACS3、凭据、显式重试、middleware、结构化错误和可选 tracing
复用共享 runtime。仅覆盖一个审核操作及选定字段，参见[生成 API 指南](generated/vpc.md)
和[扩展设计](generator-expansion.md)。

使用 `vpc.New(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: provider})` 构造。
`RegionID` 在当前调用覆盖配置地域。默认公网端点仅覆盖 cn-hangzhou、cn-shanghai、cn-beijing、
cn-shenzhen、ap-southeast-1；其他地域需要明确核实的 HTTPS 端点或 resolver。
客户端可并发共享，调用者不得并发修改输入。middleware 前复制输入、切片和指针成员，
每次 paginator 请求再独立复制。

可选布尔过滤用 `*bool`：`IPv6Enabled: nil` 不过滤 IPv6，`new(false)` 选择未启用，
`new(true)` 选择已启用。`IsDefault` 为 nil 时由服务决定省略/默认行为。
`DryRun: new(true)` 请求预检；服务成功预检的 `DryRunOperation` 仍以 `alicloud.APIError`
返回并包装在 `OperationError` 中，使用 `errors.As` 提取。

Tag 编码为 `Tag.N.Key`/`Tag.N.Value`，最多二十个。Key 必须非空；Key 和提供的 Value
最多 128 个 Unicode 字符，拒绝无效 UTF-8、`aliyun`/`acs:` 前缀和 `http://`/`https://` 子串。
`Value: nil` 省略，`Value: new("")` 显式发送空值。这些仅存在于说明中的规则保留在手写 validator。
`VPCID` 接受最多二十个逗号分隔 ID；拒绝负 owner ID 和分页参数。

Paginator 仅使用页码，零值默认第一页、每页十条，拒绝大于五十的页大小。空页或到达总数时停止；
服务失败、取消、nil 输出和不一致分页元数据均保留当前游标。返回页大小为零时使用请求大小。
单消费者遍历，先检查 `HasMorePages` 再调用 `NextPage(ctx)`；耗尽返回 `pagination.ErrNoMorePages`。
遍历期间的数据变化可能改变服务总数，不提供快照或去重保证。

英文章节的离线示例由包外 Example 验证：使用 placeholder 静态凭据 provider，按
[完整 README 示例](../README.md) 导入 context、fmt、net/http、根包、credentials、sdktest、
services/vpc，输出 `vpc-example`。真实调用需有权限凭据并移除脚本 transport；测试无需账号或网络。

响应模型展平 `Tags.Tag`、`Ipv6CidrBlocks.Ipv6CidrBlock`、`VSwitchIds.VSwitchId`，生成 JSON v2
方法在 marshal 时还原容器，仅在完整解码成功时更新模型。忽略未知字段，保留未来状态字符串，
owner ID 保持 `int64`。返回的 vSwitch ID 是子集，完整枚举需专用 API。

来源：[DescribeVpcs 契约](https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs)、
[VPC 端点表](https://help.aliyun.com/zh/vpc/developer-reference/api-vpc-2016-04-28-endpoint)。
协议元数据与 hash 固定在 metadata/vpc/manifest.json，再生成无需联网。
不宣称 VPC waiter、完整产品、真实云验收或已经发布的 pkg.go.dev 版本。
