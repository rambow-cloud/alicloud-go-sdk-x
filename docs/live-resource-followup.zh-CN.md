# 现有资源的真实云验证补充

[English](live-resource-followup.md)

- 对应 #79、父项 #87；这是 2026-10-09 的部分验收记录。
- 授权范围：通过本地 Profile 查询现有资源，不创建资源，不执行改变状态的操作。
- SDK 源码：`1a58fe9f26fb2d4ddab713bd561b5ae65c4610a9`；Go 1.27.1 windows/amd64；Aliyun CLI 3.4.11。
- SDK 使用原生 `LoadDefaultConfig`，显式选择 `oss-sftp` OAuth Profile，没有注入静态凭据快照。
- 官方来源仍由 `sources/darabonba/manifest.json` 固定。直接使用生成的 ECS/VPC 客户端，以及通过策略生成的分页器和 waiter。

## 执行前确定的范围

- 先查询配置中的杭州地域；实例、实例状态及 VPC 均为空。
- 通过生成的 ECS `DescribeRegions` 获取地域列表，再在每个地域读取一页实例和 VPC；各请求都有超时上限。
- 找到 VPC 后，在相同地域用小页验证 SDK 与 CLI 的响应，检查输入所有权、终止游标和响应元数据。
- 对已处于 Available 的 VPC，使用准确 ID 调用 `Wait` 和可复用的 `WaitForOutput`，不改变资源状态。
- 原始响应和临时脚本保留在忽略的 `.git/` 中；公开记录不含凭据、账号或资源 ID、名称及响应正文。

## 实际观察

- 发现的 33 个地域均成功返回 ECS/VPC 查询结果。
- 未找到 ECS 实例。北京、张家口各有一个 VPC，其余 31 个地域没有 VPC。
- 两个非空地域的 `DescribeVpcsPaginator` 均使用 PageSize=1，返回一个终止页。
- CLI 与 SDK 类型化响应的选定字段及其存在性一致：VpcId、RegionId、CidrBlock、Status、IsDefault、Ipv6CidrBlock、EnableIpv6。
- 原生 PageNumber、PageSize、TotalCount 一致；每次对比响应均为 HTTP 200、一次尝试，且元数据 request ID 非空并与类型化字段相符。
- 分页器和 waiter 没有修改调用者输入。两个已存在的 Available VPC 均通过 `Wait` 和 `WaitForOutput`，后者返回所请求的 ID。
- 提前取消 waiter 时，`errors.Is(context.Canceled)` 正常工作。这证明配置真实客户端后的取消行为，不代表服务端发生过状态转换。

| 后续验收项 | 结果 | 适用范围 |
|---|---|---|
| VPC 非空类型化响应及原生终止页 | PASS | 每个地域只有一个 VPC |
| 已有 VPC 的 Available waiter | PASS | 当前状态立即成功，没有状态转换 |
| VPC 真实跨页续取 | 未执行 | 没有地域包含两个 VPC；不同地域的资源不能算作一次分页遍历 |
| ECS 非空 token 续取 | 未执行 | 未找到实例 |
| ECS 多 ID waiter 状态转换 | 未执行 | 没有实例，也未授权改变状态 |
| OpenAPI Explorer 网页验证 | 未执行 | 本次只有 SDK/CLI 证据 |
| 联邦身份续期 | 未执行 | 等待另行提供配置，由 #94 跟踪 |

- #79 保持打开。空页和当前状态读取成功不等于真实跨页或状态转换验收通过。
- 原生 OAuth token 轮换及自然到期由 #94 单独验收，本次不宣称已经完成。

## 可选的网页核验

- 打开 [Explorer DescribeVpcs](https://api.aliyun.com/api/Vpc/2016-04-28/DescribeVpcs)。
- 选择同一账号，RegionId 为 `cn-beijing` 或 `cn-zhangjiakou`，PageNumber=1、PageSize=1。
- 检查 Vpcs.Vpc 容器和原生分页元数据，不在公开记录中提供标识符或凭据。
- 用户提供观察结果前，网页验收保持未执行。
