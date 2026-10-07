# Capability role validation review / 能力角色校验评审

## English

Tracking: [#44](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/44),
based on #38 / PR #43 (`issue/38-licensed-product-docs`). Refs roadmap #33.

Review of the full-DSL PR stack found that `validateCapabilityPolicy` checks paths
and scalar types independently but does not distinguish the roles sharing a model.
For example, changing DescribeImages paginator `page` to `PageSize` leaves both
`page` and `size` pointing at the same int32 field. The generated constructor and
fetcher then overwrite page size with page number. A `total` pointing at `PageNumber`
or `PageSize` likewise reads the wrong continuation metadata. Waiter `page`/`size`
aliases overwrite the forced first page; `id`/`state` aliases compare identifiers as
states. These are invalid policies even though each path exists and has the right type.
The checked-in seven operation policies use distinct roles; this review does not
claim a defect in their emitted paginators or change Alibaba wire semantics.

Before further expansion, reject aliases within each request, response or collection
member model. For dual paginators, request page, size and token-limit roles must be
distinct; response page, size and total roles must be distinct. Waiter request page
and size must differ, and member ID and state must differ. Identical paths across
different models or separate adapters remain valid: a native token may have the same
request/response name, and a paginator and waiter may intentionally share page fields.
Do not invent new field names or infer behavioral roles from types.

This is a separately tracked review fix based on the documentation stage branch;
the five-stage route and its outstanding dependency reviews remain authoritative.
Document the scope and create the actual issue before implementing regression tests
or the fix. Verify accepted alias cases fail before rendering, including both
product-generate and product-check without modifying existing owned output or
deleting stale artifacts. Verify existing policies and legal cross-model token reuse
still render, with generated SDK artifacts unchanged. Run frontend checks/tests,
both generation checks, doccheck, vet, full Go tests and formatting once after the
fix; Linux race and Windows CI provide platform acceptance. No live cloud calls
are needed for invalid local policy rejection.

## 中文

跟踪 [#44](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/44)，基于 #38 /
PR #43（`issue/38-licensed-product-docs`），关联路线父任务 #33。

评审完整 DSL 的 PR 栈时发现，`validateCapabilityPolicy` 分别检查路径和标量类型，
却没有区分同一模型内的不同职责。例如把 DescribeImages 分页策略的 `page` 改为
`PageSize` 后，页码和页大小均指向同一 int32 字段；生成的构造器和抓取器会用页码
覆盖页大小。`total` 指向 `PageNumber` 或 `PageSize` 也会读错后续页判断的元数据。
Waiter 的 `page`/`size` 重叠会覆盖强制首页，`id`/`state` 重叠会把标识符当作状态。
路径存在且类型正确仍不构成有效策略。当前七项已提交策略的角色各不相同；本评审
不表示其生成适配器存在此问题，也不改变阿里云原生线路语义。

继续扩展前，拒绝同一请求、响应或集合成员模型内的角色重叠。双模式分页的请求
页码、页大小、token limit 各不相同，响应页码、页大小、总数也各不相同；waiter
请求页码与页大小不同，成员 ID 与状态不同。不同模型或独立适配器之间的同名路径
仍允许：原生 token 可在请求/响应使用同名字段，分页与 waiter 可共用页码字段。
不发明字段名，也不根据类型推断行为角色。

此修复基于文档阶段分支单独跟踪；既有五阶段路线及待完成的依赖评审仍有效。
先记录范围、建立实际 issue，再实现回归测试和修复。验证非法角色映射在渲染前
失败，product-generate 与 product-check 均不能修改已有受管理产物或清理过期文件。
验证现有策略及合法跨模型 token 同名仍可输出，SDK 生成产物保持不变。修复后各
运行一次前端检查/测试、新旧生成检查、doccheck、vet、完整 Go 测试与格式检查，
Linux race 和 Windows CI 提供平台验收；本地策略拒绝无需真实云调用。
