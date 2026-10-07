# Product capability policy / 产品能力策略

## English

Stage #37 follows #36 / PR #41 on `issue/37-capability-policies`, stacked on
`issue/36-batch-go-emission`. Dependencies #41 -> #40 -> #39 -> #32 remain unmerged.
This specification precedes code. Policy files under `policies/` are sparse reviewed
exceptions, never a required list of every API/model/field. Bind policy to product,
API version and source-manifest hash; each operation entry has evidence references.
Missing policy means unreviewed, no generated capabilities and no Standard retry.
Unknown/unsupported actions, paths, types, names, modes or constraints fail before
writes. Optional policy absence must not prevent complete supported Go emission.

Initial coverage: native paginators for ECS DescribeInstances (token by default,
explicit page fields choose pages), DescribeInstanceStatus and DescribeImages, plus
VPC DescribeVpcs; ECS InstanceRunningWaiter; reviewed idempotency for these four
queries and ECS DescribeRegions; ClientToken generation/validation for ECS
AllocateDedicatedHosts; conservative formatting for STS AssumeRole sensitive models.
The other emitted actions remain unreviewed. Per-operation reports distinguish
reviewed capability policy, generated adapters, conservative retry and live acceptance.

Paginator policy identifies wire input/output cursor paths, collection path, limit,
page/size/total, adapter defaults and maxima. Paths are resolved against the complete
IR and emitted as typed accesses with nil guards; no runtime field-name guessing.
Constructors follow NewOperationPaginator and return errors for invalid options;
HasMorePages/NextPage(ctx, operationOptions...) reuse the shared engine. Deep-copy
input at construction and each fetch; copy options and isolate each page override.
Errors/cancellation do not advance cursors. Duplicate tokens deliver the fetched page
then stop by default; disabling protection can permit cycles. Empty token pages with
continuation still advance; token completion ignores TotalCount. Page completion
uses a reviewed total count and collection; missing totals, negative metadata,
mismatched page numbers and overlarge response sizes fail. Empty collections stop.
Bound page arithmetic to the native int32 field. Dual-mode requests reject mixing
non-nil token/limit with page/size fields, including explicit empty/zero pointers.
Status paginator uses an intentional size 50, distinct from the service default 10.
Other initial defaults are 10; status/VPC maxima are 50, image/instance maxima 100.

Waiter policy records ID input, native collection/member ID/state, first-page fields,
maximum IDs, success/retry states and evidence. Reuse shared total-time/backoff engine
and reusable Wait/WaitForOutput. InstanceRunningWaiter requires 1..50 distinct IDs,
forces page one/size 50 and requires every requested ID Running. Missing IDs retry;
Pending/Starting/Stopping/Stopped retry; duplicate requested IDs, missing/unknown
states, nil results and API errors fail. Options allow a reviewed acceptor override,
but cannot turn failed fetches/cancellation/expiry into success. Each invocation and
poll has isolated inputs/options; an immutable waiter supports concurrent waits.

Operation policies explicitly mark safe read replay; Standard remains opt-in. Neither
CLI operation_type nor HTTP verbs grant retry safety. AllocateDedicatedHosts stays
non-idempotent for retry purposes even with ClientToken: this stage deliberately
does not promote token-bearing writes to automatic retries. Generate a cryptographically
random ASCII token once on the owned input when absent, before hooks; preserve explicit
tokens, validate nonempty ASCII length <=64 before execution and again after hooks.
Caller input remains unchanged. Separate invocations get distinct generated tokens;
callers intentionally repeating an operation supply the same explicit token.

Generate typed public ValidateOperationInput functions for sparse bounds, ASCII/string
length, array cardinality/nonempty IDs and mixed-mode exclusions. Nil is an empty
request; optional API fields are not made required. Validate copied input and
post-Initialize input before serialization. Diagnostics identify paths/rules without
including values. An explicit zero pointer violates a reviewed positive bound, while
unconstrained scalar zero/false/empty continues to serialize. Keep requiredness and
constraint claims limited to the recorded rules.

Sensitive model String/GoString return a constant type/redacted marker for both values
and pointers, covering ordinary fmt output, including %#v. This hides the entire
marked model, not a guessed subset. JSON serialization and direct field access remain
explicit raw-data operations; no logger is added. Mark STS request/body/credentials/
response envelope. Naming exceptions change only Go names: initially the native
DescribeInstanceStatus InstanceId array becomes InstanceIDs. Wire names remain exact.

Acceptance: deterministic independent policy loading and emission; negative schema/
source/type/path/naming cases leave outputs untouched; multi-page token and page mocks,
cyclic tokens, empty token pages, failed/canceled fetches, nil containers, metadata
checks, input/callback isolation; all-ID waiter states, missing/duplicates, timeout/
errors/acceptor overrides and concurrent waits; signed token/presence/validator checks,
opt-in read retry and non-retrying writes, fmt redaction; offline Examples for each
adapter/policy, corresponding guides and package comments. Run Node frontend checks/
47 tests, both Go generation checks, formatting/doccheck/vet/full Go tests once after
implementation; Linux race and Windows CI; no live cloud calls or full-policy coverage.

