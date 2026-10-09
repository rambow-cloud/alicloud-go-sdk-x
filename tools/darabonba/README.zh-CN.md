# Darabonba 构建工具

[English](README.md)

- #92 [FC ROA 生成](../../docs/fc-roa-product.zh-CN.md) 使用 schema v5 和准确的官方 JSON/none 参数映射。在仓库根目录执行 `node tools/darabonba/import-product.cjs fc fc-20230330`，可增加同一固定版本中的产品。该命令拒绝已存在产品和未锁定导入，保留全部已锁定的间接模块。先审核来源绑定和端点决策，再运行发现；完整发现无需新增逐操作元数据，旧前端夹具仍限定在原产品中。

- #93 [可选 canonical 说明](../../docs/canonical-prose-enrichment.zh-CN.md)在准确的表示和类型规范化后补充缺失的字段描述。`npm run discover` 更新这些投影，`npm run check` 核对；来源整体缺失时不要求逐操作元数据。运行模型和策略仍由 DSL 驱动。

- [产品路线](../../docs/product-generator-roadmap.zh-CN.md) 优先于冲突旧逐操作前置要求。
- 遵循[服务整合 #81](../../docs/service-consolidation.zh-CN.md)。#31 仅保留五操作历史证据；#34 增加[来源规范化](../../docs/source-normalization.zh-CN.md)，#35 已提供无需逐操作快照/补充配置的完整产品发现；离线命令与覆盖/IR 契约见[产品发现](../../docs/product-discovery.zh-CN.md)。
- #36 [批量 Go 后端](../../docs/batch-go-emission.zh-CN.md) 无旧补充配置地消费完整 IR。
- 在根目录执行 `go run ./internal/cmd/sdkgen product-generate` 和只读 `go run ./internal/cmd/sdkgen product-check`，输出到 `service/`；旧 `services/` 兼容桥已移除；`sdkgen generate/check` 等同于完整产品生成/检查。
- #37 从可选且绑定来源的 `policies/<product>.json` 读取 [审核能力](../../docs/capability-policy.zh-CN.md)，完整输出无需逐操作策略条目。
- 产品报告记录策略哈希及逐操作已审核/未审核状态。

- #38 [文档自动化](../../docs/product-documentation.zh-CN.md) 消费官方语义解析器的说明/注释， 生成英文 Go 注释与对应双语使用/来源索引，报告缺失说明/中文翻译。
- 原始 example 值不进入 Go Example；输出附完整 Apache 条款及来源/转换通知。

- Issue [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31) 将生产前端从纯元数据改为真实官方产品 DSL。
- 要求 Node 22、Go 1.27；官方语义解析器 2.2.1 执行语法及导入模块语义检查，官方 repo-client 1.0.3 负责显式联网解析，tar 7.5.22 解包经校验的归档。
- package-lock.json 固定直接版本及传递依赖完整性；Node/Tea 依赖仅服务于构建工具，不进入 go.mod。

- 在仓库根目录按下方命令安装工具、生成完整 IR、输出 Go，再执行前端和 Go 检查。
- 安装需要包仓库；安装后投影、生成、检查、测试使用本地文件，普通 Go 使用者无需 Node。
- Go 生成消费固定的完整 IR 和可选来源绑定策略，不依赖逐操作元数据或模型选择补充配置；CI 在 Linux/Windows 独立重解析 DSL。
- Go 生成一致性检查运行一次即可；generate/check 与 product-generate/product-check 使用同一后端。
- 两层均先完成所有产品检查再写输出；系统写入失败可能留下部分更新，解决原因后再生成。

- 只有下方两条显式 import 命令访问上游；import-canonical 固定 CLI 样本和许可证。
- 产品 commit 固定在 import.cjs， Teafile 保留原始通配符，manifest.json 与 .libraries.json 固定实际模块版本及本地路径。
- 重新 import 可能解析到新版模块，必须在 issue 下审核源码、模块、许可变化，更新 metadata/darabonba-decisions.json 后再投影。
- 导入随获取写源码，网络失败可能需要修复来源锁；生成绝不联网刷新依赖。

- 完整发现模式及不支持原因见[产品发现](../../docs/product-discovery.zh-CN.md)。下文介绍保留的 frontend.cjs 交叉验证样本，它已不再输出 Go 客户端。
- 样本支持模式为 async operationWithOptions(request, 运行时)、判空后的直接 query 绑定、OpenApiRequest query 编码、常量 OpenApi.Params 和 callApi 交接；降低到本项目运行时，不输出导入的 Tea 程序。
- 选定模型支持有嵌套深度限制的具名/匿名对象、数组及标量。
- 不支持行为、 语言覆盖、新模型属性、漏绑、动态参数及不支持的选定结构明确失败。
- 完整产品的原生分页、waiter、命名、脱敏和幂等性使用审核过的 policies/ 策略，不使用模型选择补充配置。
- frontend.cjs 保留五操作的规范化与冲突证据。npm run check 分别检查它与完整发现；审核后的样本变更可能需要 frontend.cjs generate。完整产品发现和 Go 输出不消费这些逐操作投影。
- 两种前端都不是通用 Darabonba 编译器，不支持的行为有明确原因。

- 新增操作前阅读[迁移](../../docs/darabonba-migration.zh-CN.md)、 [决策](../../docs/darabonba-decisions.zh-CN.md)及[来源/许可](../../sources/darabonba/README.zh-CN.md)。

## 可运行命令与示例

```sh
cd tools/darabonba
npm ci --ignore-scripts --no-audit --no-fund
cd ../..
node tools/darabonba/discovery.cjs generate
node tools/darabonba/discovery.cjs report ecs
go run ./internal/cmd/sdkgen product-generate
npm --prefix tools/darabonba run check
npm --prefix tools/darabonba test
go run ./internal/cmd/sdkgen product-check
```

## 可运行命令与示例

```sh
node tools/darabonba/import.cjs
node tools/darabonba/import-canonical.cjs
```
