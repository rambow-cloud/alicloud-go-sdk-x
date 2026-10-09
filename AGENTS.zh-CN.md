# 项目协作约定

- #92 下一步内部 XML 编解码遵循[XML 模型编解码层](docs/xml-model-codec.zh-CN.md)。显式区分结构根和标量根，保持 DSL 具体字段；不推断根元素，不新增未经验证的公共 OSS 操作或签名回退。

[English](AGENTS.md)

## SDK 缺口补齐路线（2026-10-09）

- #92 后续范围遵循 [OSS XML/流式路线](docs/oss-xml-protocol.zh-CN.md)。先固定完整官方产品及网关、辅助模块来源，再将准确的主机、XML 根节点和签名行为投影到 IR。OSS 不回退到 ACS3；不猜测 XML 根节点，不在流所有权、重放和校验和契约完成前宣称支持流式处理。

- #92 遵循 [FC 二进制生成路线](docs/fc-binary-generation.zh-CN.md)。当前完整 IR 和锁文件使用 schema v6（`openapi-http-v1`），明确记录路径、query、请求头、正文及操作外观类型。旧 schema 及 72 个 JSON/无响应体操作的基线保留为历史；固定来源中的 73 个 FC 操作均已生成并通过离线验证，包括二进制 InvokeFunction。不得推断能力策略或宣称真实云验收。
- 使用 tools/darabonba/import-product.cjs 增加固定产品，复用全部已锁定的直接和间接导入。端点例外写入 metadata/endpoint-source-decisions.json；保留官方原始字节，未经批准的源码、坐标或原值变化在写入前失败。

- #91 将官方端点初始化投影到 IR schema v4，并生成公共目录。遵循[端点规则](docs/endpoint-rules.zh-CN.md)；私网规则必须有准确的审核依据，不回退到公网地址。

- 用户要求补齐剩余能力，按[缺口补齐路线](docs/gap-completion.zh-CN.md)和父 issue #87 执行。
- 本路线优先于旧下一步安排，保留已接受证据；未完成的实现、真实验证、人工体验与发布分别跟踪。

## 统一服务路径

- 用户确认的 #81 [服务整合路线](docs/service-consolidation.zh-CN.md) 替代旧兼容桥保留要求。仅支持 `service/<product>`，移除 `services/` 包及输出器；历史证据和官方源码通知继续保留。

## ECS、VPC 的明确范围验收

- #85 扩展共用 RPC 生成流程，增加明确的 query/表单位置、原生 GET 和 simple 字符串数组。产品 IR 及锁文件使用 schema v3。遵循 [VPC RPC 生成补齐](docs/vpc-rpc-completion.zh-CN.md)，不按操作名推断位置、编码或方法。
- #74/#75 按执行前确定的 [ECS 验收](docs/ecs-product-acceptance.zh-CN.md)和 [VPC 验收](docs/vpc-product-acceptance.zh-CN.md)矩阵完成。
- #85 扩展后的 RPC 生成清单为 ECS 380 个、VPC 403 个操作。#74/#75 消费者验收仍是固定提交的历史证据，扩展生成数量不表示所有操作均经过真实调用验收。离线消费者契约与真实调用的选定字段证据分别记录，不支持的 DSL 操作继续说明原因。
- 真实实例 token、waiter 状态迁移及非空 VPC 续页仍为排除项，保留 NOT RUN 或 SKIP，由后续项 #79 跟踪；不宣称整体 Beta 或全云验收。
- 消费者验收由实现代理执行，独立人工体验和自动化测试耗时继续区分。
- STS、ECS、VPC 证据固定到共用消费者和 CI 的受测版本；发布及同版本 pkg.go.dev 索引仍由 #61 完成。产品收尾不创建版本标签。

## 当前路线与验收依据

- 用户于 2026-10-09 确认的 [docs/sts-ecs-vpc-path.zh-CN.md](docs/sts-ecs-vpc-path.zh-CN.md) 优先于旧版首发仅含 STS、#60 必须等待人工验收的规定。
- #60 通过记录实际结果的代理消费者验收完成；人工体验移到可选后续项 #76，没有真实结果前仍记 NOT RUN。
- 然后完成 ECS #74、VPC #75，再执行 #61。维护实际依赖、milestone 和 Project，收尾 STS 时不发布。
- 代理测试与人工任务成功率、耗时分开记录；保留更广泛 Beta 的体验标准及已接受的真实调用、来源演练证据。
- 发布门禁要求最新 STS 代理验收及 ECS/VPC 产品记录，不能以一个产品的结果绕过其他产品。

## 产品目标与验收

