# Live profile validation / 本地 Profile 真实验证

## English

This path is written before running live acceptance against the current generated SDK.
Use the explicitly selected local Aliyun CLI profile and its region. Compare ECS
DescribeRegions, DescribeInstances (native token and legacy pages), DescribeInstanceStatus
and VPC DescribeVpcs with the same read-only parameters through Alibaba Cloud CLI.
OpenAPI Explorer supplies CLI examples on each API debugging page; CLI comparison is
not a claim that the Explorer browser UI was exercised.

Run official CLI references first so OAuth may refresh its cached temporary credentials,
then read the selected profile locally into an explicit static SDK provider. Validate
the credential expiration; never print/copy credentials into source, reports or CLI
arguments. This is an opt-in acceptance harness, not implicit SDK profile discovery or
a general OAuth provider. Profiles requiring other renewal flows are outside this path.

Compare selected output projections and pagination metadata after excluding request IDs
and timestamps. Bound traversal to at most two pages per policy. Live resources can
change between calls: distinguish a comparison mismatch from an SDK protocol error.
If a Running ECS identity exists, verify Wait and WaitForOutput using that identity;
otherwise record the waiter check as skipped. STS AssumeRole requires an explicit target
role and is not exercised without one. Retry failures, mock helpers, credential renewal
and tracing continue to rely on the existing offline tests; live reads alone cannot
establish their acceptance. No cloud resources are created or modified.

Keep raw CLI responses and the disposable Go harness under ignored .git/ storage.
Publish only sanitized pass/fail/skip evidence, SDK commit and CLI version, not resource
IDs, names, account IDs or response bodies. Use an issue established before the harness.

Browser verification: sign in to the same account, select the same region and open:

