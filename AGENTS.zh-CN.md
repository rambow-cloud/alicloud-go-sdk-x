# 项目协作约定

[English](AGENTS.md)

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
- 完整 DSL 产品使用 service/<product>；services/<product> 保留兼容参考实现，迁移按[批量生成规则](docs/batch-go-emission.zh-CN.md)执行。新旧生成器不得覆盖彼此文件。
- policies/ 中的可选策略绑定来源，按完整 IR 解析准确字段路径。无效策略写文件前失败，未列出的操作保持未审核。
- 不根据操作名称或 token 形状猜测重试、分页和等待能力；按[能力规则](docs/capability-policy.zh-CN.md)审核。输入使用独立副本，重试显式启用。
- 来源中的签名初始化也须审核；不支持或动态算法不能悄悄变成 ACS3，选中生成时须在写文件前失败。
- 每阶段都交付实现、行为测试、Example、英文 Go 文档、独立中英文指南和验证记录；文档自动化阶段不推迟前面阶段的文档。
- 按[产品文档规则](docs/product-documentation.zh-CN.md)复用授权的官方说明，保留来源位置和 Apache 通知，报告缺失语言/说明。不从说明推断运行时策略，不把上游账号或资源示例用于 Go 测试。
- 十一项基础能力及手写 ECS/STS 契约测试必须先验收，才可开发生成器；仅元数据调研可以提前，接口存在不等于通过。
- 先写路线文档，再建立真实 issue 依赖，最后各分支实现；依赖图不允许循环。生成器与基准测试分开跟踪。
- 生产生成使用固定官方 Darabonba 源码、导入模块和官方语义投影。Go 检查前使用 Node 22 执行前端检查和测试。
- 修改固定元数据、补充配置、模板或手写校验后重新生成，不直接编辑生成文件。生成器修改需运行 sdkgen check，产品修改还需 product-check。
- DSL 与元数据不一致时，保留双语决策文件和机器可读审核记录，不自动忽略；Explorer 浏览器证据与 CLI 证据分开记录。
- sources/darabonba、sources/openapi-meta 的上游原始源码、README 和许可通知保持原始字节。项目自己编写的来源指南分语言维护；不把上游许可改标为 MIT。

## 验证与沟通

- 不重复验证或盲目重试；只有修改有意义或失败原因已经明确时才重复相关检查。
- 可在浏览器确认的结果，提供准确 URL 和检查项，不用 curl 等请求重复验证。
- DNS 以 CloudFormation 服务端状态和输出为准，不做本地 DNS 验证；必要时提供用户核验步骤。
- 用户未明确要求时，不启动子代理。
