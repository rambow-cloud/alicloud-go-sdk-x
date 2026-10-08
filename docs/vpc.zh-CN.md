# VPC 用法

[English](vpc.md)

- `services/vpc` 提供 VPC 2016-04-28 的生成 `DescribeVpcs` 客户端、选定模型和 `DescribeVpcsPaginator`。
- ACS3、凭据、显式重试、中间件、结构化错误和可选 tracing 复用公共运行时。
- 仅覆盖一个审核操作及选定字段，参见[生成 API 指南](generated/vpc.md) 和[扩展设计](generator-expansion.zh-CN.md)。

- 使用 `vpc.New(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: provider})` 构造。
- `RegionID` 在当前调用覆盖配置地域。
- 默认公网端点仅覆盖 cn-hangzhou、cn-shanghai、cn-beijing、 cn-shenzhen、ap-southeast-1；其他地域需要明确核实的 HTTPS 端点或端点解析器。
- 客户端可并发共享，调用者不得并发修改输入。
- 中间件前复制输入、切片和指针成员， 每次分页器请求再独立复制。

- 可选布尔过滤用 `*bool`：`IPv6Enabled: nil` 不过滤 IPv6，`new(false)` 选择未启用， `new(true)` 选择已启用。
- `IsDefault` 为 nil 时由服务决定省略/默认行为。
- `DryRun: new(true)` 请求预检；服务成功预检的 `DryRunOperation` 仍以 `alicloud.APIError` 返回并包装在 `OperationError` 中，使用 `errors.As` 提取。

- Tag 编码为 `Tag.N.Key`/`Tag.N.Value`，最多二十个。
- Key 必须非空；Key 和提供的 Value 最多 128 个 Unicode 字符，拒绝无效 UTF-8、`aliyun`/`acs:` 前缀和 `http://`/`https://` 子串。
- `Value: nil` 省略，`Value: new("")` 显式发送空值。
- 这些仅存在于说明中的规则保留在手写 validator。
- `VPCID` 接受最多二十个逗号分隔 ID；拒绝负 owner ID 和分页参数。

- Paginator 仅使用页码，零值默认第一页、每页十条，拒绝大于五十的页大小。
- 空页或到达总数时停止； 服务失败、取消、nil 输出和不一致分页元数据均保留当前游标。
- 返回页大小为零时使用请求大小。
- 单消费者遍历，先检查 `HasMorePages` 再调用 `NextPage(ctx)`；耗尽返回 `pagination.ErrNoMorePages`。
- 遍历期间的数据变化可能改变服务总数，不提供快照或去重保证。

- 本文件的离线示例由包外 Example 验证：使用 placeholder 静态凭据提供者，按 [完整 README 示例](../README.zh-CN.md) 导入 context、fmt、net/http、根包、credentials、sdktest、 services/vpc，输出 `vpc-example`。
- 真实调用需有权限凭据并移除脚本 HTTP 传输实现；测试无需账号或网络。

- 响应模型展平 `Tags.Tag`、`Ipv6CidrBlocks.Ipv6CidrBlock`、`VSwitchIds.VSwitchId`，生成 JSON v2 方法在 marshal 时还原容器，仅在完整解码成功时更新模型。
- 忽略未知字段，保留未来状态字符串， owner ID 保持 `int64`。
- 返回的 vSwitch ID 是子集，完整枚举需专用 API。

- 来源：[DescribeVpcs 契约](https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs)、 [VPC 端点表](https://help.aliyun.com/zh/vpc/developer-reference/api-vpc-2016-04-28-endpoint)。
- 协议元数据与 hash 固定在 metadata/vpc/manifest.json，再生成无需联网。
- 不表示 VPC 状态等待器、完整产品、真实云验收或已经发布的 pkg.go.dev 版本。

## 可运行命令与示例

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
