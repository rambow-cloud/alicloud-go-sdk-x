# VPC usage

[中文](vpc.zh-CN.md)

- `services/vpc` provides the generated `DescribeVpcs` client, selected models and `DescribeVpcsPaginator` for API 2016-04-28.
- It uses the shared runtime for ACS3 signing, credentials, opt-in retries, middleware, structured errors and optional tracing.
- This is one reviewed operation, with selected fields.
- See the [generated API guide](generated/vpc.md) and [expansion design](generator-expansion.md).

- Create a client with `vpc.New(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: provider})`. `RegionID` overrides the configured region for that call.
- Default public endpoints cover cn-hangzhou, cn-shanghai, cn-beijing, cn-shenzhen and ap-southeast-1.
- Other regions require an explicit reviewed HTTPS endpoint or resolver.
- Clients may be shared concurrently; callers must not mutate inputs concurrently.
- Inputs, slices and pointer members are copied before middleware and separately for every paginator request.

- Optional boolean filters use `*bool`: `IPv6Enabled: nil` means no IPv6 filter, `new(false)` selects disabled and `new(true)` selects enabled.
- A nil `IsDefault` delegates omission/default behavior to the service. `DryRun: new(true)` requests preflight; the service's successful preflight code `DryRunOperation` is returned as `alicloud.APIError`, wrapped by `OperationError`; inspect it with `errors.As`.

- Tag filters encode as `Tag.N.Key` and `Tag.N.Value`.
- At most twenty are allowed.
- Keys must be nonempty; keys and supplied values accept at most 128 Unicode characters, reject invalid UTF-8, prefixes `aliyun`/`acs:` and substrings `http://`/`https://`. `Value: nil` omits the value; `Value: new("")` sends an explicit empty value.
- These prose-only rules remain in a handwritten validator. `VPCID` accepts up to twenty comma-separated IDs.
- Negative owner IDs and page parameters are rejected.

- The paginator uses page numbers only, starting at page one with size ten when zero.
- Sizes above fifty are rejected.
- It stops on an empty page or when the total is reached; service failures, cancellation, nil output and inconsistent page metadata preserve the current cursor.
- Returned page size zero falls back to the requested size.
- Traverse with one consumer and check `HasMorePages` before `NextPage(ctx)`; exhaustion returns `pagination.ErrNoMorePages`.
- Changing data during traversal can change the service totals; the paginator does not promise a snapshot or deduplication.

- This offline usage is exercised by external package Examples:

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

- Supply a provider such as `credentials.NewStaticProvider` with placeholder credentials and import context, fmt, net/http and the root, credentials, sdktest and services/vpc packages, as in the [complete README example](../README.md).
- Output: `vpc-example`.
- Real calls need authorized credentials; remove the scripted transport.
- Tests require neither an account nor network access.

- Response models flatten `Tags.Tag`, `Ipv6CidrBlocks.Ipv6CidrBlock` and `VSwitchIds.VSwitchId`; generated JSON v2 methods restore those containers on marshal and publish a decoded model only on success.
- Unknown fields are ignored and future status strings are retained.
- Owner IDs remain `int64`.
- The service-returned vSwitch ID subset is not a complete vSwitch enumeration; use its dedicated API when needed.

- Sources: [DescribeVpcs contract](https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs), [VPC endpoint table](https://help.aliyun.com/zh/vpc/developer-reference/api-vpc-2016-04-28-endpoint).
- Protocol metadata and hashes are pinned in metadata/vpc/manifest.json; regeneration is offline.
- No VPC waiter, complete product coverage, live-cloud acceptance or published pkg.go.dev version is claimed.