Evidence inspected 2026-10-07: [DescribeInstances](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-describeinstances),
[DescribeInstanceStatus](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-describeinstancestatus),
[DescribeImages](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-describeimages),
[DescribeVpcs](https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs),
[AllocateDedicatedHosts](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-allocatededicatedhosts),
and pinned DSL/earlier accepted reference policies. Web descriptions are separate
behavior evidence, not embedded mutable generation inputs. Waiter acceptance and
conservative write retry are our reviewed SDK policy, not official waiter declarations.

## 中文

阶段 #37 继 #36 / PR #41，在 `issue/37-capability-policies` 叠加
`issue/36-batch-go-emission`；依赖 #41 → #40 → #39 → #32 未合并。本规格先于代码。
`policies/` 为少量审核例外，不是全部 API/模型/字段清单；绑定产品、版本和来源 manifest
哈希，各操作策略有证据。无策略表示未审核、不生成能力、不被 Standard 重试。未知/
不支持操作、路径、类型、名称、模式、约束在写前失败；缺策略不妨碍完整支持 Go 输出。

首批为 ECS DescribeInstances（默认 token，显式页码字段选页码）、DescribeInstanceStatus、
DescribeImages 及 VPC DescribeVpcs 分页；ECS InstanceRunningWaiter；上述四查询及
DescribeRegions 的读取幂等策略；AllocateDedicatedHosts ClientToken 生成/校验；STS
AssumeRole 敏感模型格式化保护。其余操作保持未审核，报告区分已审核策略、生成适配器、
保守重试及真实验收。

分页策略声明准确线游标路径、集合路径、limit/页码/大小/总数、适配器默认值及上限，
对完整 IR 解析为带 nil 保护的强类型访问，不在运行时猜字段。构造器遵循
NewOperationPaginator 并为无效选项返回 error，HasMorePages/NextPage(ctx, 操作选项)
复用公共引擎。构造/每次抓取深复制输入，复制选项且单页覆盖不延续。错误/取消不推进；
重复 token 默认先交付当页再停止，关闭保护可循环。空 token 页仍有游标则继续，不以
TotalCount 判断 token 结束。页码按审核总数/集合判断，缺总数、负元数据、页码不符或
响应页大小超上限失败，空集合终止；算术受原生 int32 上限保护。双模式拒绝非 nil 的
token/limit 与页码/大小混用，包括显式空/零指针。状态分页刻意使用 50，区别于服务
默认 10；其余首批默认 10，状态/VPC 上限 50，镜像/实例上限 100。

Waiter 策略声明 ID 输入、原生集合及成员 ID/状态、首页字段、ID 上限、成功/重试状态
与证据；复用总时长/退避引擎及可复用 Wait/WaitForOutput。InstanceRunningWaiter 要求
1..50 不重复 ID，固定页 1/大小 50，所有请求 ID 都 Running 才成功；缺 ID 或
Pending/Starting/Stopping/Stopped 重试，重复请求 ID、缺/未知状态、nil 结果、API 错误
失败。可覆盖 acceptor，但不能把失败请求/取消/超时变成成功。每次等待/轮询独立复制
输入/选项，不可变 waiter 支持并发等待。

策略明确读取可重放，Standard 仍需显式启用；不由 CLI operation_type 或 HTTP verb
推断安全。AllocateDedicatedHosts 即使有 ClientToken 仍不被自动重试，本阶段保守
不提升带 token 写操作的重试权限。缺 token 时在独占输入上、hooks 前生成一次随机
ASCII token，保留显式 token，执行前及 hooks 后校验非空 ASCII/长度不超过 64；调用者
输入不变，各独立调用生成不同 token，刻意重复操作应显式传同一个 token。

生成公开强类型 ValidateOperationInput，支持稀疏数值边界、ASCII/字符串长度、数组
数量/非空 ID、混合模式排斥。nil 表示空请求，不把可选 API 字段变必填；复制输入与
Initialize 后序列化前都校验，诊断仅包含路径/规则，不带值。显式零指针违反审核的
正数下限；无约束零/false/空仍准确编码。必填/约束声明仅覆盖已记录规则。

敏感模型值/指针的 String/GoString 返回固定类型/已隐藏标记，覆盖普通 fmt 及 %#v，
隐藏整个审核模型，不猜子集。JSON/直接字段访问仍明确返回原数据，不增加 logger。
标记 STS 请求/响应体/credentials/响应 envelope。命名例外只改 Go 名，首例为原生
DescribeInstanceStatus 的 InstanceId 数组改 InstanceIDs，线名仍准确保留。

验收包括确定性、独立策略加载及无效 schema/来源/类型/路径/命名的写前失败、多页
token/页码 mock、循环/空 token 页、失败/取消/nil 容器/元数据校验/输入选项隔离；
waiter 全 ID/缺失/重复/状态/超时/错误/自定义 acceptor/并发；签名 token/存在语义/
validator、显式读重试/写不重试、fmt 隐藏；各适配器/策略离线 Example、对应指南/
包注释。实现后各执行一次前端检查/47 测试、两种 Go 再生成检查、格式/doccheck/vet/
完整 Go 测试，Linux race/Windows CI；不做真实调用或宣称全策略覆盖。

证据于 2026-10-07 阅读，链接与英文章节的五个官方 API 页面一致，另有固定 DSL/已
验收参考策略。网页说明作为独立行为证据，不成为可变生成输入；waiter 接受条件及
保守写重试是本 SDK 的审核策略，不是官方 waiter 声明。
