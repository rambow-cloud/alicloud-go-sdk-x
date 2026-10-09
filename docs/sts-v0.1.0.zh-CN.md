# v0.1.0 STS 交付

[English](sts-v0.1.0.md)

## 当前执行路线（2026-10-09）

- 按[STS、ECS 与 VPC 路线](sts-ecs-vpc-path.zh-CN.md)执行：先完成 #60，再完成 ECS #74 和 VPC #75，最后进行 #61 发布与索引。
- #60 改为如实记录的代理消费者验收；独立人工体验移到可选后续项 #76，人工记录仍为 NOT RUN，不阻塞本次发布。
- 下文原先首版只发布 STS、#60 必须等待人工验收的安排属于历史记录，已被本次决策替代。
- 保留实际技术、真实调用和来源演练证据；代理测试耗时不能当作人工任务耗时。
- 两个产品验收 issue 通过后再发布，本变更不创建版本标签。

### 权威与范围

- 用户于 2026-10-08 指定首版目标为 **v0.1.0：STS 完整走通 Darabonba 生成链路**。
- 这个限定的实验版本优先于旧 ECS/VPC/STS 候选 Beta 的排期，但保留其更广泛的 [验收标准](product-acceptance.zh-CN.md)及历史证据；完成本里程碑不代表整体 Beta。
- 本次规划不创建版本标签或新云资源。

- 采用固定官方 STS 2015-04-01 DSL/导入模块、官方语义解析器、规范化完整 IR、 本项目 Go 后端及公共运行时。
- `service/sts` 交付固定来源的四操作：AssumeRole、 GetCallerIdentity、AssumeRoleWithOIDC、AssumeRoleWithSAML；升级来源时重新核对清单， 不表示永久全量或全云覆盖。
- 保留 Go 1.27、直接 encoding/json/v2、AWS 调用范式、 显式凭据提供者、输入与配置的所有权、安全错误信息和带许可的双语文档。

### 基线及缺口

- 规划时基线发现四操作，只输出 AssumeRole；OIDC/SAML 的 DSL_PROTOCOL_PROFILE 对应 Anonymous RPC/doRPCRequest，GetCallerIdentity 的 DSL_OPERATION_SIGNATURE 对应没有请求模型参数。
- 分 issue 修复前端/IR/后端/运行时契约，不手写操作、不静默签名匿名请求、不隐式发现密钥、不直接改生成文件。
- 匿名认证是显式逐操作协议契约， 不允许需要签名的操作省略凭据。

- 已接受 #51 完整 DSL 凭据提供者与缓存组合、#53 STS 优先指南及显式来源、#55 真实 900 秒到期续期（三次签发、四次生成 ECS 读取，临时 IAM 已全部清理）。
- 这些关闭 issue 作为交付证据纳入，不为新里程碑重建云环境或重复同一真实测试。

### 顺序与证据

1. 支持无请求模型 RPC 签名，生成 GetCallerIdentity，保留操作名、context、Options、
   测试替身接口、准确完整响应及元数据。
2. 支持审核的 Anonymous RPC/doRPCRequest，生成 OIDC/SAML，保留线路字段及缺失语义，
   不签名/不读取凭据提供者，默认格式隐藏令牌或身份断言及响应凭据。前两项可在独立
   issue 分支推进。
3. 两项完成后验收四操作：可重现的离线请求格式、响应与错误场景契约、凭据提供者与缓存回归、固定
   版本外部消费者及官方 v2 对比、授权 GetCallerIdentity 真实证据，以及真实上游版本
   升级/漂移演练。已有 #55 继续作为 AssumeRole/续期证据，除非相关修改使其失效。
4. 准备兼容/迁移说明、许可、双语版本说明、对应版本 pkg.go.dev 检查；发布前门槛
   全通过才发布，发布/索引证据完成后才能关闭里程碑。

