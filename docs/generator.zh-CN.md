# 生成器开发路径

[English](generator.md)

- 新方向以 [product-generator-roadmap.md](product-generator-roadmap.zh-CN.md) 为准，优先于下文冲突要求：完整 DSL 为主、元数据可选、自动发现操作/模型、补充配置仅补策略/兼容。
- 下文描述五操作兼容桥与历史验收，不作为未来产品生成器前置要求。

- 基础 #19 验收通过后开始本阶段。
- #8 分成三个顺序执行、可独立评审的 issue：元数据/IR、 生成/集成、再生成门禁/验收。
- 基准 #20 独立；先创建 issue 再写代码。

- 执行依赖：#19 → #21（元数据/IR）→ #22（生成/集成）→ #23（再生成/CI）；全部验收后关闭父任务 #8。

- #31 后生产流水线为官方产品 DSL → 官方 Darabonba 语义语义解析器 → 确定性协议/绑定/模型投影

* 固定公共元数据 + 审核补充配置 → 校验 IR → 格式化 Go 代码及双语指南。

- 见[迁移/工具](darabonba-migration.zh-CN.md)和[来源决策](darabonba-decisions.zh-CN.md)。
- 首版支持 RPC、HTTPS、POST `/`、query 参数、JSON 200 响应；操作为 ECS 的 DescribeRegions、 DescribeInstances、DescribeInstanceStatus 和 STS AssumeRole。
- 选中但不支持的结构直接失败， 不猜测 ROA、body 编码或端点。
- #25 添加 VPC DescribeVpcs 和纯页码分页器。
- #24 扩展支持有深度限制的离线本地 schema 引用，见[结构矩阵](generator-expansion.zh-CN.md)。

- 导入是显式联网命令，生成与检查离线。
- manifest 记录产品/版本、官方 URL、获取时间、原始及提取快照的 SHA-256。
- 快照只保留协议事实，不保留上游说明/示例。
- 官方页面未为说明文字声明再分发许可证，不将其标成 MIT；本项目的提取、补充配置、模板、生成注释为原创，遵循 LICENSE，快照保留来源。

- Overlay 选择公开字段子集，提供中英文说明、Go 命名、显式幂等性、JSON 字符串数组转换、时间转换、 敏感模型脱敏、校验扩展和分页/状态等待器规则。
- DSL 提供操作常量、判空绑定及模型结构，元数据交叉校验方法、位置、API 必填、类型和响应路径；明确批准的差异保留公共契约，新差异直接失败； 补充配置不能静默发明请求或响应字段或改变类型。
- manifest/补充配置未知成员报错；未选中的上游新增字段允许存在。
- 选中字段删除、类型/style 改变、新必填输入、不支持的选中引用、非法标识符或规则字段缺失均在写入前失败。

- 生成客户端复用运行时的执行、签名、重试、凭据、端点、中间件、错误和 tracing； 分页/状态等待器适配器复用共享引擎。
- 仅存在于说明中的校验保留手写扩展（STS 语法、ECS 分页参数互斥和 VPC tag 语法）。
- 保留现有公共名称及字段。
- 默认端点规则继续由基础端点解析器管理，不自动信任或扩展元数据里的 host。

- 生成器管理明确的文件集合，使用 `Code generated` 标记，先渲染全部产品再写入，拒绝覆盖无标记文件。
- Check 模式不写文件，报告缺失、变化和多余生成文件。
- 输出不含本地时间或绝对路径；排序及 go/format 消除 map 顺序差异。
- CI 检查再生成、已有协议测试、跨能力集成、Examples、公共注释、双语指南、race 及 Windows。

- 仓库根目录执行 `go run ./internal/cmd/sdkgen generate` 或 `go run ./internal/cmd/sdkgen check`， 均离线；`-root PATH` 可指定另一根目录。
- Check 不写文件。
- Generate 修复生成文件差异，删除带本生成器准确标记的多余文件；无标记手写文件受到保护。
- 输入及所有权在写入前检查；系统 I/O 失败可能留下部分多文件更新，修复原因后再运行 generate。
- 每个文件通过临时文件和 rename 替换。

- 源码/投影变化时先执行 [tools/darabonba](../tools/darabonba/README.zh-CN.md) 的 Node 安装与前端命令。
- 生产 Go 生成必须有官方固定投影；纯元数据 Load/Render 仅保留用于合成后端测试。
- CI 在 Go 再生成之外独立检查完整官方语义投影。

- 新增操作：先建 issue 和协议证据，显式导入元数据（[命令](../metadata/README.zh-CN.md)），评审 manifest/补充配置， 选择具名字段与准确响应路径，明确幂等性，提供双语说明及离线示例，再生成并运行门禁。
- Validator 指向 `func(OperationInput) error` 的本地手写函数；测试负责编译，生成器不执行。
- 规则字段对选定输入/输出模型检查。
- schema v1 每产品支持审核后的 `paginators` 与 `waiters` 集合。
- mode=tokens 为纯 token、pages 为纯页码、省略为双模式；旧单项策略仍可读取，但不能与对应集合混用。
- 生成名称须唯一，策略必须引用原生字段；专属分页及每页服务选项见[分页指南](pagination.zh-CN.md)。
- 可选标量指针保留显式 false/空/零，location=input 模型绑定 repeatList 项并深复制。
- 嵌套响应投影使用生成 JSON v2 方法。
- 本地 #/components/schemas 引用最多 32 个节点，外部/缺失/循环及结构兄弟成员失败；composition/map/通用嵌套 query 对象不支持。
- 端点继续复用共享端点解析器。

- 验收映射：#21 对应 metadata_test.go（含合成 IR 和 schema 漂移）；#22 对应 emit_test.go （GOPROXY=off 的隔离合成客户端编译）、已有 ECS/STS 协议和分页/状态等待器/辅助组件测试及 foundation_test.go； #23 对应 generate_test.go、CLI 测试、CI 再生成及文档/语言门禁。
- 扩展 #24 对应 expansion_test.go（隔离生成编译、存在语义/复制、嵌套 JSON 和有深度限制的本地引用）； #25 对应 services/vpc 测试（签名请求编码、重试所有权、分页边界、错误及可执行 Example）和标签分类测试。

- 后续 ROA/body、通用嵌套请求及更多产品须独立 issue 和协议证据；首个可用生成器不代表全量阿里云 schema 覆盖，也不代表 v1 前公共 API 稳定承诺。
- 官方来源见参考资料。

## 可运行命令与示例

```sh
go run ./internal/cmd/sdkgen generate
go run ./internal/cmd/sdkgen check
```

## 参考资料

- [official metadata guide](https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/)
- [ACS3](https://help.aliyun.com/zh/sdk/product-overview/v3-request-structure-and-signature)
