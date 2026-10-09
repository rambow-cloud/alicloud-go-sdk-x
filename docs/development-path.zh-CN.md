# 开发路线

[English](development-path.md)

- 下一步：[补齐 VPC RPC 生成支持](vpc-rpc-completion.zh-CN.md)，增加 query/表单分别绑定、整模型 GET query 及显式 simple 数组。[RPC 扩展 #83](dsl-rpc-expansion.zh-CN.md)已合并。

## 统一服务路径

- 在 #61 前完成[服务整合 #81](service-consolidation.zh-CN.md)：移除兼容桥，将基础层及 provider 契约迁移到完整 DSL 客户端，并刷新消费者证据。本路线替代此前保留兼容桥的要求。

## ECS、VPC 的明确范围验收

- #74/#75 按执行前确定的 [ECS 验收](ecs-product-acceptance.zh-CN.md)和 [VPC 验收](vpc-product-acceptance.zh-CN.md)矩阵完成。
- #85 扩展后的 RPC 生成清单为 ECS 380 个、VPC 403 个操作。#74/#75 消费者验收仍是固定提交的历史证据，扩展生成数量不表示所有操作均经过真实调用验收。离线消费者契约与真实调用的选定字段证据分别记录，不支持的 DSL 操作继续说明原因。
- 真实实例 token、waiter 状态迁移及非空 VPC 续页仍为排除项，保留 NOT RUN 或 SKIP，由后续项 #79 跟踪；不宣称整体 Beta 或全云验收。
- 消费者验收由实现代理执行，独立人工体验和自动化测试耗时继续区分。
- STS、ECS、VPC 证据固定到共用消费者和 CI 的受测版本；发布及同版本 pkg.go.dev 索引仍由 #61 完成。产品收尾不创建版本标签。

- [完整 DSL 真实调用证据](product-live-validation.zh-CN.md): 完整 DSL ECS/VPC 的历史只读证据（#47）：镜像读取两页，实例和 VPC 返回空页，真实 waiter 检查跳过；#74/#75 产品验收仍单独进行。

## 当前执行路线（2026-10-09）

- 按[STS、ECS 与 VPC 路线](sts-ecs-vpc-path.zh-CN.md)执行：先完成 #60，再完成 ECS #74 和 VPC #75，最后进行 #61 发布与索引。
- #60 改为如实记录的代理消费者验收；独立人工体验移到可选后续项 #76，人工记录仍为 NOT RUN，不阻塞本次发布。
- 下文原先首版只发布 STS、#60 必须等待人工验收的安排属于历史记录，已被本次决策替代。
- 保留实际技术、真实调用和来源演练证据；代理测试耗时不能当作人工任务耗时。
- 两个产品验收 issue 通过后再发布，本变更不创建版本标签。

## 当前状态

- #58、#59 已合并：生成器支持本次范围内的四个 STS 操作。
- #68 已合并：默认配置与原生 CLI Profile/OAuth 已有明确范围的真实验证记录。
- #60 的独立开发者验收仍为 NOT RUN。
- #61 的发布与 pkg.go.dev 索引仍为 NOT RUN。
- #70 调整文档格式：中英文分文件、改善中文、issue 仅用英文。
- 文档格式调整不改变 SDK 功能和发布验收要求。
- #72：[STS 可复用性评审](sts-reuse-review.zh-CN.md)。发布前统一 provider 校验、隔离每次调用的选项，并验证改名后的 DSL 操作仍能通过同一套解析器和生成后端。

## 已确认的路线与历史验收

- 用户于 2026-10-08 最新修正要求发布 v0.1.0 前完成[默认配置与原生 CLI Profile/OAuth #68](default-configuration.zh-CN.md)，先实现 loader/Profile/cache/刷新契约与证据，再更新 #60 消费者交接，之后 #61。
- 该路线优先于下文旧全面禁止发现及推迟原生 Profile/OAuth 范围；长期密钥仍需显式启用，直接服务构造保持只接受凭据提供者校验，#53/#55 历史证据不证明原生 OAuth 续期。

- 用户指定首版现为 [v0.1.0 STS](sts-v0.1.0.zh-CN.md)：同名 milestone、Project 3、父项 #57。
- 下一步 #58 无请求 GetCallerIdentity，再 #59 匿名 OIDC/SAML RPC；两项共同解锁 #60 四操作验收/消费者/来源升级，之后 #61 发布/索引。
- 限定实验排期优先于下文旧整体候选 Beta 队列，保留其标准/历史，不表示该 Beta。
- #51/#53/#55 为已接受证据；本规划不新建云对象或标签。

- 用户指定的下一项凭据工作 [#53](credentials.zh-CN.md) 将 STS 凭据提供者与缓存作为主要应用指南。
- 长期 AK/SK 和环境来源必须显式注册凭据提供者，遵循 AWS 的凭据提供者注入范式；Config/Options 不新增裸密钥字段或隐式环境回退。
- 保留自定义凭据提供者与现有 StaticProvider API，发送请求前拒绝 nil/带类型的 nil 来源， 通过运行时/生成客户端及操作覆盖测试证明，并提供完整离线 STS→ECS Example 和配对指南， 然后继续体验对照。

- 用户现授权本地沙箱外[真实 STS 续期](live-sts-renewal.zh-CN.md)，并创建专用角色。
- 先建独立 issue， 只创建授权的专用角色/来源，使用最小权限并清理；区分强制刷新与真实时间的自然到期续期， 再记录 AC-05 的真实证据。

- #55 已完成三次真实签发和四次生成 ECS 读取，包含缓存复用、强制刷新及真正 900 秒到期后的自动续期，全部新建临时 IAM 对象已清理。
- 这是限定 AC-05 证据；独立体验及其他产品/真实要求继续分别验收。

- 后续 issue 依照[产品验收](product-acceptance.zh-CN.md)（#49），候选 Beta 范围、AC/UX 编号、 离线/真实/用户体验/发布证据及必需项状态共同决定完成。
- 完整 DSL STS 凭据提供者组合与主要凭据指南 #51/#53 已合并，[#55](live-sts-renewal.zh-CN.md)记录限定真实时间续期；下一步为消费者/官方 v2 对比、剩余授权真实缺口、来源升级演练及发布。
- 基准 #20 独立，Smithy 仍为隔离实验。

- 首项后续为[完整 DSL STS 凭据提供者组合 #51](sts-credentials.zh-CN.md)，对应 AC-05/UX-04，保留旧辅助组件并建立离线集成；真实角色/刷新及独立开发者任务证据仍分开验收。

- 五阶段实现和评审修复 [#44](capability-role-review.zh-CN.md) 已于 2026-10-08 集成 main， [评审证据](generator-integration.zh-CN.md) 记录验收范围与合并提交；同模型游标/状态等待器角色重叠写前拒绝。
- 后续能力或协议扩展另建明确 issue，以当前固定 RPC 实现为起点。

- 用户于 2026-10-07 确认的权威路线：完整官方产品 DSL → 官方 Darabonba 语义语义解析器 → 规范化操作/模型/绑定 IR → 本项目 Go 后端 → 已有运行时。
- [product-generator-roadmap.md](product-generator-roadmap.zh-CN.md) 和 AGENTS.md 规定新路线， 优先于冲突的旧元数据优先、逐操作快照/补充配置、字段选择、手写说明前置要求。
- 下文旧阶段及 issue 引用仅作为历史验收记录。

- 按来源规范化 → 完整操作/模型发现及 IR → 批量 Go 输出 → 少量能力策略 → 文档自动化/ 协议扩展推进。
- 每阶段同步双语文档、相关测试、Example 及 pkg.go.dev，最后的文档自动化不表示前面可以推迟文档。
- 固定 canonical 元数据在规范化后可选补充/交叉验证， 新扫描不以每 API 的元数据/补充配置为前提。

- 首个里程碑是规范化真实表示并交付完整固定 ECS 的 inventory/覆盖报告，再批量生成支持的 RPC。
- #31/PR #32 是五操作兼容桥，不是产品全量验收。
- Overlay 仅补审核的兼容例外、分页/状态等待器、幂等、敏感字段及特殊校验，不重述每个模型/字段。
- 未支持操作明确列原因，选中不支持行为写前失败；先规范化 Filter 索引绑定及 itemName 响应包装再判冲突，浏览器、CLI 本地校验、真实 HTTP 调用证据分别记录。

- 文档 → 建/更新真实 issue 依赖 → 每项独立分支 → 实现/验证/记录 → 关联 PR。
- 路线父任务在所有阶段验收前保持打开，保留已验收运行时/AWS 范式、Go 1.27/JSON v2、 标准库核心；基准独立。

- 目标是先完成统一 Go 运行时，用少量手写 ECS/STS 参考操作验证，然后建设产品生成器。
- 要求 Go 1.27、直接使用 encoding/json/v2；核心运行时导入保持标准库，OpenTelemetry 为可选集成。

| 阶段            | 交付                                                         | 门槛                                      |
| --------------- | ------------------------------------------------------------ | ----------------------------------------- |
| A：规格         | 双语文档、英文 issue、无环依赖与语言规则                     | 写在代码之前                              |
| B：请求基础     | 分阶段中间件、endpoint、ACS3、HTTP、结构化错误、测试辅助     | 离HTTP 请求/响应与并发契约                |
| C：可靠性与身份 | 显式凭据链、过期缓存、设有次数和时间上限的重试、STS 辅助组件 | 取消、新鲜度、可重放与幂等性              |
| D：参考行为     | 手写 ECS、统一分页、状态等待器、可选 OTel                    | 强类型用法与跨能力集成                    |
| E：基础验收     | 十一项能力、示例、双语指南、Linux race 与 Windows CI         | 实现及测试，而非仅有接口                  |
| F：生成器       | 固定元数据生成强类型客户端、编码、文档与规则                 | #19 已通过，按 #8 下 #21 → #22 → #23 执行 |

- 依赖顺序：测试/中间件/endpoint/错误及签名 → HTTP 运行时 → ECS/STS → 分页/状态等待器/AssumeRole 集成。
- 凭据链/缓存与重试策略可先独立实现，再接运行时；OTel 依赖中间件和请求元数据。
- 签名单独验收不依赖 HTTP，集成验收归 HTTP issue，解除之前的循环。

- 十一项均必须交付：分页、状态等待器、重试/退避、测试替身接口、凭据提供者、STS 辅助组件、endpoint、中间件、 OTel、结构化错误、测试辅助。
- 共享公共扩展契约，产品特殊规则采用具体类型。

- 首版协议为 OpenAPI ACS3-HMAC-SHA256。
- 参考操作：ECS DescribeRegions、DescribeInstances（选定响应字段）、 DescribeInstanceStatus，以及 STS AssumeRole。
- 范围不包含全量 ECS/STS、OSS/SLS 数据面或真实云验收。
- Endpoint 使用明确规则，不支持的地域报错，不猜测域名。

- 保留默认不重试；显式启用的标准策略受次数与时间限制，且可以取消、带 jitter/Retry-After，仅适用于明确幂等且可重放的操作。
- Waiter 总期限与单次重试分开；凭据刷新有刷新超时，并合并并发请求，一个等待者取消不影响其他调用者。
- 不隐式执行外部进程或读取 metadata 凭据。

- 每项 issue 包含包文档、确定性外部 Example、对应双语指南。
- GitHub 是执行状态来源。
- #1/#2 保留历史完成记录，#3-#7/#9 细化为基础任务，#8 只负责生成器；基准独立跟踪。
- 新增 issue 的真实编号在索引与里程碑中登记，不预设编号。
- 协议及设计参考来源与所列参考资料一致。

- 生成器架构、支持范围、来源及验收见 [generator.md](generator.zh-CN.md)；#21 开始前，基础门槛已在 28684e4 通过。
- 下一真实产品门槛为 #24 → #25，见 [generator-expansion.md](generator-expansion.zh-CN.md)。
- 验收后，扩展更多 profile 前按 #26 → #27 → #28 → #29 修复 review，见 [AWS 风格修复路径](aws-style-remediation.zh-CN.md)。
- 后续前端重构使用官方产品 DSL 与官方语义解析器，交付顺序及偏差验证见 [Darabonba 迁移路径](darabonba-migration.zh-CN.md)。

## STS 交付状态

- #58/PR #63 与 #59/PR #64 已合并，四个固定 STS 操作输出并保留原生签名/匿名分离。
- #60 交付消费者/真实来源/真实 identity 证据及独立开发者验收包，用户安排的 Go 开发者提交实际结果前 UX 门槛保持开放。
- #61 发布/索引以此必需验收为前提，下文规划数量是旧基线不是当前覆盖。
- 见[验收证据](sts-v010-acceptance-report.zh-CN.md)。

## 参考资料

- [ACS3](https://help.aliyun.com/zh/sdk/product-overview/v3-request-structure-and-signature)
- [metadata](https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/)
- [AWS middleware](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/middleware.html)
- [AWS testing](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/unit-testing.html)
