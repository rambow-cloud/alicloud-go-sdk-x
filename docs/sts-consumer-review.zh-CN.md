# STS 补充消费者评审

[English](sts-consumer-review.md)

## 当前执行路线（2026-10-09）

- 按[STS、ECS 与 VPC 路线](sts-ecs-vpc-path.zh-CN.md)执行：先完成 #60，再完成 ECS #74 和 VPC #75，最后进行 #61 发布与索引。
- #60 改为如实记录的代理消费者验收；独立人工体验移到可选后续项 #76，人工记录仍为 NOT RUN，不阻塞本次发布。
- 下文原先首版只发布 STS、#60 必须等待人工验收的安排属于历史记录，已被本次决策替代。
- 保留实际技术、真实调用和来源演练证据；代理测试耗时不能当作人工任务耗时。
- 两个产品验收 issue 通过后再发布，本变更不创建版本标签。

- 本报告归属 [#60](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/60)。
- **五项补充技术用例通过**，集成前还需准确提交的 Linux race/Windows CI。
- 本次为实现代理单独执行的评审，不冒充独立 Go 开发者仅按文档验收。
- 人员结果仍 NOT RUN，#60/#61 及发布/索引门禁保持开放。

- 评审 main 基线为 `3e6a5a720e9ba6afd5c995102b7b7b56a7f955b3`；Go 1.27.1、 Windows/amd64、Node 22.21.1。
- 隔离消费者模块本地 replace SDK，官方对比固定 STS v2.1.0、OpenApi v2.1.13、Tea v1.3.13。
- 本次只补测试与文档，运行时、生成物、来源和策略沿用该基线。

### 方法与发现

- [consumer_review_test.go](../examples/stsacceptance/consumer_review_test.go) 的独立 `main_test` 包只导入 SDK 公开包，自建 HTTP 传输实现/测试替身，不复用原消费者测试数据或内部工具；HTTP 传输实现不建立网络连接，密钥、身份、token、ARN 全为虚构值。
- 组合依据上述 STS、凭据、缓存、产品与匿名协议指南。
- 评审还查阅公开声明及签名证据，因此不标为无人协助、仅按文档的独立人员任务。

| 用例                      | 观察                                                                                                                                                                                                         | 结果 |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---- |
| 身份/options/取消         | 原生字段与 RequestID/status/attempt metadata 正确；操作 endpoint 覆盖不影响下一次调用；调用中取消保留 errors.Is 与 OperationError                                                                            | PASS |
| 原生角色/凭据提供者与缓存 | 构造凭据提供者后修改调用方 session，不影响原始 session 和显式 900 秒；取消首个等待者不取消共享签发，24 个并发消费者使用角色 token，仅签发一次，不写应用刷新循环                                              | PASS |
| 小测试替身/错误           | 只实现 AssumeRoleAPI；errors.As/Is 保留同一 APIError，已取消读取不进入测试替身                                                                                                                               | PASS |
| 匿名联邦                  | OIDC/SAML 准确保留 `+ / = &` 和空格、公共 query 与空 POST body；AnonymousProvider 及不可用自定义凭据提供者的四调用均零来源读取；签名 identity 拒绝 marker，构造拒绝 nil/带类型的 nil；默认模型格式隐藏 token | PASS |
| 非安全重试/错误格式       | Standard 下 AssumeRole 遇 503 只执行一次；结构错误保留 code/status/request ID/attempts，显式 Message 可读取，默认 Error 字符串不带 Message                                                                   | PASS |

- 2026-10-08 13:15:03 UTC 首次执行四项通过；cache 用例因评审者误以为 `Credential=<key>` 后跟 AWS 风格斜线而失败。
- 阿里云 ACS3 在 SignedHeaders 前使用逗号； 修正测试数据断言并补提前错误报告后，单独重跑该项通过，无需修复 SDK。
- 这印证设计边界： Go 调用范式参考 AWS，线路语义保留阿里云。

- 本次官方固定版本的四操作测试数据也通过。
- 该版本使用 RuntimeOptions 和 Body envelope， 本 SDK 使用 context-first、服务函数式选项、原生输出加 Metadata 及窄操作接口。
- 未测性能、未覆盖所有官方版本、未评估官方 credentials 库刷新能力。

- 发现并修正一项文档问题：旧报告将 AC-06 至 AC-12 与不准确的能力说明分组对应。
- 现按[权威验收表](product-acceptance.zh-CN.md)依次对应重试、endpoint、中间件/OTel、 错误、测试、文档和生成，不增加覆盖声明。

### 验证与限制

- 英文记录中的三个命令在根目录设置本地 Go cache 后运行；首次失败及定向修正均保留记录， 合并本地结果覆盖五项。
- 最终完整隔离模块 Linux race/Windows 由 PR 准确提交 CI 提供， 相关文档/vet/根模块测试也在 PR 记录。
- 测试耗时不是独立开发者任务耗时。
- 本次无新增真实调用、IAM/IdP 创建或来源升级；既有[真实续期](live-sts-renewal.zh-CN.md)、 [身份记录](acceptance/sts-identity-live.json)、[来源演练](sts-source-rehearsal.zh-CN.md) 仍为分别记录的证据。

- 本限定评审未发现阻塞运行时缺陷。
- OIDC/SAML 成功真实联邦仍 NOT RUN，预先位于必需真实范围外；本 STS 任务无原生分页/状态等待器。
- 本报告不证明原生 Profile/OAuth 发现、更多产品验收、独立人员任务易用性/耗时、发布或同版本 pkg.go.dev 浏览器索引。

## 可运行命令与示例

```text
go -C examples/stsacceptance test -count=1 -v -run '^TestConsumerReview' ./...
go -C examples/stsacceptance test -count=1 -v -run '^TestConsumerReviewNativeRoleCacheAndCanceledWaiter$' ./...
go -C examples/stsacceptance test -count=1 -v -run '^TestPinnedOfficialWorkload$' ./...
```
