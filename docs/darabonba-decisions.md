# DSL and metadata decisions / DSL 与元数据决策

## English

Issue [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31), 2026-10-07.
DSL revision ec489e5c3deae95496daae2b41503ac58b221adb; public metadata versions and
raw/extracted SHA-256 are recorded in metadata/{ecs,sts,vpc}/manifest.json.
Machine-readable approvals live in metadata/darabonba-decisions.json, pinned by
each product manifest. Both frontend and Go checks reject new input coverage or
requiredness differences; selected wire type/path/binding differences also fail.
Approval edits require issue evidence and an equivalent update to this document.
Approved DSL-only fields must remain optional. Requiredness approvals only allow API
required / DSL optional; reversing that direction requires a new supported policy.

| Difference | Reviewed disposition | Behavior evidence |
| --- | --- | --- |
| ECS/VPC DSL adds OwnerAccount, OwnerId, ResourceOwnerAccount, ResourceOwnerId absent from public snapshots | Keep outside the public subset. DSL presence does not grant a supported contract | Source comparison; no live owner-override calls |
| DescribeInstances DSL has a repeated Filter model; snapshots expose Filter.1.Key/Value through Filter.4.Key/Value | #34 normalizes eight aliases and removes Filter from DSL-only approvals. Preserve the legacy Go subset; indexes are not inferred server limits | Both sources checked and offline leaf/type/case/requiredness tests; Filter not tested live |
| DescribeInstances.Tag metadata contains optional lowercase key/value, absent from current DSL | Explicit metadataOnlyFields approval for optional string Tag[].key and Tag[].value; retain exact case and keep outside the public subset. Other missing fields or type/requiredness changes fail | Pinned canonical/public snapshot and semantic DSL comparison; no live deprecated-field calls |
| DescribeInstances, DescribeInstanceStatus, DescribeVpcs DSL marks RegionId optional; metadata requires it | Retain API requiredness and existing configured-region fallback | Offline missing-region checks and prior #30 authorized reads with configured region |
| AssumeRole DSL marks RoleArn/RoleSessionName optional; metadata requires both | Retain local required-field validation before HTTP; DSL optionality models unset values, not successful omission | Existing offline STS tests; live AssumeRole skipped without an explicit role |
| Metadata allows GET and POST; DSL fixes POST for these operations | Use DSL POST and preserve current signed RPC encoding | Existing signed-wire fixtures and prior #30 reads; GET not exposed |
| DSL response wraps headers/statusCode/body; metadata describes the body | Project body; common runtime provides response metadata and structured errors | Existing response/error/middleware tests and prior #30 reads |

These are deliberate SDK contract decisions, not a claim that hidden fields are
invalid or that service-side requiredness has been experimentally proven. API
requiredness comes from pinned public metadata; it is conservative client validation.
The full product source/imports pass official semantic analysis; lowering accepts
only the documented profile. Numeric Go widths, nullable/missing-value behavior,
JSON-string arrays, time conversion and API constraints remain reviewed metadata/
overlay policy. DSL prose is not automatically translated into validators.

For a material new conflict: preserve the two source versions, reproduce with the
same parameters using Explorer/its CLI example, classify service behavior separately
from CLI local validation, record sanitized result and request date, decide the
supported contract, then update approvals/tests/docs before regeneration. Never
resolve a selected incompatible wire type/path by silently choosing one source.

Open the exact Explorer pages and use the same authorized account/region:

