# Official source lock / 官方来源锁

## English

Products come from [aliyun/alibabacloud-sdk](https://github.com/aliyun/alibabacloud-sdk/tree/ec489e5c3deae95496daae2b41503ac58b221adb)
at commit ec489e5c3deae95496daae2b41503ac58b221adb. Keep main.tea, Teafile and
api-info.json bytes intact for ECS 2014-05-26, STS 2015-04-01 and VPC 2016-04-28.
The upstream [Apache notice](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/LICENSE)
is preserved as LICENSE.upstream. Product source descriptions are part of this
licensed corpus; public metadata snapshots separately exclude descriptions/examples.
Product comments now reuse licensed parser descriptions/annotations under #38; paired
usage/contracts/source indexes record language and prose coverage. Source example
values remain excluded from executable Examples. Package LICENSE/NOTICE preserve
Apache terms, copyright and transformation attribution; original runtime/tooling use MIT.

The lock records product paths, original URLs, per-file SHA-256, parser version,
module registry URLs, archive SHA-1/SHA-256 and resolved scope/name/version. Product
.libraries.json maps wildcard specs to local versioned modules; generation performs
no registry calls. Imported module artifacts, README and notices remain byte-exact.
Our bilingual guides explain their use; language checks distinguish these preserved
third-party artifacts from project-authored Markdown.

| Module                    | Pinned version |
| ------------------------- | -------------- |
| alibabacloud Credential   | 0.5.17         |
| alibabacloud EndpointUtil | 0.2.1          |
| alibabacloud GatewayPOP   | 0.1.4          |
| alibabacloud GatewaySPI   | 0.0.15         |
| alibabacloud OpenApi      | 0.3.23         |
| alibabacloud OpenApiUtil  | 0.2.11         |
| alibabacloud Paginator    | 0.0.3          |
| darabonba Array           | 0.1.1          |
| darabonba EncodeUtil      | 0.0.6          |
| darabonba Map             | 0.0.5          |
| darabonba SignatureUtil   | 0.0.11         |
| darabonba String          | 0.0.13         |
| darabonba Util            | 0.2.19         |
| darabonba XML             | 0.1.14         |

License evidence is stored in licenses/\*.NOTICE with file SHA-256 and upstream Git
blob identities in the manifest. These contain original LICENSE or README license
declarations; full Apache terms are included in the gateway/paginator/crypto notices.
Module source repositories declare Apache-2.0. Credential's registry archive supplies
no license declaration; its referenced Node runtime declares MIT, recorded only as
related-runtime evidence with sourceSPDX=NOASSERTION. That evidence does not establish
the registry DSL's license. Preserve this provenance distinction for release review;
do not relabel the entire corpus MIT or infer one implementation's license for another.
Build-tool package licenses/integrities are separately recorded in package-lock.json.

Complete product/import semantic analysis does not mean every operation is lowered
or every helper is executed. The compatibility bridge recognizes five SDK functions;
the complete product backend now emits 579 supported RPC operations using our runtime.
[Review/integration evidence](../../docs/generator-integration.md) records scope and
remaining unsupported behavior. Module sources supply parser declarations and type checks. The pinned Paginator
module is a params factory declaration, not our unified paginator/waiter engine.
No imported Go Tea runtime implementation is added to the SDK.

See [tools](../../tools/darabonba/README.md),
[migration](../../docs/darabonba-migration.md) and
[decisions](../../docs/darabonba-decisions.md).

## 中文

产品来源为 [aliyun/alibabacloud-sdk](https://github.com/aliyun/alibabacloud-sdk/tree/ec489e5c3deae95496daae2b41503ac58b221adb)，
commit 固定 ec489e5c3deae95496daae2b41503ac58b221adb。ECS 2014-05-26、STS
2015-04-01、VPC 2016-04-28 的 main.tea、Teafile、api-info.json 保留原始字节。
上游[Apache 通知](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/LICENSE)
保存在 LICENSE.upstream。产品说明属于该授权源码集合；公共元数据快照仍单独排除
说明和示例。#38 生成注释复用授权 parser 说明/摘要；双语使用/契约/来源索引明确
语言及说明覆盖，原始 example 值不进入可执行 Example。包 LICENSE/NOTICE 保留
Apache 条款、版权及转换归属，原创 runtime/工具使用 MIT。

锁文件记录产品路径、原始 URL、各文件 SHA-256、parser 版本、模块仓库 URL、归档
SHA-1/SHA-256 和实际 scope/name/version。产品 .libraries.json 将通配符映射到本地
固定模块，生成不访问仓库。模块源码、README 和通知保留上游原始字节，本项目的
双语指南说明用途；语言门禁区分这些第三方原件和本项目编写的 Markdown。

固定模块与版本对应英文章节表格：Credential 0.5.17、EndpointUtil 0.2.1、GatewayPOP
0.1.4、GatewaySPI 0.0.15、OpenApi 0.3.23、OpenApiUtil 0.2.11、Paginator 0.0.3、
Array 0.1.1、EncodeUtil 0.0.6、Map 0.0.5、SignatureUtil 0.0.11、String 0.0.13、
Util 0.2.19、XML 0.1.14；scope 与表格对应。

licenses/\*.NOTICE 保存原 LICENSE 或 README 许可声明，manifest 固定文件哈希和
上游 Git blob；gateway/paginator/crypto 通知含完整 Apache 条款。模块源码仓库声明
Apache-2.0。Credential 的注册归档未提供许可声明，引用的 Node runtime 声明 MIT，
仅记录为相关运行时证据，sourceSPDX=NOASSERTION，不能据此确定注册 DSL 的许可证。
发布评审需保留此来源区别，不把整体源码标为 MIT，不由某语言实现推断另一份源码
许可。构建工具依赖许可及完整性单独记录在 package-lock.json。

全产品及导入模块语义解析不代表全操作降低或执行全部 helper。兼容桥识别五个 SDK
函数，完整产品后端现输出 579 个支持 RPC 操作并使用本项目运行时；
[评审/集成证据](../../docs/generator-integration.md) 记录范围及剩余不支持行为。
导入源码提供 parser 声明及类型检查。固定 Paginator 模块
是 params factory 声明，不是本项目统一分页/waiter 引擎；SDK 不引入 Go Tea 运行时。

详见[工具](../../tools/darabonba/README.md)、[迁移](../../docs/darabonba-migration.md)
和[决策](../../docs/darabonba-decisions.md)。