- 构建独立的阿里云 Go SDK，提供符合 Go 使用习惯的 API、小依赖核心、可预测的请求行为，以及 pkg.go.dev 文档。本项目不是官方 SDK。
- 修改公共 API 前，阅读 [调研](docs/research.zh-CN.md)、[设计](docs/design.zh-CN.md)和[开发路线](docs/development-path.zh-CN.md)。
- 产品范围、AC/UX 标准和证据以[产品验收](docs/product-acceptance.zh-CN.md)为准。逐项记录 PASS、FAIL、SKIP、NOT RUN；必需项未通过时，验收保持开放。
- 生成数量、编译成功、PoC 或关闭实现 issue，都不能代替产品或发布验收。Smithy 保持为隔离实验，改变路线需另建 issue 和决策。
- 首版 v0.1.0 限定为 STS，对应 milestone v0.1.0、Project 3、父 issue #57；不代表 ECS/VPC/STS 整体 Beta。
- 路线为 #58 无请求参数的 GetCallerIdentity → #59 匿名 OIDC/SAML RPC → #60 四操作、消费者及来源升级验收 → #61 发布和索引。#58/#59 是 #60 的独立前提。
- #68 要求默认配置及原生 CLI Profile/OAuth，取代旧的全面禁止发现规则。保留 #51/#53/#55 的历史证据；OIDC/SAML 真实联邦调用仍为 NOT RUN，预先排除在首版必需真实范围之外。
- 维护真实的 issue 层级和依赖、milestone、唯一状态 label 及对应 Project 状态，不声称它们自动同步。发布和索引完成前，父项和里程碑不关闭。

## Issue 驱动

- 每项代码或行为修改先建 issue，包含问题、证据、范围、依赖、验收条件和文档要求。
- 使用 rambow-cloud/alicloud-go-sdk-x 的 GitHub issues。离线草稿放在 docs/issues/，打开 PR 前同步；不编造编号。
- 每项可独立评审的工作使用 issue/<number>-<slug> 分支。实现前说明验收方式。
- 完整交付的 PR 使用 Closes #N，部分交付使用 Refs #N，并写明行为、验证和文档。
- 代码、检查、示例和文档都满足验收条件才算完成；接口存在不等于路线任务完成。

## Go API 与运行时

- 服务 Options、NewFromConfig、操作函数式选项、分页器及可复用的 Wait/WaitForOutput 采用 AWS SDK for Go v2 的调用方式。
- 保留阿里云操作名、字段名和原生分页语义；纯页码 API 不增加 NextToken。说明默认值与 v0 迁移差异，不承诺上游源码兼容。
- 所有阻塞操作以 context.Context 为第一参数；errors.Is 必须能识别取消和超时。
- 使用 net/http、可注入的 HTTP 客户端和 time.Duration。最低 Go 1.27，直接导入 encoding/json/v2；不引入旧 JSON 包、第三方替代或实验开关。
- 构造后配置保持私有，不修改共享状态或调用者输入；扩展接口必须说明并发与数据归属。
- 请求和响应使用具体类型，errors.As 提取服务错误。指针只表达有意义的缺失，不要求所有普通值使用辅助函数。
- 仅在幂等性与错误策略允许时重试；限制次数和总耗时，退避期间响应取消，任意写操作默认不重试。
- 添加依赖须在 issue 中说明必要性。核心运行时只依赖标准库，除非具体需求证明无法满足。
- 签名、编码和重试实现保持内部；确有扩展需求时再公开稳定接口。云操作必须有协议证据。
- 默认不记录凭据、授权头或原始请求/响应体。不提交真实凭据，单元测试不访问真实云资源。
- 应用优先使用可续期 STS 角色凭据和 credentials.Cache。Config/服务 Options 只接受凭据提供者，不增加裸 AK/SK/token 字段。
- config.LoadDefaultConfig 可发现临时环境凭据和原生 CLI Profile/OAuth，并按[配置契约](docs/default-configuration.zh-CN.md)续期。
- 长期密钥必须显式注入 StaticProvider、EnvProvider，或明确启用 Profile provider。保留自定义来源；不强制所有运行时只能使用 STS。
- 直接服务构造拒绝 nil 和带类型的 nil，不自动查找来源，也不在构造时通过 HTTP 获取凭据。
- 已审核的匿名 STS 操作显式配置 AnonymousProvider，不读取来源凭据或签名；签名操作仍必须提供签名凭据。

## pkg.go.dev 完成标准

- 每个公共包提供 doc.go，说明用途、用法和限制。
- 每个导出声明、字段和接口方法都有以名称开头的原生 Go 注释；按需说明零值、默认值、错误、取消、数据归属和并发，不使用 @param 风格。
- 每个公共包至少一个可执行的外部 Example，输出确定且不依赖网络或云账号；操作实现时同步示例。
- 保持 LICENSE、规范 module 导入和 README 可运行示例准确。
- 有意义的修改后，文档检查、vet、Go 测试和格式检查各运行一次；Linux CI 使用 race。
- 发布与浏览器索引按[发布流程](docs/releasing.zh-CN.md)执行。本地文档检查通过不代表已发布或已索引。

## 文档语言与写法