| 门槛        | 必需证据                                                                                             |
| ----------- | ---------------------------------------------------------------------------------------------------- |
| 生成        | 四操作发现/转换为 IR/生成代码，完整可达模型，确定性前端/IR/Go 检查，选中不支持行为写前失败           |
| 行为        | 四操作编译及独立离线协议/字段/响应/错误/取消测试与外部 Example，签名/匿名分离和敏感格式              |
| 凭据        | 显式凭据提供者，生成 AssumeRole→凭据提供者与缓存→消费者，保留所有权/并发/取消契约及 #55 限定真实续期 |
| 消费者/维护 | 固定版本的外部 STS 任务与官方 v2 示例，独立 Go 用户仅按文档完成任务的证据，真实上游升级/漂移审核     |
| 真实范围    | #55 AssumeRole/续期及授权 GetCallerIdentity；SDK/CLI 与 Explorer 浏览器证据分别记录                  |
| 文档/发布   | 英文 pkg.go.dev 注释、配对指南、许可、Linux race/Windows CI、不可变 v0.1.0 标签/发布及同版本网页证据 |

- OIDC/SAML 真实成功联邦调用需要另行授权的 IdP/凭据提供者/身份断言 环境，**预先明确不属于首版必需真实范围**；记录 NOT RUN，并在版本说明披露，不替代四操作必需离线正确性。
- 用户后续 #68 修正使默认配置/原生 Profile/OAuth 续期成为发布前必需，遵循 [更新路线](default-configuration.zh-CN.md)；联邦 credential-凭据提供者辅助组件仍为后续范围。
- 本例 STS 没有分页/状态等待器任务，不编造分页，保留共享引擎。
- ECS 仅作为已有角色凭据消费者；ECS/VPC 全产品验收和基准 #20 独立。

### GitHub 协调

- 建立 `v0.1.0` milestone 和关联本仓库的 Projects v2 项目 `alicloud-go-sdk-x: v0.1.0 STS`。
- Milestone 表示发布范围，issue 承载验收/依赖/证据， Project Status 表示执行进度。
- 每 issue 只有一个状态 label：满足前提为 ready/Todo； 执行中为 in-progress/In Progress；依赖未完成为 blocked/Todo 并列出前提；验收完成为 done/Done。
- 随 issue 状态变化维护 Project，不表示已自动同步；代码合并不等于发布完成。

- 父 issue 跟踪子项直到发布门槛通过；子项关联父 issue 表示归属，不把父 issue 作为阻塞依赖，保持无环。
- 已有关闭 STS 证据加入里程碑/项目；历史生成器/基础 issue 和其他项目保留原范围。
- 不随意设截止日期；选择 v0.1.0 下一项前阅读本文。

- [Milestone v0.1.0](https://github.com/rambow-cloud/alicloud-go-sdk-x/milestone/4) | [Project](https://github.com/orgs/rambow-cloud/projects/3) | [父 issue #57](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/57)

| Issue | 工作                                 | 阻塞前提                     |
| ----- | ------------------------------------ | ---------------------------- |
| #58   | 无请求模型 GetCallerIdentity 生成    | 已接受生成器基线             |
| #59   | 匿名 OIDC/SAML RPC 生成              | 已接受生成器基线             |
| #60   | 四操作验收、消费者对照及来源升级演练 | #58、#59；已完成 #51/#53/#55 |
| #61   | 限定发布及对应版本 pkg.go.dev 证据   | #60                          |

- 下一实现先 #58，再 #59，不手写 API 绕过门槛。
- 父项汇总交付，不阻塞子项。

## STS 交付状态

- #58/PR #63 与 #59/PR #64 已合并，四个固定 STS 操作输出并保留原生签名/匿名分离。
- #60 交付消费者/真实来源/真实 identity 证据及独立开发者验收包，用户安排的 Go 开发者提交实际结果前 UX 门槛保持开放。
- #61 发布/索引以此必需验收为前提，下文规划数量是旧基线不是当前覆盖。
- 见[验收证据](sts-v010-acceptance-report.zh-CN.md)。
