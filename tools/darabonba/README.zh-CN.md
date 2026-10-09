# Darabonba 构建工具

[English](README.md)

- [产品路线](../../docs/product-generator-roadmap.zh-CN.md) 优先于冲突旧逐操作前置要求。
- #31 仍为五操作兼容桥；#34 增加[来源规范化](../../docs/source-normalization.zh-CN.md)， #35 已提供无需逐操作快照/补充配置的完整产品发现；离线命令与覆盖/IR 契约见 [产品发现](../../docs/product-discovery.zh-CN.md)。
- #36 [批量 Go 后端](../../docs/batch-go-emission.zh-CN.md) 无旧补充配置地消费完整 IR。
- 在根目录执行 `go run ./internal/cmd/sdkgen product-generate` 和只读 `go run ./internal/cmd/sdkgen product-check`，输出到 `service/`；旧 `services/` 兼容桥已移除；`sdkgen generate/check` 等同于完整产品生成/检查。
- #37 从可选且绑定来源的 `policies/<product>.json` 读取 [审核能力](../../docs/capability-policy.zh-CN.md)，完整输出无需逐操作策略条目。
- 产品报告记录策略哈希及逐操作已审核/未审核状态。

- #38 [文档自动化](../../docs/product-documentation.zh-CN.md) 消费官方语义解析器的说明/注释， 生成英文 Go 注释与对应双语使用/来源索引，报告缺失说明/中文翻译。
- 原始 example 值不进入 Go Example；输出附完整 Apache 条款及来源/转换通知。

- Issue [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31) 将生产前端从纯元数据改为真实官方产品 DSL。
- 要求 Node 22、Go 1.27；官方语义解析器 2.2.1 执行语法及导入模块语义检查，官方 repo-client 1.0.3 负责显式联网解析，tar 7.5.22 解包经校验的归档。
- package-lock.json 固定直接版本及传递依赖完整性；Node/Tea 依赖仅服务于构建工具，不进入 go.mod。

- 仓库根目录按下方命令安装工具、导出投影、生成 Go、检查前端、运行前端测试、 检查 Go 再生成。
- 安装需要包仓库；安装后投影、生成、检查、测试使用本地文件，普通 Go 使用者无需 Node。
- Go 生成消费哈希固定的投影、验证来源锁并对照元数据/策略， CI 在 Linux/Windows 独立重解析 DSL。
- 两层均先完成所有产品检查再写输出；系统写入失败可能留下部分更新，解决原因后再生成。

- 只有下方两条显式 import 命令访问上游；import-canonical 固定 CLI 样本和许可证。
- 产品 commit 固定在 import.cjs， Teafile 保留原始通配符，manifest.json 与 .libraries.json 固定实际模块版本及本地路径。
- 重新 import 可能解析到新版模块，必须在 issue 下审核源码、模块、许可变化，更新 metadata/darabonba-decisions.json 后再投影。
- 导入随获取写源码，网络失败可能需要修复来源锁；生成绝不联网刷新依赖。

- 支持模式为 async operationWithOptions(request, 运行时)、判空后的直接 query 绑定、 OpenApiRequest query 编码、常量 OpenApi.Params 和 callApi 交接；降低到本项目运行时， 不输出导入的 Tea 程序。
- 选定模型支持有嵌套深度限制的具名/匿名对象、数组及标量。
- 不支持行为、 语言覆盖、新模型属性、漏绑、动态参数及不支持的选定结构明确失败。
- 原生分页、状态等待器、 Go 命名、缺失、脱敏及幂等性仍由审核的补充配置/运行时管理。
- 这是仅支持五个操作的前端， 不是通用 Darabonba 编译器。

- 新增操作前阅读[迁移](../../docs/darabonba-migration.zh-CN.md)、 [决策](../../docs/darabonba-decisions.zh-CN.md)及[来源/许可](../../sources/darabonba/README.zh-CN.md)。

## 可运行命令与示例

```sh
cd tools/darabonba
npm ci --ignore-scripts --no-audit --no-fund
cd ../..
node tools/darabonba/frontend.cjs generate
go run ./internal/cmd/sdkgen generate
node tools/darabonba/frontend.cjs check
node tools/darabonba/discovery.cjs generate
node tools/darabonba/discovery.cjs report ecs
node tools/darabonba/discovery.cjs check
npm --prefix tools/darabonba test
go run ./internal/cmd/sdkgen check
```

## 可运行命令与示例

```sh
node tools/darabonba/import.cjs
node tools/darabonba/import-canonical.cjs
```