- 用户 2026-10-09 的[写作规范](docs/documentation-style.zh-CN.md)取代旧的中英混排规则。
- 英文使用 name.md，中文使用同目录下的 name.zh-CN.md，互相提供语言切换链接，并在同一次修改中更新。
- 中文按自然的阅读习惯表达；英文使用短句和要点列表。两份文件各自包含必要命令、来源、默认值、限制和证据。
- GitHub issue、issue 草稿、表单及 PR 模板只用英文；PR 正文、诊断、标识符和 Go 注释也使用英文。
- API 名称、测试中的真实字符串形式和上游原始通知保持原样；不为翻译改动协议字段或证据。

## 生成路线

- 用户 2026-10-07 确认的 [开发路线](docs/development-path.zh-CN.md)和[产品生成路线](docs/product-generator-roadmap.zh-CN.md)优先于旧的元数据优先、逐操作快照/补充配置、字段选择和手写文档前置计划；旧验收记录保留。
- 初始五阶段及 #44 修复已集成 main，见[集成记录](docs/generator-integration.zh-CN.md)。新增产品、协议或能力需独立 issue 和明确验收范围。
- 顺序为：来源规范化 → 自动发现完整 DSL 操作和模型并转换为 IR → 批量生成 Go → 少量审核策略 → 文档自动化和协议扩展。
- 完整官方 DSL 和官方语义解析器是主要输入；canonical 元数据规范化后只作可选补充或交叉核对。
- 自动发现操作与可达模型，不要求逐 API 元数据、字段选择或手写文档。补充配置只提供审核策略和兼容例外；#31 是五操作兼容桥，不是全产品验收。
- 分别报告发现、IR 转换、生成、编译和真实调用覆盖及不支持原因。选中不支持的行为时，在写文件前失败。
- 判断来源冲突前，先规范化索引输入和 itemName 响应包装。
- 产品统一使用 service/<product>，#81 已移除旧兼容桥及输出器，迁移按[服务整合规则](docs/service-consolidation.zh-CN.md)执行。产品变更运行 sdkgen product-check，sdkgen check 调用同一检查。
- policies/ 中的可选策略绑定来源，按完整 IR 解析准确字段路径。无效策略写文件前失败，未列出的操作保持未审核。
- 不根据操作名称或 token 形状猜测重试、分页和等待能力；按[能力规则](docs/capability-policy.zh-CN.md)审核。输入使用独立副本，重试显式启用。
- 来源中的签名初始化也须审核；不支持或动态算法不能悄悄变成 ACS3，选中生成时须在写文件前失败。
- 每阶段都交付实现、行为测试、Example、英文 Go 文档、独立中英文指南和验证记录；文档自动化阶段不推迟前面阶段的文档。
- 按[产品文档规则](docs/product-documentation.zh-CN.md)复用授权的官方说明，保留来源位置和 Apache 通知，报告缺失语言/说明。不从说明推断运行时策略，不把上游账号或资源示例用于 Go 测试。
- #93 可选 canonical 字段说明遵循 docs/canonical-prose-enrichment.zh-CN.md，先规范化索引及 itemName 表示，再按准确 DSL 字段和类型匹配；保留原始字节、哈希、JSON 指针及排除原因。不得改变运行模型或策略，也不能成为逐操作前置要求。IR 更新后重新生成说明，Go 门禁前执行其离线检查和测试。
- 十一项基础能力及手写 ECS/STS 契约测试必须先验收，才可开发生成器；仅元数据调研可以提前，接口存在不等于通过。
- 先写路线文档，再建立真实 issue 依赖，最后各分支实现；依赖图不允许循环。生成器与基准测试分开跟踪。
- #92 二进制响应按[响应流契约](docs/response-streaming.zh-CN.md)执行：先在共享运行时实现有界读取、取消、流归属和编解码器发布规则，再接入 DSL 和 Go 生成。不因返回后的读取失败重试，不增加无限制请求流。
- 二进制 ROA 生成按[FC 二进制路线](docs/fc-binary-generation.zh-CN.md)执行：使用 schema v6/openapi-http-v1，审核完整程序语义，以模型和字段改名夹具验证复用。保留请求字节和流归属，不按操作名写特殊分支，不推断重试安全。
- 生产生成使用固定官方 Darabonba 源码、导入模块和官方语义投影。Go 检查前使用 Node 22 执行前端检查和测试。
- 修改固定元数据、补充配置、模板或手写校验后重新生成，不直接编辑生成文件。生成器修改需运行 sdkgen check，产品修改还需 product-check。
- DSL 与元数据不一致时，保留双语决策文件和机器可读审核记录，不自动忽略；Explorer 浏览器证据与 CLI 证据分开记录。
- sources/darabonba、sources/openapi-meta 的上游原始源码、README 和许可通知保持原始字节。项目自己编写的来源指南分语言维护；不把上游许可改标为 MIT。

## 验证与沟通

- 不重复验证或盲目重试；只有修改有意义或失败原因已经明确时才重复相关检查。
- 可在浏览器确认的结果，提供准确 URL 和检查项，不用 curl 等请求重复验证。
- DNS 以 CloudFormation 服务端状态和输出为准，不做本地 DNS 验证；必要时提供用户核验步骤。
- 用户未明确要求时，不启动子代理。
