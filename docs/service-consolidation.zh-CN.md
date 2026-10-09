# 统一服务包路径

[English](service-consolidation.md)

- 决定：SDK 客户端统一使用 `service/<product>`，在 v0.1.0 发布前移除只覆盖五个操作的 `services/` 兼容桥。
- 本次用户确认的路线替代此前保留兼容桥及其生成产物的要求。历史验收证据和上游源码通知继续保留。
- 这是 v0 阶段的不兼容变更。旧导入和选择性响应模型移除，调用者需要使用原生指针字段及完整响应容器。

## 工作范围

- 将根包的跨能力契约测试迁移到完整生成的 ECS、STS 客户端。
- 角色凭据 provider 使用生成的 STS 适配器。`NewAssumeRoleProvider` 改为接受 `service/sts` 类型；保留 `NewAssumeRoleProviderFromClient`，转发到同一实现，供当前消费者继续使用。
- 在完整 DSL 路径保留 provider/cache 轮换、输入所有权、取消、分页、waiter、重试和遥测验证。
- 删除旧客户端包、旧 Go 输出器及生成入口。解析器测试仍使用的来源规范化和元数据交叉验证证据继续保留。
- `sdkgen generate/check` 改为完整产品生成的入口，避免现有命令重新生成兼容桥；保留 `product-generate/product-check` 命令名。
- 同步配对指南、生成文档模板、项目约束和 CI。
- 重新执行受影响的消费者验收，记录实际版本。历史真实调用证据仍作为历史记录；本次不调用云服务、不创建标签、不发布。

## 验证方式

- Node 22 前端检查与测试、语言检查和自动化测试。
- 完整 DSL 生成一致性、包文档、格式、vet 和 Go 测试。
- 独立模块的 STS、ECS/VPC 消费者测试及可运行示例；CI 验证 Linux race 和 Windows。
- 生产代码不再导入 `services/`，任何生成命令都不能恢复旧包。支持清单保持 STS 4 个、ECS 283 个、VPC 296 个操作。
- 运行一次 `sdkgen product-check`；`sdkgen check` 调用同一后端，不重复执行等价生成门禁。

## 迁移方式

- 将 `github.com/rambow-cloud/alicloud-go-sdk-x/services/<product>` 改为 `github.com/rambow-cloud/alicloud-go-sdk-x/service/<product>`。
- 使用 `NewFromConfig`；可选字段按需要传指针，响应读取原生容器。
- 将生成的 `sts.AssumeRoleInput` 传给 `stscreds.NewAssumeRoleProvider`。`FromClient` 函数名调用同一实现。
- 公共运行时、凭据 provider 和缓存继续复用。

- #81 任务版本 e33e5f93d856027569969056e9d4c25e92e51212：Node 22 前端检查及 57 项测试、25 项自动化测试、完整 DSL 生成一致性、文档检查（15 个公共包）、vet、根模块测试及受版本控制 Go 文件的格式检查全部通过。
- 独立模块消费者通过：STS 11 项、ECS 10 项及 14 个子用例、VPC 八项及十个子用例；两个模块的 vet 和离线程序均通过。更新后的机器记录固定到同一版本。
- 最终提交还需关联 PR 上的 Linux race、Windows CI。本次没有新的真实调用、标签或发布。

- #81 修正工具指南，使完整发现先于 Go 输出。全部 29 项消费者用例和 24 个产品子用例在 29d468ad5f8999e92b4e60e4140cb8432efb8a13 上通过，机器记录已固定到这一版本。Go、运行时、来源及策略行为与已完成本地检查的 e33e5f93d856027569969056e9d4c25e92e51212 相同，未受影响的门禁复用原结果；最终提交仍需 PR #82 的 CI 通过。
