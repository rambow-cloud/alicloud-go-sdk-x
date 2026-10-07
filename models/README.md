# Product IR artifacts / 产品 IR 产物

## English

Generated build-time artifacts for [#35](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/35).
Read [product-discovery.md](../docs/product-discovery.md) before consuming them.
Schema version 1 and profile rpc-query-json-v1 record the pinned official product DSL,
parser version, complete source-lock hash, source/API catalog hashes and Apache-2.0
provenance. The original license is [LICENSE.upstream](../sources/darabonba/LICENSE.upstream).
Do not relabel source-derived definitions under the project MIT license. Descriptions/
examples reference coordinates in the licensed source; prose is not copied or converted
to validators. Imported module license evidence remains separately recorded in the
source lock, including its documented NOASSERTION limitation.

manifest.json hashes all product ir.json and coverage.json files. Regenerate using
`node tools/darabonba/discovery.cjs generate` and check using `node tools/darabonba/discovery.cjs check`.
Do not edit generated records directly. Model IDs retain named/anonymous identity;
operation reachableModels lists local model references. Field required records DSL
optionality, not service-side API requiredness; API constraints/policies remain unassessed.
Imported RuntimeOptions and
other module types are explicit external references. Protocol/binding evidence includes
unsupported constructs, while executable bindings/protocol exist only for lowered
operations. Numeric DSL type names remain alongside normalized wire types.

| Product | Discovered | Lowered | Unsupported | Reachable named / declared models | Reachable anonymous models |
| --- | ---: | ---: | ---: | ---: | ---: |
| ECS 2014-05-26 | 380 | 283 | 97 | 1140 / 1150 | 913 |
| STS 2015-04-01 | 4 | 1 | 3 | 11 / 11 | 8 |
| VPC 2016-04-28 | 403 | 295 | 108 | 1208 / 1212 | 520 |

Counts describe revision ec489e5c3deae95496daae2b41503ac58b221adb under the current
strict lowering profile. IR generation emits no Go clients; Go emission, compilation,
policy and live acceptance of these operations are **not assessed**. Existing five-operation
Go bridge acceptance is separate. Unsupported operations retain stable reasons and
source locations. Source coordinates use parser line/column conventions; file paths
are relative to sources/darabonba. Report ECS reasons with
`node tools/darabonba/discovery.cjs report ecs`.

## 中文

本目录为 [#35](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/35) 生成的构建
产物，消费前阅读[产品发现](../docs/product-discovery.md)。Schema 版本 1、profile
rpc-query-json-v1 记录固定官方 DSL、parser 版本、完整来源锁哈希、源码/API catalog
哈希和 Apache-2.0 来源。原始许可见
[LICENSE.upstream](../sources/darabonba/LICENSE.upstream)，不把来源派生定义重新标为
项目 MIT。description/example 引用带许可来源坐标，不复制说明或转成 validator。
导入模块许可独立记录在来源锁，保留已说明的 NOASSERTION 限制。

manifest.json 固定各产品 ir.json/coverage.json 哈希。按英文章节命令生成和检查，
不手工编辑产物。模型 ID 保留具名/匿名身份；操作 reachableModels 列出本地模型
引用。字段 required 表示 DSL 可选性，不是服务端 API 必填规则；API 约束/策略尚未评估。
导入 RuntimeOptions 等明确为外部引用。协议/绑定证据包含不支持行为，
可执行 bindings/protocol 只存在于已降低操作。原始数值 DSL 类型与规范化线类型并存。

| 产品 | 已发现 | 已降低 | 不支持 | 可达具名 / 声明模型 | 可达匿名模型 |
| --- | ---: | ---: | ---: | ---: | ---: |
| ECS 2014-05-26 | 380 | 283 | 97 | 1140 / 1150 | 913 |
| STS 2015-04-01 | 4 | 1 | 3 | 11 / 11 | 8 |
| VPC 2016-04-28 | 403 | 295 | 108 | 1208 / 1212 | 520 |

数量对应 revision ec489e5c3deae95496daae2b41503ac58b221adb 和当前严格降低模式。
IR 生成不输出 Go 客户端；这些操作的 Go 输出、编译、策略、真实验收**尚未评估**。
既有五操作 Go 桥验收单独记录。不支持项保留稳定原因与源码位置。坐标采用 parser
行列规则，文件路径相对 sources/darabonba；英文章节 report ecs 命令打印 ECS 原因。
