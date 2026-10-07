# Generator development path / 生成器开发路径

## English

Foundation #19 passed before this work. Generator #8 is delivered in three ordered,
independently reviewable issues: metadata/IR, emission/integration, then regeneration
and acceptance. Benchmarks #20 remain separate. Issues are created before code.

Execution dependencies: #19 -> #21 (metadata/IR) -> #22 (emission/integration) ->
#23 (regeneration/CI). Parent #8 closes only after all three pass acceptance.

The pipeline is official operation metadata -> pinned protocol snapshots -> reviewed
overlay -> validated intermediate representation (IR) -> formatted Go and paired guides.
The first supported profile is RPC over HTTPS, POST `/`, query parameters and a JSON
200 response. Initial operations are ECS DescribeRegions, DescribeInstances,
DescribeInstanceStatus and STS AssumeRole. Unsupported selected shapes fail generation;
the generator does not guess ROA, body serialization or endpoints. Expansion #25 adds
VPC DescribeVpcs with a page-only paginator. Expansion #24 supports
bounded offline local schema references; see [the schema matrix](generator-expansion.md).

The importer is an explicit network command. Generation and verification are offline.
Each manifest records product/version, the official source URL, retrieval time, raw
SHA-256 and extracted snapshot SHA-256. Snapshots retain protocol facts, not upstream
descriptions/examples. Alibaba's metadata page states no redistribution license for its
prose; do not label that prose MIT. Our extraction, overlays, templates and generated
comments are original project work under LICENSE. Keep source attribution with snapshots.

Overlays select the public API subset and supply English/Chinese field guidance,
idiomatic names, explicit idempotency, JSON-string-array conversions, time conversions,
sensitive-model redaction, validation hooks and reviewed paginator/waiter rules.
Metadata controls supported methods, parameter locations, required fields, wire types
and response paths. Overlays cannot silently invent a wire field or change its type.
Unknown JSON members in an overlay/manifest fail; unselected upstream additions are
tolerated. Removed selected fields, changed types/styles, new required inputs, unsupported
selected reference forms, invalid identifiers and missing policy fields fail before output.

Generated clients delegate execution, signing, retries, credentials, endpoints,
middleware, errors and tracing to the existing runtime. Paginator/waiter adapters
delegate their engines to the shared packages. Validation extensions stay handwritten
where constraints are prose-only (STS syntax, ECS mutually exclusive paging modes and VPC tag syntax).
Existing public names and selected fields are preserved. Default endpoint rules stay
in the foundation resolver; metadata hosts are not automatically trusted or expanded.

Generation owns an explicit file set, marked `Code generated`. It renders all products
before writing and refuses to overwrite unmarked files. Check mode never writes and
reports missing, changed or stale generated files. Deterministic output excludes local
timestamps/absolute paths; stable sorting and go/format remove map-order differences.
CI checks regeneration, existing protocol fixtures, cross-capability integration,
Examples, public docs, bilingual guides, race tests and Windows portability.

Commands from the repository root:

```sh
go run ./internal/cmd/sdkgen generate
go run ./internal/cmd/sdkgen check
```

Both are offline; `-root PATH` selects another repository root. Check never writes.
Generate repairs owned drift and removes stale files bearing sdkgen's exact marker;
unmarked handwritten files remain protected. Inputs and ownership are preflighted
before mutation. An OS I/O failure can leave a partial multi-file update; rerun generate
after resolving it. Files are replaced individually using temporary files and rename.

To add an operation: establish an issue and protocol evidence, explicitly import its
metadata ([commands](../metadata/README.md)), review the manifest/overlay, select named
fields and exact response paths, mark idempotency, supply bilingual guidance and an
offline example, then generate and run the existing gates. Validators name local
handwritten functions of `func(OperationInput) error`; they are compiled by tests, not
executed by the generator. Policy fields are checked against selected input/output
models. Schema version 1 supports reviewed `paginators` and `waiters` collections per product.
Paginator mode `tokens` selects token-only, `pages` selects page-only, and omitted mode selects
dual traversal. Legacy singular policies remain readable but cannot be mixed with their
collection counterparts. Generated names must be unique; native policy fields are required.
Dedicated paginator options and per-page service options are described in [pagination](pagination.md). Optional
scalar pointers preserve explicit false/empty/zero; location=input models bind repeatList
items and are deeply copied. Nested response projections use generated JSON v2 methods.
Local #/components/schemas chains have a 32-node bound; external/dangling/cyclic references
and structural siblings fail. Composition/maps and general nested query objects are not
supported. Endpoints continue to use the shared resolver.

Acceptance mapping: #21 -> metadata_test.go (including synthetic IR and schema drift);
#22 -> emit_test.go (isolated synthetic client compilation with GOPROXY=off), existing
ECS/STS protocol and paginator/waiter/helper tests plus foundation_test.go;
#23 -> generate_test.go, CLI tests, CI check and docs/language gates.
Expansion #24 -> expansion_test.go (isolated generated compilation, presence/copying,
nested JSON and bounded local references); #25 -> services/vpc tests (signed wire, retry
ownership, page boundaries, errors and runnable Examples) and label-classifier tests.

Next profiles (ROA/body encodings, general nested requests and wider service coverage)
require separate issues and protocol evidence; this first working generator is not
full Alibaba Cloud schema coverage or an API compatibility guarantee before v1.

