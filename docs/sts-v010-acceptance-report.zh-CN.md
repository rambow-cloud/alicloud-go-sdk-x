# v0.1.0 STS 验收证据

[English](sts-v010-acceptance-report.md)

## 当前执行路线（2026-10-09）

- 按[STS、ECS 与 VPC 路线](sts-ecs-vpc-path.zh-CN.md)执行：先完成 #60，再完成 ECS #74 和 VPC #75，最后进行 #61 发布与索引。
- #60 改为如实记录的代理消费者验收；独立人工体验移到可选后续项 #76，人工记录仍为 NOT RUN，不阻塞本次发布。
- 下文原先首版只发布 STS、#60 必须等待人工验收的安排属于历史记录，已被本次决策替代。
- 保留实际技术、真实调用和来源演练证据；代理测试耗时不能当作人工任务耗时。
- 两个产品验收 issue 通过后再发布，本变更不创建版本标签。

- #60 的技术验收和独立开发者门槛分别记录。
- #58/PR #63、#59/PR #64 已合并。
- 真实/消费者运行的实现版本为 `38cf05ac458d2e6ed3af165350fb77a10e7e8817`，通过 `f31ad13e4ed3ba9974dfd5d5a9efff8c01aaf3a6` 进入 main。
- Go 1.27.1、Windows/amd64、 Node 22.21.1、语义解析器 2.2.1；Linux race/Windows CI 单独记录。
- 验收分支补维护门禁/ 测试数据/文档，不改变已接受 STS 来源、模型、策略、生成字节或签名运行时行为。

### 限定覆盖

| 原生操作           | 发现/转换为 IR/生成代码/编译 | 独立离线契约/Examples                                     | 真实                                                                     |
| ------------------ | ---------------------------- | --------------------------------------------------------- | ------------------------------------------------------------------------ |
| AssumeRole         | PASS                         | PASS；凭据提供者与缓存消费及小测试替身                    | PASS #55：三次签发、四次 ECS 读取、真实 900 秒到期续期与清理             |
| GetCallerIdentity  | PASS                         | PASS；完整输出、线路/存在/错误/取消                       | PASS #60：六身份字段及缺失状态与 CLI 一致、metadata 一致、一次 ACS3 尝试 |
| AssumeRoleWithOIDC | PASS                         | PASS；匿名 RPC/token 编码/凭据提供者隔离/完整响应/脱敏    | NOT RUN；成功联邦预先位于 v0.1.0 必需真实范围外                          |
| AssumeRoleWithSAML | PASS                         | PASS；匿名 RPC/身份断言 编码/凭据提供者隔离/完整响应/脱敏 | NOT RUN；成功联邦预先位于 v0.1.0 必需真实范围外                          |

- 四个固定操作输出 19 个可达命名/内联模型；空 identity Input 不算官方 DSL 模型。
- 生成报告的编译/真实仍为 not-assessed，本报告另提供验收证据。
- 本范围 STS 无原生分页/状态等待器任务，其他产品保留各自证据，不升级为首版产品验收。

### 消费者与维护

- 隔离模块固定上述官方 STS v2.1.0、OpenApi v2.1.13、tea v1.3.13 和 go.sum；本 SDK 显式本地 replace，独立开发者记录检出提交。
- 两 SDK 测试数据都跑四操作，本任务另用原生 AssumeRole → 凭据提供者与缓存 → 两次角色签名 identity 读取。
- 三个外部测试及程序 PASS，无应用 translator/刷新循环；未测性能或官方 credentials 库刷新。
- 不表示源码兼容或优劣，只比较固定版本 context/options/envelope/测试替身的实际差异。

- 真实修订演练比较上述 c321 与 d2c，保留原始许可/blob/固定导入。
- 原生字段/模型/类型/ 必需性及语义说明不变，initializer 签名/endpoint 映射、匿名 handoff 和说明坐标变化。
- 历史 v2 初始化与 Anonymous/callApi 不支持，不能当作 ACS3 验收。
- 当前候选四操作发现、 生成、独立编译契约/Examples PASS；旧策略及历史签名选择写前失败。
- 候选 STS 字节与生产相同，原来源/IR/策略/生成物未变。
- [命令/审核](sts-source-rehearsal.zh-CN.md)及 [机器证据](acceptance/sts-source-rehearsal.json)区分真实历史变化、无变化生产更新与合成负面案例。

- 2026-10-08 11:33:27 UTC 的授权只读 identity 使用 CLI 登录后的 `oss-sftp` OAuth/STS 内存快照显式注入。
- [脱敏证据](acceptance/sts-identity-live.json)无密钥/账号/ARN/token/ 原始响应，不建云资源，不代表原生 Profile/OAuth 续期。
- Explorer 浏览器记 NOT RUN， 与 CLI/SDK 证据分开。
- 签名读取/签名/cache/AssumeRole 来源行为保留，匿名扩展不影响该路径，因此复用 #55 自然续期证据，不重复已完成云操作。

### 门槛与交接

- 后续 #68 新增 config.LoadDefaultConfig、可续期原生 CLI Profile/OAuth，外部消费者共 9 项测试，包含原生 loader。
- 2026-10-08 15:35:31 UTC、上述实现 SHA 的真实原生验证通过 3 次生成身份读取、1 次主动 OAuth 交换、缓存复用、仅认证字段持久化、会话重建复用及锁释放。
- [脱敏证据](acceptance/profile-oauth-live.json)不含身份/凭据，未启动 CLI 子进程或新增云资源。
- 刚登录访问令牌有效，真实 refresh-token 轮换及等待自然 OAuth 到期仍 NOT RUN。
- [配置契约](default-configuration.zh-CN.md)记录 PascalCase/camelCase 协议决策。
- 独立人员需使用新增行为，人员 NOT RUN 及发布/索引门槛保持不变。

- 限定映射：AC-01/02 强类型 context/options、完整建模字段/存在/协议/严格 JSON；AC-05 显式凭据提供者与缓存及 #55；AC-06 审核显式重试与签发不重试；AC-07 endpoint 解析及独立覆盖；AC-08 中间件/OTel 生命周期与敏感隔离；AC-09 结构化安全错误/取消/ metadata；AC-10 窄测试替身/离线工具；AC-11 双语指南/英文 Go docs/Examples/许可来源， 实际发布索引仍待完成；AC-12 确定生成/审核策略/真实来源演练，均不表示更广 Beta。
- 另见[补充消费者评审](sts-consumer-review.zh-CN.md)，新外部包证据不替代独立人员验收。
- PR 记录前端/56 Node 案例、两生成器、doccheck、格式/双语、vet、 Go 测试/Examples、隔离消费者及再生成 STS；技术合并前准确提交 Linux race/Windows CI 必须通过。

- UX-04/05 独立文档任务：**NOT RUN，用户安排另一位 Go 开发者**。
- 实现者测试数据不证明独立耗时/成功。
- 按固定 PR 提交交付[任务](sts-consumer-acceptance.zh-CN.md)和 [双语模板](sts-independent-result-template.zh-CN.md)，实际独立结果通过前 #60 开放；必需证据阻塞 #61 发布/索引，不表示 tag 或 pkg.go.dev 已索引。

## 参考资料

- [STS v2.1.0](https://github.com/alibabacloud-go/sts-20150401/tree/v2.1.0)

## 固定提交

- `c321394a58d9b6e513fabb898ee9857a2c6df852`
- `d2c0338636a58a6cafc5316d2ed5d158f1f5b162`
- `656ce39dda0b89ae743645ba1328974a937dc780`
