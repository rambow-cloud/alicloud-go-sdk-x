# Real-product generator expansion / 真实产品生成器扩展

## English

After #8/#21–#23 passed, the next gate is a real RPC product outside ECS/STS.
Use Vpc API 2016-04-28 DescribeVpcs: reviewed read-only/idempotent query, POST `/`,
JSON response. Pin official operation metadata before implementing the client. Generic
schema work #24 precedes VPC integration #25 (which depends on #24).
ROA/body profiles, benchmark #20 and preview release are subsequent stages.

| Shape | Previous profile | Expansion and acceptance |
| --- | --- | --- |
| Boolean | rejected | response bool; optional request *bool preserves nil vs false |
| Explicit scalar zero | omitted | *string/*int/*int64 preserve empty/zero vs absence; required and bounds checked |
| repeatList object array | rejected | named input model, Tag.N.Key/Value, recursive input copying; pointer Value preserves explicit empty string |
| Nested response | top-level paths only | named models can flatten nested paths; JSON v2 methods decode/encode reviewed containers atomically |
| Local references | rejected | offline #/components/schemas refs, RFC6901 key escapes, bounded chains; dangling/cyclic/external refs and structural siblings fail |
| Pagination | token + legacy pages | reviewed page-only profile reuses pagination engine; start/default/size/total/error/overflow checks |
| Unsupported schema | error | composition/maps/general nested query objects remain errors; unselected optional structures do not expand coverage |

Reference resolution is lazy over selected paths, without network or source mutation.
No arbitrary schema-wide expansion or guessed types. Named model dependencies must be
acyclic; unsupported or newly required selected model members fail before output.
Input models are declared explicitly with location=input and a parameter path ending
in [] for repeatList items. Models stay bound to their reviewed operation/schema path.
Pointers are used only for explicit presence; ordinary fields retain current APIs.
Generated copying owns input pointers, slices and nested model members before hooks.
Scalar bool input must be a pointer because zero and absence differ.

VPC coverage selects region/VPC/name/resource-group filters, boolean filters, owner ID,
page parameters and tags; outputs select identifiers, CIDRs, booleans, tags, IPv6 blocks,
vSwitch IDs and page metadata. Prose-only tag syntax and positive paging constraints
remain a handwritten validator, not a generator service branch. A generated page-only
DescribeVpcsPaginator uses totals and nonempty pages, with overflow-safe termination.
The existing shared runtime handles signing, credentials, retry, errors, middleware and
tracing unchanged. Review the same five public-region VPC endpoints explicitly in the
default resolver; unknown regions/partitions still require explicit reviewed rules.

Validation is offline: generator fixtures exercise all new scalar presence states,
named input copying, nested response JSON and local-ref failure modes; generated VPC
tests assert actual signed query encoding, body/response handling, caller ownership,
bounded retry, page-only traversal, cancellation, errors and endpoint rules. Preserve
existing ECS/STS protocol/acceptance tests. Public doc.go, English symbol comments,
deterministic external Examples, bilingual guides and CI regeneration are mandatory.
No live account, complete VPC coverage or pkg.go.dev indexing is claimed.

Sources: [DescribeVpcs](https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs),
[VPC endpoints](https://help.aliyun.com/zh/vpc/developer-reference/api-vpc-2016-04-28-endpoint),
[official metadata](https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/).

## 中文

#8/#21–#23 验收后，下一门槛是 ECS/STS 之外的真实 RPC 产品。
选 Vpc 2016-04-28 DescribeVpcs：审核为只读/幂等 query、POST `/`、JSON 响应。
实现客户端前固定官方元数据；先做通用 schema #24，再集成依赖 #24 的 VPC #25。
ROA/body、基准 #20 和预览发布属于后续阶段。

| 结构 | 原有支持 | 扩展与验收 |
| --- | --- | --- |
| 布尔值 | 拒绝 | 响应 bool，可选请求 *bool 区分 nil 和 false |
| 显式标量零值 | 省略 | *string/*int/*int64 区分空/零与缺失，检查必填和范围 |
| repeatList 对象数组 | 拒绝 | 具名输入模型、Tag.N.Key/Value、递归复制；指针 Value 保留显式空字符串 |
| 嵌套响应 | 仅顶层路径 | 具名模型可展平嵌套路径；JSON v2 方法原子读写审核容器 |
| 本地引用 | 拒绝 | 离线 #/components/schemas、RFC6901 转义、有界引用链；缺失/循环/外部引用及结构兄弟成员失败 |
| 分页 | token 加旧页码 | 审核纯页码规则复用 pagination；检查起点/默认值/大小/总数/错误/溢出 |
| 不支持 schema | 报错 | composition/map/通用嵌套 query 对象仍报错；未选可选结构不扩大支持范围 |

引用仅按选中路径惰性解析，不联网、不修改源，不任意展开整个 schema 或猜类型。
具名模型依赖无环；不支持或新增未覆盖必填的选中模型成员在写入前失败。
输入模型显式 location=input，repeatList 项的参数路径以 [] 结尾，并绑定审核操作/schema 路径。
指针仅表达显式存在，普通字段保留原 API；hook 前生成代码复制指针、切片和嵌套模型成员。
请求 bool 必须为指针，以区别零值和缺失。

VPC 选择地域/VPC/名称/资源组过滤、布尔过滤、owner ID、分页及 tag；响应选择标识符、CIDR、布尔、tag、
IPv6 块、vSwitch ID 和分页元数据。仅说明中的 tag 语法及正分页约束保留手写 validator，不添加服务专属生成分支。
生成纯页码 DescribeVpcsPaginator 使用总数与非空页，有防溢出终止规则。
签名、凭据、重试、错误、middleware、tracing 继续复用现有 runtime；在默认 resolver 人工审核添加同五个公网地域的
VPC 端点；未知地域/分区继续需要明确审核规则。

离线验收覆盖：生成器标量存在状态、具名输入复制、嵌套响应 JSON、本地引用失败；生成 VPC 的实际签名 query、
body/响应、调用者所有权、有界重试、纯页码、取消、错误及端点。保留 ECS/STS 既有协议/组合测试。
公共 doc.go、英文导出注释、确定输出外部 Example、双语指南、CI 再生成均必须交付。
不宣称真实账号、完整 VPC 覆盖或 pkg.go.dev 收录。官方来源与英文章节一致。