Sources: [official metadata guide](https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/),
[ACS3](https://help.aliyun.com/zh/sdk/product-overview/v3-request-structure-and-signature).

## 中文

基础 #19 验收通过后开始本阶段。#8 分成三个顺序执行、可独立评审的 issue：元数据/IR、
生成/集成、再生成门禁/验收。基准 #20 独立；先创建 issue 再写代码。

执行依赖：#19 → #21（元数据/IR）→ #22（生成/集成）→ #23（再生成/CI）；全部验收后关闭父任务 #8。

流水线为官方操作元数据 → 固定协议快照 → 审核 overlay → 校验 IR → 格式化 Go 代码及双语指南。
首版支持 RPC、HTTPS、POST `/`、query 参数、JSON 200 响应；操作为 ECS 的 DescribeRegions、
DescribeInstances、DescribeInstanceStatus 和 STS AssumeRole。选中但不支持的结构直接失败，
不猜测 ROA、body 编码或端点。#25 添加 VPC DescribeVpcs 和纯页码 paginator。
#24 扩展支持有界离线本地 schema 引用，见[结构矩阵](generator-expansion.md)。

导入是显式联网命令，生成与检查离线。manifest 记录产品/版本、官方 URL、获取时间、原始及
提取快照的 SHA-256。快照只保留协议事实，不保留上游说明/示例。官方页面未为说明文字声明再分发
许可证，不将其标成 MIT；本项目的提取、overlay、模板、生成注释为原创，遵循 LICENSE，快照保留来源。

Overlay 选择公开字段子集，提供中英文说明、Go 命名、显式幂等性、JSON 字符串数组转换、时间转换、
敏感模型脱敏、校验扩展和分页/waiter 规则。方法、参数位置、必填属性、线类型、响应路径由元数据确定；
overlay 不能静默发明线字段或改变类型。manifest/overlay 未知成员报错；未选中的上游新增字段允许存在。
选中字段删除、类型/style 改变、新必填输入、不支持的选中引用、非法标识符或规则字段缺失均在写入前失败。

生成客户端复用 runtime 的执行、签名、重试、凭据、端点、middleware、错误和 tracing；
分页/waiter 适配器复用共享引擎。仅存在于说明中的校验保留手写扩展（STS 语法、ECS 分页参数互斥和 VPC tag 语法）。
保留现有公共名称及字段。默认端点规则继续由基础 resolver 管理，不自动信任或扩展元数据里的 host。

生成器管理明确的文件集合，使用 `Code generated` 标记，先渲染全部产品再写入，拒绝覆盖无标记文件。
Check 模式不写文件，报告缺失、变化和多余生成文件。输出不含本地时间或绝对路径；排序及 go/format
消除 map 顺序差异。CI 检查再生成、已有协议测试、跨能力集成、Examples、公共注释、双语指南、race 及 Windows。

仓库根目录执行 `go run ./internal/cmd/sdkgen generate` 或 `go run ./internal/cmd/sdkgen check`，
均离线；`-root PATH` 可指定另一根目录。Check 不写文件。Generate 修复生成文件差异，删除带本生成器
准确标记的多余文件；无标记手写文件受到保护。输入及所有权在写入前检查；系统 I/O 失败可能留下部分多文件
更新，修复原因后再运行 generate。每个文件通过临时文件和 rename 替换。

新增操作：先建 issue 和协议证据，显式导入元数据（[命令](../metadata/README.md)），评审 manifest/overlay，
选择具名字段与准确响应路径，明确幂等性，提供双语说明及离线示例，再生成并运行门禁。
Validator 指向 `func(OperationInput) error` 的本地手写函数；测试负责编译，生成器不执行。
规则字段对选定输入/输出模型检查。schema v1 每产品支持审核后的 `paginators` 与 `waiters` 集合。
mode=tokens 为纯 token、pages 为纯页码、省略为双模式；旧单项策略仍可读取，但不能与对应集合混用。
生成名称须唯一，策略必须引用原生字段；专属分页及每页服务选项见[分页指南](pagination.md)。
可选标量指针保留显式 false/空/零，location=input 模型绑定 repeatList 项并深复制。
嵌套响应投影使用生成 JSON v2 方法。本地 #/components/schemas 引用最多 32 个节点，外部/缺失/循环及结构兄弟
成员失败；composition/map/通用嵌套 query 对象不支持。端点继续复用共享 resolver。

验收映射：#21 对应 metadata_test.go（含合成 IR 和 schema 漂移）；#22 对应 emit_test.go
（GOPROXY=off 的隔离合成客户端编译）、已有 ECS/STS 协议和分页/waiter/helper 测试及 foundation_test.go；
#23 对应 generate_test.go、CLI 测试、CI 再生成及文档/语言门禁。
扩展 #24 对应 expansion_test.go（隔离生成编译、存在语义/复制、嵌套 JSON 和有界本地引用）；
#25 对应 services/vpc 测试（签名线编码、重试所有权、分页边界、错误及可执行 Example）和标签分类测试。

后续 ROA/body、通用嵌套请求及更多产品须独立 issue 和协议证据；首个可用 generator 不代表全量
阿里云 schema 覆盖，也不代表 v1 前公共 API 稳定承诺。官方来源与英文章节一致。