- [DescribeRegions](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeRegions): inspect the public request fields; do not send owner overrides.
- [DescribeInstances](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstances): inspect RegionId and Filter visibility; compare MaxResults=10 with PageNumber=1/PageSize=10 separately.
- [DescribeInstanceStatus](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstanceStatus): inspect RegionId requiredness and page response fields.
- [DescribeVpcs](https://api.aliyun.com/api/Vpc/2016-04-28/DescribeVpcs): inspect RegionId and native page metadata.
- [AssumeRole](https://api.aliyun.com/api/Sts/2015-04-01/AssumeRole): inspect RoleArn/RoleSessionName requiredness; successful debug calls require an explicitly authorized target role.

Browser verification status: **NOT RUN**. UI schema hints, CLI-generated examples,
CLI local validation and actual HTTP responses are different evidence. Prior live
SDK/CLI evidence is recorded in [live-validation.md](live-validation.md); it predates
this frontend migration and does not verify owner fields, Filter or omitted required
parameters. Regenerated Go service output is checked byte-for-byte against the
previous metadata backend by TestOfficialDSLJoinsAllProductsWithoutChangingGoContracts.
Do not report previous runs as new migration or Explorer browser passes.

## 中文

Issue [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31)，2026-10-07。
DSL revision 为 ec489e5c3deae95496daae2b41503ac58b221adb，公共元数据版本及原始/提取
SHA-256 在 metadata/{ecs,sts,vpc}/manifest.json。机器审核记录在
metadata/darabonba-decisions.json，由各产品 manifest 固定；前端和 Go 检查均拒绝
新输入范围/必填差异，选定线类型、路径、绑定变化也失败。修改批准记录必须有 issue
证据并同步本文对应语言。
批准的 DSL 独有字段必须保持可选；必填差异只允许 API 必填、DSL 可选，反转方向
需要建立新的支持策略。

| 偏差 | 审核处理 | 行为证据 |
| --- | --- | --- |
| ECS/VPC DSL 多出 OwnerAccount、OwnerId、ResourceOwnerAccount、ResourceOwnerId，公共快照没有 | 不进入公开子集；DSL 出现不等于支持契约 | 源码比较；未真实调用 owner 覆盖 |
| DescribeInstances DSL 使用重复 Filter 模型，快照列 Filter.1.Key/Value 到 Filter.4.Key/Value | #34 规范化八个绑定，移除 Filter 的 DSL 独有批准；保留旧 Go 子集，不由索引推断服务上限 | 两来源及离线叶子/类型/大小写/必填测试；Filter 未真实测试 |
| DescribeInstances.Tag 元数据包含当前 DSL 没有的可选小写 key/value | metadataOnlyFields 明确批准可选字符串 Tag[].key 和 Tag[].value，保留大小写、不进入公共子集；其他缺成员或类型/必填变化失败 | 固定 canonical/公共快照和语义 DSL 对照；未真实调用弃用字段 |
| DescribeInstances、DescribeInstanceStatus、DescribeVpcs 的 RegionId 在 DSL 可选、元数据必填 | 保留 API 必填及既有配置地域回退 | 离线缺地域检查和此前 #30 配置地域真实读取 |
| AssumeRole 的 RoleArn/RoleSessionName 在 DSL 可选、元数据必填 | 保留 HTTP 前本地必填校验；DSL 表达可 unset，不表示省略后成功 | 既有 STS 离线测试；未提供明确 role，真实调用跳过 |
| 元数据允许 GET/POST，DSL 固定 POST | 使用 DSL POST，保留既有签名 RPC 编码 | 既有签名协议测试和此前 #30 读取；不公开 GET |
| DSL 响应包裹 headers/statusCode/body，元数据描述 body | 投影 body，由共同运行时提供响应元数据/结构化错误 | 既有响应/错误/middleware 测试和此前 #30 读取 |

以上是明确的 SDK 契约决策，不表示隐藏字段无效，也不表示已实验确认服务端必填性。
API 必填规则来自固定公共元数据，属于保守客户端校验。全产品及模块通过官方语义
解析，降低仅接受已声明模式。Go 数值位宽、nullable/缺值、JSON 字符串数组、时间
转换及 API 约束仍由审核后的元数据/overlay 管理，DSL 说明不自动转为 validator。

新增实质冲突时：保留两个版本，以同参数通过 Explorer/其 CLI 示例复现，区分服务端
行为和 CLI 本地校验，记录脱敏结果及日期，决定支持契约，更新批准记录、测试、双语
文档后再生成。不静默选择某个来源以掩盖选定线类型/路径冲突。

打开英文章节对应的五个准确 Explorer 链接，使用相同授权账号和地域：DescribeRegions
检查公开字段，不发 owner 覆盖；DescribeInstances 检查 RegionId 和 Filter 可见性，分别
验证 MaxResults=10 与 PageNumber=1/PageSize=10；DescribeInstanceStatus 检查 RegionId
必填和页响应字段；DescribeVpcs 检查 RegionId 及原生页元数据；AssumeRole 检查两个
必填字段，成功调试需明确授权目标 role。

浏览器核验状态为 **NOT RUN**。UI schema 提示、CLI 示例、CLI 本地校验和实际 HTTP
响应分别记录。此前真实 SDK/CLI 证据见 [live-validation.md](live-validation.md)，早于
本次前端迁移，未验证 owner、Filter 或省略必填参数。测试
TestOfficialDSLJoinsAllProductsWithoutChangingGoContracts 将再生成 Go 服务代码与
旧元数据后端逐字节比较，不把旧运行记录当作本次迁移或 Explorer 浏览器的新通过结果。
