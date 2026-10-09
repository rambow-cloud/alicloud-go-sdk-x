# 产品 IR 产物

[English](README.md)

- 当前降低及生成范围为 ECS 380/380、VPC 396/403、STS 4/4，见 [RPC 扩展 #83](../docs/dsl-rpc-expansion.zh-CN.md)。下表保留历史数量。
- v1 schema 增加可选绑定 `encoding: "json"`、布尔字段属性 `attributes.deprecated`，以及仅用于已审核 JSON 转换内动态值的 `kind: "json", dslType: "any"`。未知转换仍拒绝。

- 本目录为 [#35](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/35) 生成的构建产物，消费前阅读[产品发现](../docs/product-discovery.zh-CN.md)。
- Schema 版本 1、profile rpc-query-json-v1 记录固定官方 DSL、语义解析器版本、完整来源锁哈希、源码/API catalog 哈希和 Apache-2.0 来源。
- 原始许可见 [LICENSE.upstream](../sources/darabonba/LICENSE.upstream)，不把来源派生定义重新标为项目 MIT。
- description/example 引用带许可来源坐标；#38 包含语义解析器说明/摘要文本， 用于英文 Go 注释、对应双语指南与文档覆盖。
- example 值仍不复制，说明不转成 validator。
- 导入模块许可独立记录在来源锁，保留已说明的 NOASSERTION 限制。

- manifest.json 固定各产品 ir.json/coverage.json 哈希。
- 按下方命令生成和检查， 不手工编辑产物。
- 模型 ID 保留具名/匿名身份；操作 reachableModels 列出本地模型引用。
- 字段 required 表示 DSL 可选性，不是服务端 API 必填规则；API 约束/策略尚未评估。
- 导入 RuntimeOptions 等明确为外部引用。
- 协议/绑定证据包含不支持行为， 可执行 bindings/protocol 只存在于已转换为 IR 的操作。
- 原始数值 DSL 类型与规范化线类型并存。

| 产品           | 已发现 | 已转换为 IR | 不支持 | 可达具名 / 声明模型 | 可达匿名模型 |
| -------------- | -----: | ----------: | -----: | ------------------: | -----------: |
| ECS 2014-05-26 |    380 |         283 |     97 |         1140 / 1150 |          913 |
| STS 2015-04-01 |      4 |           1 |      3 |             11 / 11 |            8 |
| VPC 2016-04-28 |    403 |         295 |    108 |         1208 / 1212 |          520 |

- 数量对应 revision ec489e5c3deae95496daae2b41503ac58b221adb 和当前严格降低模式。
- IR 生成本身不输出 Go；独立的 #36 [批量后端](../docs/batch-go-emission.zh-CN.md) 在 `service/` 输出 ECS 283、VPC 295、STS 1 个操作，独立报告在 docs/products。
- 发现报告刻意将下游验收标为**尚未评估**；编译证据属于输出 PR，产品能力策略/真实验收继续分别记录。
- 既有五操作 Go 桥验收单独记录。
- 不支持项保留稳定原因与源码位置。
- 坐标采用语义解析器行列规则，文件路径相对 sources/darabonba；下方 `report ecs` 命令打印 ECS 原因。
