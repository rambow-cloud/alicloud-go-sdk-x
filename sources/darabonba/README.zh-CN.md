# 官方来源锁

[English](README.md)

- 产品来源为 [aliyun/alibabacloud-sdk](https://github.com/aliyun/alibabacloud-sdk/tree/ec489e5c3deae95496daae2b41503ac58b221adb)， commit 固定 ec489e5c3deae95496daae2b41503ac58b221adb。
- ECS 2014-05-26、STS 2015-04-01、VPC 2016-04-28 的 main.tea、Teafile、api-info.json 保留原始字节。
- 上游[Apache 通知](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/LICENSE) 保存在 LICENSE.upstream。
- 产品说明属于该授权源码集合；公共元数据快照仍单独排除说明和示例。
- #38 生成注释复用授权语义解析器说明/摘要；双语使用/契约/来源索引明确语言及说明覆盖，原始 example 值不进入可执行 Example。
- 包 LICENSE/NOTICE 保留 Apache 条款、版权及转换归属，原创运行时/工具使用 MIT。

- 锁文件记录产品路径、原始 URL、各文件 SHA-256、语义解析器版本、模块仓库 URL、归档 SHA-1/SHA-256 和实际 scope/name/version。
- 产品 .libraries.json 将通配符映射到本地固定模块，生成不访问仓库。
- 模块源码、README 和通知保留上游原始字节，本项目的双语指南说明用途；语言门禁区分这些第三方原件和本项目编写的 Markdown。

## 固定的导入模块

| 模块                      | 固定版本 |
| ------------------------- | -------- |
| alibabacloud Credential   | 0.5.17   |
| alibabacloud EndpointUtil | 0.2.1    |
| alibabacloud GatewayPOP   | 0.1.4    |
| alibabacloud GatewaySPI   | 0.0.15   |
| alibabacloud OpenApi      | 0.3.23   |
| alibabacloud OpenApiUtil  | 0.2.11   |
| alibabacloud Paginator    | 0.0.3    |
| darabonba Array           | 0.1.1    |
| darabonba EncodeUtil      | 0.0.6    |
| darabonba Map             | 0.0.5    |
| darabonba SignatureUtil   | 0.0.11   |
| darabonba String          | 0.0.13   |
| darabonba Util            | 0.2.19   |
| darabonba XML             | 0.1.14   |

- licenses/\*.NOTICE 保存原 LICENSE 或 README 许可声明，manifest 固定文件哈希和上游 Git blob；gateway/分页器/crypto 通知含完整 Apache 条款。
- 模块源码仓库声明 Apache-2.0。
- Credential 的注册归档未提供许可声明，引用的 Node 运行时声明 MIT， 仅记录为相关运行时证据，sourceSPDX=NOASSERTION，不能据此确定注册 DSL 的许可证。
- 发布评审需保留此来源区别，不把整体源码标为 MIT，不由某语言实现推断另一份源码许可。
- 构建工具依赖许可及完整性单独记录在 package-lock.json。

- 全产品及导入模块语义解析不代表全部操作都能转换为 IR或执行全部辅助组件。
- 兼容桥识别五个 SDK 函数，首次验收的完整产品后端当时输出 579 个支持 RPC 操作并使用本项目运行时； [评审/集成证据](../../docs/generator-integration.zh-CN.md) 记录范围及剩余不支持行为。
- 导入源码提供语义解析器声明及类型检查。
- 固定 Paginator 模块是 params factory 声明，不是本项目统一分页/状态等待器引擎；SDK 不引入 Go Tea 运行时。

- 详见[工具](../../tools/darabonba/README.zh-CN.md)、[迁移](../../docs/darabonba-migration.zh-CN.md) 和[决策](../../docs/darabonba-decisions.zh-CN.md)。
