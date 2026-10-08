# 产品发现与 IR

[English](product-discovery.md)

- 阶段 [#35](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/35) 按 [权威路线](product-generator-roadmap.zh-CN.md) 继 #34 规范化执行。
- issue/35-product-discovery 分支叠加在 issue/34-source-normalization / PR #39，后者原先依赖 #32。
- 本规格在实现前提交，完整依赖链现已[集成 main](generator-integration.zh-CN.md)。

- 构建命令使用 Node 22、官方语义解析器 2.2.1、完整固定产品/导入源码；先核验来源哈希和导入路径，再离线语义解析。
- 不读取旧 metadata manifest、决策、快照或补充配置， 发现不需要 canonical 补充。
- 不增加 Go 运行时依赖、生成服务 API 或自动重试策略。

- 操作候选来自 DSL WithOptions 函数、API 声明及构造 OpenApi.Params 的函数。
- 常量 action 保留准确大小写，歧义/动态 action 带来源证据记为不支持。
- 可选固定 api-info.json 用于覆盖交叉核验，仅 catalog 或仅 DSL 条目均可见，不把 catalog 当白名单。
- 普通转发函数不单独算 API。
- 不认识的候选行为记入报告；来源/语义解析器失败、不合法清单在写输出前整体失败。

- 版本化产品 IR 记录来源、操作声明位置、协议常量、请求/响应根、可达具名/匿名模型及准确的 API 字段名。
- 引用保留模型身份，不递归复制模型；保留原始数值类型、可选性、属性和坐标；description/example 指向带许可证的原始来源，不变成 validator。
- RuntimeOptions/导入 HTTP 传输实现模型明确为外部引用，不可达声明也有统计。

- 覆盖区分已发现、已转换为 IR、不支持。
- IR 转换仅支持既有 HTTPS POST RPC 直接 query/json 响应模式，核验全部 API 输入与响应结构。
- ROA、body/stream、辅助组件转换、未解析/递归/ 继承的协议模型及其他不支持行为有稳定原因代码、说明、源码位置。
- 已转换为 IR 不表示已输出 Go、编译、建立运行时策略或真实云验收；后续验收在此均尚未评估。

- 在仓库根目录执行本文件所列的 generate、report ecs、check 和严格选择命令。
- 产物为 models/{ecs,sts,vpc}/ir.json、coverage.json 及哈希固定的 models/manifest.json。
- 全部产品检查后写入，check 不写、不联网；系统写入失败可能部分更新。
- 可选 --operations 提供严格验收：选中未知/不支持操作在写前失败，未选中的不支持操作仍保留全量清单。
- 报告打印实际数量、原因、位置，不编造覆盖。
- 当前固定来源的发现/IR 转换/不支持数量为 ECS 380/283/97、STS 4/1/3、VPC 403/295/108，见 [产物指南](../models/README.zh-CN.md)。

- 验收包含真实全量来源的确定性再生成和完整统计、没有旧元数据/补充配置/canonical 仍可发现、协议/绑定/模型/来源异常、严格选择写前失败、双语文档和 CLI 使用示例。
- 执行前端/发现检查与测试、sdkgen check、doccheck、vet、Go 测试、格式及语言检查； CI 使用 Linux race、Windows Go 1.27。
- 公共 Go 包继续直接 JSON v2、既有离线 Example。
- 本阶段不做浏览器/真实调用；全产品 Go 输出属于 #36、能力策略 #37、文档自动化 #38。

## 可运行命令与示例

```sh
node tools/darabonba/discovery.cjs generate
node tools/darabonba/discovery.cjs report ecs
node tools/darabonba/discovery.cjs check
node tools/darabonba/discovery.cjs check --operations ecs/DescribeImages,sts/AssumeRole
```
