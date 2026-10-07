# Batch Go emission / 批量 Go 输出

## English

Stage [#36](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/36) consumes the
complete hash-pinned `models/*/ir.json`, following #35 / PR #40. This specification
precedes implementation. Official DSL and its semantic IR are sufficient; no legacy
metadata, per-operation field/model selection or documentation overlay is required.

Generate supported products into `service/<product>` (singular), following AWS-style
service imports. The earlier `services/<product>` packages remain the five-operation
compatibility/reference bridge, with their existing paginator/waiter contracts intact.
They expose selected fields and are not aliases of the complete new models. Migration
is explicit: change the import, use optional scalar pointers, preserve native response
containers, and pass DSL string fields such as JSON-encoded IDs as strings. No upstream
source compatibility is claimed. New product capability adapters belong to #37.

The backend emits every lowered operation, its complete reachable models, Input/Output,
context-first method, service Options/NewFromConfig and small operation API interface.
Input aliases represent request roots; Output represents the HTTP JSON response body
plus runtime Metadata. Response-envelope models are also retained as DSL types; they
are not the return envelope. Optional scalar/model fields use pointers; nil omits them,
non-nil scalar pointers preserve explicit zero/false/empty. Arrays/maps retain their
element types, numeric widths and exact wire case. Initialisms and anonymous names are
deterministic; collisions fail rather than dropping fields. DSL optionality is not API
requiredness. This stage introduces no guessed required-field constraints or retry safety.

Use a private standard-library RPC helper for input snapshots and recursive query
encoding. Snapshot all pointers, slices and maps before Initialize; hooks receive
owned models. Query serialization preserves dotted member names and one-based array
indexes, JSON string fields remain strings, and nil values are omitted. Context,
signing, endpoint resolution, hooks, response limits, structured errors and retries
remain the accepted shared runtime. All new operations conservatively disallow retry
until reviewed #37 policy exists. Credentials and bodies are not logged.

`sdkgen product-generate` and read-only `product-check` operate offline. Explicit
`-operations product/Name,...` asserts support without silently narrowing the generated
product set. Unsupported or unknown selections, corrupted hashes, invalid shapes,
protocols or naming collisions fail before writes. Preflight every owned output and
reject symlinks/unmarked files; legacy and product generation have separate ownership.
Publish a deterministic emission report with discovered/lowered/emitted counts and
unsupported reasons. Compilation and live validation are separate recorded evidence,
not inferred from successful rendering.

Acceptance: deterministic regeneration; complete real-product model/operation counts;
isolated temporary-module compilation; pre-write failures and stale-file reconciliation;
signed offline HTTP tests for presence, repeated/nested query fields, case, full response
containers, copying under middleware, call-option isolation, cancellation and errors;
small-interface mocks; English Go docs and deterministic external operation Examples;
corresponding Chinese guides; Go 1.27/JSON v2, frontend checks/tests, format/doccheck,
both generation checks, vet and tests, Linux race and Windows CI. No live calls are
required and no full ECS/VPC/STS coverage is claimed while unsupported operations remain.

Pagination provenance: existing shared engine was committed directly on main as
`61d2581` (issue #7); native generated adapters/options as `89e1d07` (issue #28),
without a dedicated historical PR. New product-wide paginator/waiter policies and
adapters will be reviewed in the PR for #37, after this emission PR. Policies name
input/output cursors, collection paths, size/defaults and termination rules, rather
than guessing capabilities from field or operation names.

## 中文

阶段 [#36](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/36) 在 #35 / PR #40
之后读取完整且哈希固定的 `models/*/ir.json`；本规格先于实现。官方 DSL/语义 IR
即可生成，不依赖旧元数据、逐操作字段/模型选择或手写文档 overlay。

完整生成包使用单数 `service/<product>`，遵循 AWS 风格的服务导入。原复数
`services/<product>` 保留五操作兼容/参考桥和已有 paginator/waiter 契约，其选择性
模型不与新完整模型互为别名。迁移需明确修改导入、可选标量指针、原生响应容器；
DSL 字符串字段（例如 JSON 编码的 IDs）仍传字符串。不宣称上游源码兼容；新产品
能力适配器属于 #37。

自动输出每个已降低操作及完整可达模型、Input/Output、context 优先的方法、服务
Options/NewFromConfig、小操作 API 接口。Input 为请求根；Output 为 HTTP JSON
响应体加 runtime Metadata。DSL 响应 envelope 模型也保留，但不是方法返回包装。
可选标量/模型使用指针：nil 省略，非 nil 保留零/false/空字符串。数组/map 保留元素
类型、数值宽度和准确线大小写；首字母缩写/匿名类型名确定性，命名冲突失败而非丢
字段。DSL 可选性不等于 API 必填，本阶段不猜必填校验或重试安全。

私有标准库 RPC helper 在 Initialize 前深复制所有指针、slice/map，使 hooks 得到
独占模型；递归编码保留点分成员、从 1 起数组索引、原字符串，省略 nil。context、
签名、endpoint、hooks、响应限制、结构化错误及重试复用公共 runtime；所有新操作
在 #37 审核策略前保守禁止重试，不记录凭据/请求响应体。

离线命令为 `sdkgen product-generate`、只读 `product-check`。显式
`-operations product/Name,...` 验证支持情况，不悄悄缩小生成产品集合。不支持/未知
选择、哈希损坏、无效形状/协议、命名冲突在写前失败；预检全部输出并拒绝链接/
无生成标记文件。新旧生成各自管理产物。确定性报告区分发现/降低/输出和不支持
原因；编译和真实验证是独立证据，不因渲染成功自动宣称。

验收包括确定性再生成、真实产品完整数量、临时独立 module 编译、写前失败/过期
产物清理、离线签名 HTTP 的存在语义/重复及嵌套参数/大小写/完整响应容器/hooks
复制/调用选项隔离/取消/错误、小接口 mock、英文 Go 文档及确定性外部操作 Example
和对应中文指南、Go 1.27/JSON v2、前端检查测试、格式/doccheck/两种再生成检查/
vet/测试及 Linux race/Windows CI。无需真实云调用，仍有不支持操作时不宣称 ECS/
VPC/STS 全覆盖。

分页历史：共享引擎在 main 直接提交 `61d2581`（issue #7），原生生成适配器/选项
提交 `89e1d07`（issue #28），没有独立历史 PR。新产品完整 paginator/waiter 策略/
适配器在本次输出 PR 后的 #37 PR 评审；明确输入/输出游标、结果集合路径、页大小/
默认值及终止规则，不凭字段或操作名猜能力。
