# Issue 35: complete product discovery / Issue 35：完整产品发现

## English

Actual issue: [#35](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/35), stage
of #33, following #34 normalization. Specification [product-discovery.md](../product-discovery.md)
was committed at 3e63e39 before implementation. Branch issue/35-product-discovery is
stacked on #39 / issue/34-source-normalization, which depends on #32.

Deliverable: automatic complete official DSL candidate discovery and cycle-safe,
identity-preserving reachable model IR. Protocol/binding evidence, exact members,
numeric source types, document coordinates and source locks are retained. Optional
upstream api-info is a cross-check, never a whitelist. No legacy snapshots/overlays/
decisions/canonical prerequisite. Strict selection fails before writes; check/report
are read-only and all commands are offline. Generated artifacts are under models/.

Verification includes complete corpus accounting/determinism, missing enrichment,
unsupported/dynamic/ambiguous behavior, unknown and recursive wire types, duplicate
members/catalogs, source tampering, strict selection and read-only checks. Run the
frontend/discovery tests/checks and all documented Go/doc gates; record actual results
in the linked PR. Linux race and Windows CI remain required. Model graph references
are not executable DSL. Lowering is distinct from Go emission/compilation/live coverage;
batch Go output remains #36. Parent #33 and unmerged dependencies stay open.

## 中文

真实 issue 为 [#35](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/35)，父任务
#33，继 #34 规范化。[产品发现规格](../product-discovery.md) 在实现前提交 3e63e39。
issue/35-product-discovery 叠加在 #39 / issue/34-source-normalization，后者依赖 #32。

交付为完整官方 DSL 候选自动发现、循环安全且保留模型身份的可达 IR。保留协议/
绑定证据、准确成员、原数值类型、说明坐标和来源锁。可选上游 api-info 为交叉核对，
不是白名单；无需旧快照/overlay/决策/canonical。严格选择在写前失败；check/report
只读，所有命令离线；产物在 models/。

验证覆盖全量来源统计/确定性、缺补充、未知/动态/歧义行为、未知/递归线类型、
重复成员/catalog、来源修改、严格选择及只读检查。执行前端/发现测试检查和已说明
Go/文档门禁，实际结果记入关联 PR；Linux race/Windows CI 必须通过。模型引用不是
可执行 DSL；降低和 Go 输出/编译/真实覆盖分开，批量 Go 输出属于 #36。父任务 #33
及未合并依赖保持打开。