- [DescribeRegions](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeRegions)
- [DescribeInstances](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstances)
- [DescribeInstanceStatus](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstanceStatus)
- [DescribeVpcs](https://api.aliyun.com/api/Vpc/2016-04-28/DescribeVpcs)

Use CLI Example to inspect the reference commands. Debug with RegionId set to the CLI
region: token mode uses MaxResults=10, page mode uses PageNumber=1/PageSize=10 (no token
parameters). Check HTTP success, response containers, token/page metadata and selected
fields. Browser checks must be recorded separately from SDK/CLI execution.

Source: [official CLI/Explorer guide](https://help.aliyun.com/en/openapi/developer-reference/new-cli-guide).

### Recorded acceptance

Issue [#30](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/30), 2026-10-07,
SDK commit `618303bf2091afc37c1713025ffa6e9248dd0b8c`, Aliyun CLI 3.4.11, cn-hangzhou.
Initial OAuth/STS cache was expired; the first CLI reference failed authentication.
After the user renewed the profile, all five CLI references and corresponding SDK reads
succeeded. Credentials, resource identities and raw responses are excluded from evidence.

| Check | Result | Evidence / limit |
| --- | --- | --- |
| ECS DescribeRegions | PASS | selected region fields match CLI; successful HTTP/request-ID/one-attempt metadata |
| ECS DescribeInstances token paginator | PASS | selected instances and continuation presence match; one terminal page |
| ECS DescribeInstances page paginator | PASS | selected instances and native page metadata match; one terminal page |
| ECS DescribeInstanceStatus paginator | PASS | selected statuses and page metadata match; one terminal page |
| VPC DescribeVpcs paginator | PASS | selected VPC projections and page metadata match; one terminal page |
| ECS Wait / WaitForOutput | SKIP | no Running identity in the bounded status sample |
| STS AssumeRole | SKIP | no explicit target role supplied |
| OpenAPI Explorer browser UI | NOT RUN | no browser tool/session available; manual links and steps above |

The live service ended each paginator on the first page, so multi-page continuation,
duplicate tokens, overflow, retries and wait behavior remain covered by offline regression
fixtures rather than this account's live run. No write/resource lifecycle calls were made.
This demonstrates the current SDK's explicit temporary credential injection; it does not
add OAuth renewal or native Profile loading to the public credential chain.

## 中文

本路径在当前生成 SDK 的真实验收之前建立。显式选择本地 Aliyun CLI Profile 及其地域，
使用相同只读参数对比 ECS DescribeRegions、DescribeInstances（原生 token 和旧页码）、
DescribeInstanceStatus 与 VPC DescribeVpcs。OpenAPI Explorer 各 API 调试页提供 CLI 示例；
CLI 对比不代表操作过 Explorer 浏览器页面。

先执行官方 CLI 参考，使 OAuth 可以刷新缓存临时凭据，再从选定 Profile 在本地读取并显式
注入 SDK static provider。检查凭据期限，凭据不输出、不进入源码/报告/CLI 参数。这是显式
验收脚本，不是 SDK 隐式 Profile 发现或通用 OAuth provider，其他刷新模式不在本路径范围。

排除 request ID/时间后比较选定输出字段与分页元数据；每种分页最多两页。两次调用期间真实
资源可能变化，区分比较差异与协议错误。有 Running ECS 实例时，用其 ID 验证 Wait 和
WaitForOutput，否则标记跳过。AssumeRole 需要明确目标 role，缺少时不调用。重试错误、
mock、凭据刷新与 tracing 仍由离线测试验收，真实只读调用不能证明这些能力。不会创建或修改云资源。

原始 CLI 响应及一次性 Go 验收脚本仅存忽略的 .git/ 中；公开记录脱敏 pass/fail/skip、SDK
提交及 CLI 版本，不包含资源 ID/名称、账号 ID 或原始响应。编写脚本前建立 issue。

网页核验：登录同一账号，选择同地域，打开英文章节四个准确 API 链接。通过 CLI Example
检查参考命令；调试 RegionId 为 CLI 地域。Token 模式 MaxResults=10，页码模式
PageNumber=1/PageSize=10 且不混入 token 参数。检查 HTTP 成功、响应容器、token/页元数据及
选定字段。网页检查单独记录，不与 SDK/CLI 执行混称。官方来源与英文章节相同。

### 验收记录

Issue [#30](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/30)，2026-10-07，
SDK 提交 `618303bf2091afc37c1713025ffa6e9248dd0b8c`，Aliyun CLI 3.4.11，cn-hangzhou。
初始 OAuth/STS 缓存过期，首次 CLI 参考认证失败；用户重新登录后，五组 CLI 与对应 SDK
真实读取均成功。凭据、资源标识及原始响应不进入公开证据。

| 检查 | 结果 | 证据 / 边界 |
| --- | --- | --- |
| ECS DescribeRegions | PASS | 选定地域字段与 CLI 相符；HTTP/request ID/单次尝试元数据有效 |
| ECS DescribeInstances token paginator | PASS | 实例投影与续取标志相符；第一页即结束 |
| ECS DescribeInstances 页码 paginator | PASS | 实例投影及原生分页元数据相符；第一页即结束 |
| ECS DescribeInstanceStatus paginator | PASS | 状态投影及分页元数据相符；第一页即结束 |
| VPC DescribeVpcs paginator | PASS | VPC 投影及分页元数据相符；第一页即结束 |
| ECS Wait / WaitForOutput | SKIP | 有界状态样本中没有 Running 实例 |
| STS AssumeRole | SKIP | 未提供明确目标 role |
| OpenAPI Explorer 网页 UI | 未运行 | 无可用浏览器工具/会话；上文提供人工链接及步骤 |

真实服务均在第一页结束，因此跨页续取、重复 token、溢出、重试及等待行为仍由离线回归
验证，不由这次账号读取证明。没有写入或资源生命周期调用。此次验证证明当前 SDK 可显式
注入临时凭据，并未给公共 credential chain 增加 OAuth 刷新或原生 Profile 加载。
