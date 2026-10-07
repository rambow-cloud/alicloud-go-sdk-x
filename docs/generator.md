# Generator development path / 生成器开发路径

## English

Foundation #19 passed before this work. Generator #8 is delivered in three ordered,
independently reviewable issues: metadata/IR, emission/integration, then regeneration
and acceptance. Benchmarks #20 remain separate. Issues are created before code.

The pipeline is official operation metadata -> pinned protocol snapshots -> reviewed
overlay -> validated intermediate representation (IR) -> formatted Go and paired guides.
The first supported profile is RPC over HTTPS, POST `/`, query parameters and a JSON
200 response. Initial operations are ECS DescribeRegions, DescribeInstances,
DescribeInstanceStatus and STS AssumeRole. Unsupported selected shapes fail generation;
the generator does not guess ROA, body serialization, schema references or endpoints.

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
selected references, invalid identifiers and missing policy fields fail before output.

Generated clients delegate execution, signing, retries, credentials, endpoints,
middleware, errors and tracing to the existing runtime. Paginator/waiter adapters
delegate their engines to the shared packages. Validation extensions stay handwritten
where constraints are prose-only (STS syntax and ECS mutually exclusive paging modes).
Existing public names and selected fields are preserved. Default endpoint rules stay
in the foundation resolver; metadata hosts are not automatically trusted or expanded.

Generation owns an explicit file set, marked `Code generated`. It renders all products
before writing and refuses to overwrite unmarked files. Check mode never writes and
reports missing, changed or stale generated files. Deterministic output excludes local
timestamps/absolute paths; stable sorting and go/format remove map-order differences.
CI checks regeneration, existing protocol fixtures, cross-capability integration,
Examples, public docs, bilingual guides, race tests and Windows portability.

Next profiles (ROA/body encodings, general nested requests and wider service coverage)
require separate issues and protocol evidence; this first working generator is not
full Alibaba Cloud schema coverage or an API compatibility guarantee before v1.

Sources: [official metadata guide](https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/),
[ACS3](https://help.aliyun.com/zh/sdk/product-overview/v3-request-structure-and-signature).

## 中文

基础 #19 验收通过后开始本阶段。#8 分成三个顺序执行、可独立评审的 issue：元数据/IR、
生成/集成、再生成门禁/验收。基准 #20 独立；先创建 issue 再写代码。

流水线为官方操作元数据 → 固定协议快照 → 审核 overlay → 校验 IR → 格式化 Go 代码及双语指南。
首版支持 RPC、HTTPS、POST `/`、query 参数、JSON 200 响应；操作为 ECS 的 DescribeRegions、
DescribeInstances、DescribeInstanceStatus 和 STS AssumeRole。选中但不支持的结构直接失败，
不猜测 ROA、body 编码、schema 引用或端点。

导入是显式联网命令，生成与检查离线。manifest 记录产品/版本、官方 URL、获取时间、原始及
提取快照的 SHA-256。快照只保留协议事实，不保留上游说明/示例。官方页面未为说明文字声明再分发
许可证，不将其标成 MIT；本项目的提取、overlay、模板、生成注释为原创，遵循 LICENSE，快照保留来源。

Overlay 选择公开字段子集，提供中英文说明、Go 命名、显式幂等性、JSON 字符串数组转换、时间转换、
敏感模型脱敏、校验扩展和分页/waiter 规则。方法、参数位置、必填属性、线类型、响应路径由元数据确定；
overlay 不能静默发明线字段或改变类型。manifest/overlay 未知成员报错；未选中的上游新增字段允许存在。
选中字段删除、类型/style 改变、新必填输入、不支持的选中引用、非法标识符或规则字段缺失均在写入前失败。

生成客户端复用 runtime 的执行、签名、重试、凭据、端点、middleware、错误和 tracing；
分页/waiter 适配器复用共享引擎。仅存在于说明中的校验保留手写扩展（STS 语法和 ECS 分页参数互斥）。
保留现有公共名称及字段。默认端点规则继续由基础 resolver 管理，不自动信任或扩展元数据里的 host。

生成器管理明确的文件集合，使用 `Code generated` 标记，先渲染全部产品再写入，拒绝覆盖无标记文件。
Check 模式不写文件，报告缺失、变化和多余生成文件。输出不含本地时间或绝对路径；排序及 go/format
消除 map 顺序差异。CI 检查再生成、已有协议测试、跨能力集成、Examples、公共注释、双语指南、race 及 Windows。

后续 ROA/body、通用嵌套请求及更多产品须独立 issue 和协议证据；首个可用 generator 不代表全量
阿里云 schema 覆盖，也不代表 v1 前公共 API 稳定承诺。官方来源与英文章节一致。
