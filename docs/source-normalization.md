# Source normalization / 来源规范化

## English

Stage [#34](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/34) implements the
first stage of the [authoritative roadmap](product-generator-roadmap.md), on top of
the #31 five-operation compatibility bridge. Complete official DSL and its semantic
parser remain primary. Optional pinned CLI metadata uses a versioned adapter and
retains provenance, English/Chinese descriptions and CLI/backend attributes separately
from wire properties. No new runtime dependency or public API is introduced.

Canonical ECS DescribeImages/DescribeRegions/DescribeInstances fixtures, version.json
and Apache-2.0 LICENSE are pinned in [sources/openapi-meta](../sources/openapi-meta/README.md).
The source lock verifies revision URLs, SHA-256, paths and inventory before bridge
projection writes. The Node frontend checks present optional enrichment against the
selected DSL inputs/protocol; Go independently checks all selected snapshots against
the projection, including indexed leaves and explicit compatibility exceptions.
Go generation consumes the pinned DSL projection without Node or canonical fixtures.

| Source representation | Normalized meaning |
| --- | --- |
| name: region_id; raw_name: RegionId; options: --biz-region-id | Exact wire RegionId; CLI names/options retained as attributes |
| bool/int; location; required; param_style: repeatList; element.fields | Boolean/integer, query binding, source requiredness and repeated structural models |
| Key/Value and deprecated key/value | Distinct case-sensitive members, never merged or renamed |
| Filter.1.Key/Value through Filter.4.Key/Value | Eight literal aliases of the DSL Filter array's Key/Value leaves; no inferred maximum length |
| Images array with itemName: Image | Images object with Image array; nested wrappers restored recursively |
| backendName/nullToEmpty/valueMapping | Preserved source annotations, never automatic SDK transforms |
| Top-level GET\|POST versus operation.method POST | Allowed CLI methods retained separately; actual recipe POST is cross-checked |
| operation_type: read | Source annotation; does not establish retry safety |

Index binding accepts positive decimal positions without leading zeros, bounded
structural nesting, and exact member case. Direct and flattened bindings competing
for the same root, missing/forged projection aliases, type/case drift and indexed
requiredness disagreement fail before outputs. A required indexed root without API
requiredness evidence also fails. Explicit API-required/DSL-optional approvals remain
unchanged; DSL optionality is not proof that the service accepts omission.

Source review found a real compatibility exception: DescribeInstances.Tag metadata
contains optional Tag[].key/value absent from the current DSL. The paired decision
record and machine policy explicitly list those two optional string metadata-only paths. They remain
outside the current public Go subset. New missing fields, changed case, or newly required
members are rejected; no case-folding or inferred wire alias is permitted. Filter is
removed from the DSL-only approval inventory because its eight bindings now match.

The legacy bridge still emits five operations with its existing selection policy.
Full operation/model discovery without per-operation snapshots is stage #35, followed
by batch emission #36. Normalization does not claim those stages are complete or
that every canonical response field has been live verified. Selected response/model
fields still pass the existing Go cross-checks; real DescribeImages nested wrappers
are compared to official parser output in adapter tests.

From the repository root after installing tools:

```sh
node tools/darabonba/frontend.cjs generate
go run ./internal/cmd/sdkgen generate
node tools/darabonba/frontend.cjs check
cd tools/darabonba
npm test
cd ../..
go run ./internal/cmd/sdkgen check
go run ./internal/cmd/doccheck
go vet ./...
go test ./...
```

Generation/check/test commands are offline. Tests cover real case/style/requiredness/
wrapper data, malformed metadata and source tampering, array positions beyond the
declared sample, and write-before-validation prevention. The existing cross-backend
test checks all generated Go file bytes, so no public Example or package contract
changes are expected. Keep Go 1.27, direct JSON v2 and English-primary comments.

Browser/live status: **NOT RUN** for this stage. For further evidence, open
[DescribeImages](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeImages) and inspect
the request Tag/Filter fields and response Images.Image wrapper; open
[DescribeInstances](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstances) to
inspect Filter and native pagination. Record UI hints separately from CLI local
validation and actual authorized HTTP responses in the paired decision document.
Prior #30 reads are historical and do not verify these normalization changes.

## 中文

阶段 [#34](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/34) 在 #31 五操作
兼容桥上实现[权威路线](product-generator-roadmap.md)第一阶段。完整官方 DSL 和语义
parser 为主；可选固定 CLI 元数据经版本化适配，保留来源、中英文说明、CLI/backend
属性，和线属性分开。不增加运行时依赖或公共 API。

Canonical ECS DescribeImages/DescribeRegions/DescribeInstances、version.json 及
Apache-2.0 LICENSE 固定在 [sources/openapi-meta](../sources/openapi-meta/README.md)。
来源锁在桥接投影写入前检查 revision URL、SHA-256、路径和清单。Node 前端对存在的
可选补充核对选定 DSL 输入/协议；Go 独立对照选定快照及投影，检查索引叶子和显式
兼容例外。Go 生成读取固定 DSL 投影，不需要 Node 或 canonical 样本。

| 来源表示 | 规范化语义 |
| --- | --- |
| name: region_id、raw_name: RegionId、options: --biz-region-id | 线名准确为 RegionId，CLI 名称/选项保存为属性 |
| bool/int、location、required、param_style: repeatList、element.fields | 布尔/整数、query 绑定、来源必填性和重复结构模型 |
| Key/Value 与已弃用 key/value | 大小写不同成员，不合并或重命名 |
| Filter.1.Key/Value 到 Filter.4.Key/Value | DSL Filter 数组 Key/Value 叶子的八个字面绑定，不推断长度上限 |
| Images 数组并带 itemName: Image | Images 对象内 Image 数组，递归恢复嵌套包装 |
| backendName/nullToEmpty/valueMapping | 保留来源注解，不自动转换 SDK 数据 |
| 顶层 GET\|POST 与 operation.method POST | CLI 允许方法单独保留，实际 POST 配方交叉核对 |
| operation_type: read | 来源注解，不能证明重试安全 |

索引接受没有前导零的正十进制位置，结构嵌套有界，成员大小写准确。同一根字段同时
直接/展开绑定、投影漏绑/伪造、类型/大小写漂移、索引必填性不一致都在写输出前
失败。没有 API 必填证据的必填索引根也失败。既有 API 必填/DSL 可选批准不变；
DSL 可选不证明服务端接受省略。

来源评审发现真实兼容例外：DescribeInstances.Tag 元数据包含当前 DSL 没有的可选
Tag[].key/value。双语决策文档及机器策略明确列出这两个可选字符串 metadata-only 路径，继续
不进入当前 Go 公开子集。新缺字段、大小写变化、成员新必填均拒绝，不忽略大小写
或推断线别名。Filter 八个绑定对应后，从 DSL 独有批准清单移除。

兼容桥仍按既有选择策略输出五操作。无需逐操作快照的全量操作/模型发现属于 #35，
批量输出属于 #36。本阶段不宣称二者完成或所有 canonical 响应字段已真实验证。
选定响应/模型字段继续执行 Go 交叉核对；适配器测试比较真实 DescribeImages 嵌套
包装与官方 parser 输出。

安装工具后在根目录执行英文章节的投影、生成、检查、npm test、sdkgen check、
doccheck、vet 和 Go 测试命令；均离线。测试包含真实大小写/风格/必填/包装数据、
不合法元数据和来源修改、超出样本索引的位置、检查前不写文件。既有跨后端测试
逐字节检查所有生成 Go 文件，公共 Example/包契约不应改变。保留 Go 1.27、直接
JSON v2、英文为主注释。

本阶段浏览器/真实核验为 **NOT RUN**。进一步证据可打开英文章节准确 Explorer
链接：DescribeImages 检查请求 Tag/Filter 和响应 Images.Image 包装；DescribeInstances
检查 Filter 与原生分页。在双语决策文档中把 UI 提示、CLI 本地校验、授权真实 HTTP
响应分别记录。此前 #30 读取属于历史证据，不能证明本次规范化变更。
