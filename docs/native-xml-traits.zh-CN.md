# 原生 XML 根元素发现

[English](native-xml-traits.md)

- 对应 #92，接续公共 XML 编解码层及固定 OSS/helper 对比。
- 操作和模型仍以完整官方 DSL 与语义 parser 为准。原生 helper 声明仅补充序列化事实，不替代 DSL schema，不选择手写模型。
- 使用标准库 Go AST 读取完整 helper 注册表和模型文件。把源码当数据读取，不执行；输入绑定准确的模块版本、提交、校验和及文件哈希。
- 自动发现全部注册操作及明确的 XML 根标签，区分结构根和标量根；保留源码坐标、XML/JSON 名称和原生类型事实，不从操作或类型名猜测根元素。
- 拒绝畸形、动态或重复的注册绑定及有歧义的根声明；不支持的模型形状记录明确原因。提取完整成功后才返回结果，错误时不写 SDK 文件。
- 与完整固定 OSS DSL 中全部 79 个 XML 响应声明对照；先规范化结构包装层，再比较字段，记录缺失或不一致的根。本阶段仅发现根元素，不代表完整嵌套模型验证或 OSS 生成验收。
- 验收包括独立的改名注册表/模型合成夹具、来源哈希不匹配时拒绝、取消、确定的 JSON 清单/Example及完整固定来源对比；执行前端检查、相关 Go/格式/文档和最终提交 CI 门禁。
- 复用独立 tools/ossxml 模块已下载并校验的来源，不把原生实现或说明复制到运行时，不增加依赖，不调用云 API。生产来源固定、IR/Gateway、签名、命名空间和校验和接入另行完成。

## 验证

- [完整对比清单](research/oss-native-xml-roots.json)：DSL 共 90 个操作，其中 79 个声明 XML 响应；原生来源共 683 个模型、82 个明确注册的根元素。36 个响应 body 的根层字段名称一致，42 个 DSL 响应结构没有 body 字段，`ListBuckets` 一项名称冲突。这些数量不代表嵌套类型或运行时行为通过验收。
- `ListBuckets` 的 DSL 字段为 `buckets`、`isTruncated`、`marker`、`maxKeys`、`nextMarker`、`owner`、`prefix`；原生字段对应为 `Buckets`、`IsTruncated`、`Marker`、`MaxKeys`、`NextMarker`、`Owner`、`Prefix`。此外，原生 `Buckets` 多一层模型包装，DSL 的 `buckets` 则是数组。本阶段不批准大小写转换或列表包装层展开；后续接入生成器前需单独审核来源决策和夹具。Explorer 验证尚未执行。
- 原生标签未注明命名空间 URI 时，明确采用 Go XML 的局部名称匹配规则；注明绝对 URI 时要求精确匹配。公共 XML 编解码器目前仅支持精确命名空间，接入原生局部名称匹配另行完成。
- 静态注册表仅接受全局空 `make(map[string]reflect.Type)`，以及 `init` 顶层使用字符串常量键和空模型 `reflect.TypeOf` 的赋值；支持 reflect 导入别名。重复、动态或条件绑定、同名遮蔽、注册表别名、整表替换、间接修改和模型文件引用注册表均拒绝。不支持的根标签或类型保留原因。
- [准确的 helper 来源固定](../metadata/native-helper-pins/oss.json)：v0.0.6，提交 `cd82cbd16bcb3f0988e125ee63679e887867f225`。提取前校验完整原始文件哈希；不将原生源码纳入本项目，也不调用云 API。
- 使用已校验的完整 OSS 私有来源副本和已下载的原生模块复现；路径必须显式指定，命令不会下载源码或生成 SDK 文件：

```powershell
node tools/darabonba/oss-xml-roots.cjs --source-root <complete-verified-source-root> --registry <native-client.go> --models <native-structs.go>
```

- 合成测试覆盖改名根、标量/结构绑定、来源坐标、不支持的标签、注册歧义、来源变化、取消、确定输出、数据所有权和短写入。无需账号的 `ExampleParseXML` 使用独立来源夹具。
- 已通过：Node 22 前端/IR/说明检查及全部 110 项前端测试；doccheck（17 个公共包、JSON v2、标准库核心）、product-check、全量 vet、Go 测试及 Example（codegen 耗时 125.035 秒）、受版本管理的 Go 文件格式、中英文配对与本地链接及 diff 检查。最终读取器重新生成的固定来源清单与记录完全一致。最终提交 CI 仍是单独的 PR 门禁。
