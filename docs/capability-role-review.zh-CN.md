# 能力角色校验评审

[English](capability-role-review.md)

- 跟踪 [#44](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/44)，基于 #38 / PR #43（`issue/38-licensed-product-docs`），关联路线父任务 #33。

- 评审及 [main 集成](generator-integration.zh-CN.md) 已于 2026-10-08 在 PR #45 完成， 下述分支/依赖计划保留实现前范围。

- 评审完整 DSL 的 PR 栈时发现，`validateCapabilityPolicy` 分别检查路径和标量类型， 却没有区分同一模型内的不同职责。
- 例如把 DescribeImages 分页策略的 `page` 改为 `PageSize` 后，页码和页大小均指向同一 int32 字段；生成的构造器和抓取器会用页码覆盖页大小。
- `total` 指向 `PageNumber` 或 `PageSize` 也会读错后续页判断的元数据。
- Waiter 的 `page`/`size` 重叠会覆盖强制首页，`id`/`state` 重叠会把标识符当作状态。
- 路径存在且类型正确仍不构成有效策略。
- 当前七项已提交策略的角色各不相同；本评审不表示其生成适配器存在此问题，也不改变阿里云原生线路语义。

- 继续扩展前，拒绝同一请求、响应或集合成员模型内的角色重叠。
- 双模式分页的请求页码、页大小、token limit 各不相同，响应页码、页大小、总数也各不相同；状态等待器请求页码与页大小不同，成员 ID 与状态不同。
- 不同模型或独立适配器之间的同名路径仍允许：原生 token 可在请求/响应使用同名字段，分页与状态等待器可共用页码字段。
- 不发明字段名，也不根据类型推断行为角色。

- 此修复基于文档阶段分支单独跟踪；既有五阶段路线及待完成的依赖评审仍有效。
- 先记录范围、建立实际 issue，再实现回归测试和修复。
- 验证非法角色映射在渲染前失败，product-generate 与 product-check 均不能修改已有受管理产物或清理过期文件。
- 验证现有策略及合法跨模型 token 同名仍可输出，SDK 生成产物保持不变。
- 修复后各运行一次前端检查/测试、新旧生成检查、doccheck、vet、完整 Go 测试与格式检查， Linux race 和 Windows CI 提供平台验收；本地策略拒绝无需真实云调用。

- 2026-10-08 本地验证：修复前八项非法角色映射均被接受；修复后完整 Go 测试（含隔离产品编译和 Example）、官方前端/发现检查、50 项 Node 测试、新旧生成检查、 16 包 doccheck、vet 和格式通过，已有生成 SDK 产物无漂移。
- 新增测试验证合法跨模型/适配器共用路径及只读/写入失败时的产物保护。
- CI 和集成证据记录在关联 PR， 本地结果不表示 Linux race 已验收或已合并 main。
