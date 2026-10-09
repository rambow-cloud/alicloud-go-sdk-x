# STS 代理消费者验收收尾

[English](sts-agent-closeout.md)

## 范围与验收者

- 对应 #60。用户于 2026-10-09 指定由实现代理完成消费者验收。
- 这里记录代理执行，不是独立人工体验，也不是无协助、仅按文档完成的人工任务。
- 可选人工后续项 #76 保留原始 NOT RUN 记录；自动化测试耗时不能替代人工任务耗时。
- 遵循[新路线](sts-ecs-vpc-path.zh-CN.md)：完成 STS，再完成 ECS #74、VPC #75，最后进行 #61 发布。
- 本次无需新的云调用、资源创建、来源升级、版本标签或发布。

## 固定的消费者任务

- SDK、任务版本和实际环境记录在[机器证据](acceptance/sts-agent-result.json)中。
- `examples/stsacceptance` 是独立模块，只导入公共 SDK 包；官方对照固定为 STS v2.1.0、OpenApi v2.1.13、Tea v1.3.13。
- 通过显式 local replace 使用当前检出的 SDK；官方对照依赖不进入运行时模块。
- 所有夹具使用虚构 Profile、凭据及模拟 HTTP，传输层不建立网络连接。
- 生产操作和模型仍来自官方 Darabonba 与完整 IR，凭据适配器、缓存和 Profile 基于生成接口组合。

## 验收任务

| 任务        | 必需消费者证据                                                                                         |
| ----------- | ------------------------------------------------------------------------------------------------------ |
| 配置与身份  | 加载 StsToken，OAuth 刷新、轮换、持久化和重建，长期凭据显式启用，选项、Metadata 和取消                 |
| 凭据与缓存  | 原生生成的 AssumeRole 接入共享 provider/cache，输入所有权、角色签名消费、24 次并发读取及取消等待者隔离 |
| Mock 与错误 | 小型 AssumeRoleAPI，errors.Is/As，预取消，签发不重试及安全错误格式化                                   |
| 匿名操作    | OIDC/SAML 原生字段和 token 编码，不读取来源凭据、不签名，显式匿名标记及敏感字段格式化                  |
| 官方对照    | 固定官方 v2 的四操作夹具和已记录调用方式差异；不测量性能或宣称优越性                                   |

- 外部 OAuth 测试只使合成的 access token 过期，验证轮换后的 refresh token 复用、配置持久化以及重建后的缓存凭据。
- 这些测试证明离线行为，不把真实 OAuth 轮换、自然到期或真实联邦身份的 NOT RUN 改成 PASS。
- `.github/scripts/sts-consumer-record.cjs` 只记录必需测试的成功事件。缺失、重复、跳过或失败不能记作 PASS，公开记录不包含原始输出。

## 验证

```powershell
go -C examples/stsacceptance test -json -count=1 ./...
go -C examples/stsacceptance run .
node --test .github/scripts/*.test.cjs
node .github/scripts/check-doc-language.cjs
go run ./internal/cmd/doccheck
go vet ./...
go test ./...
```

- Go 门禁前先执行 Node 22 前端检查和测试；CI 另执行两个生成器检查、来源升级演练、Linux race 和 Windows 测试。
- 最终提交的 CI 链接记录在 #60 及其 PR；相关行为没有变化时复用已接受的来源和真实调用证据。
- 只读发布检查同时要求 STS 代理证据和 ECS/VPC 产品报告；任一产品仍为 NOT RUN 时保持阻塞。

## 结果

- #60 初次收尾为 PASS：五组任务的 11 项外部消费者测试全部通过，固定于 [3d54426](https://github.com/rambow-cloud/alicloud-go-sdk-x/commit/3d54426d15b6b79b238138a0d8a52bf9ebc0d62c)。
- 环境：Go 1.27.1、windows/amd64、Node 22.21.1。执行时间：2026-10-09 00:14:07.135–00:14:09.855 UTC。
- 命令含编译共耗时 2.720 秒。按 Go 测试事件累计，身份组为 0.080 秒、官方对照为 0.010 秒，其余组经 Go 舍入后为零。这些数值是自动化执行耗时，不是人工任务耗时或性能基准。
- PASS：可运行的外部消费者、官方 STS v2.1.0 对照、前端检查及 57 项前端测试、25 项自动化测试、两个生成器检查、文档检查（18 个公共包）、vet、根模块 Go 测试及格式检查。
- PASS：配对文档和本地链接检查。本次无需修改运行时、生成 SDK、来源、IR 或能力策略。
- 发布按预期保持阻塞：只读检查返回 `BLOCKED: ecs product acceptance is not PASS`。ECS 和 VPC 验收记录仍为 NOT RUN。
- 合并前，将最终提交的 Linux race、Windows 和自动化结果记录到收尾 PR 及 #60。机器记录仍绑定实际受测的任务提交，后续仅补充证据的修改不会改变这一绑定。
- 独立人工体验仍为 NOT RUN，由可选后续项 #76 跟踪。真实联邦身份和 OAuth 自然到期不属于本次离线结果。

- #75 更新共用任务后，全部 11 项在 85795cf3cf1604afe59b0e8af03d5df0c9d3ba38 上重新通过；机器记录包含最新时间和耗时，上文 #60 初次执行详情继续作为历史记录保留。运行时及 STS 来源行为未改变，本次更新覆盖新增消费者和 CI 的检查基线。

- #81 服务整合：生成的 service/sts 与公共适配器在 e33e5f93d856027569969056e9d4c25e92e51212 上通过全部 11 项；更新后的机器记录包含实际耗时，原收尾及真实调用证据继续作为历史记录。

- #81 修正工具指南，使完整发现先于 Go 输出。全部 29 项消费者用例和 24 个产品子用例在 29d468ad5f8999e92b4e60e4140cb8432efb8a13 上通过，机器记录已固定到这一版本。Go、运行时、来源及策略行为与已完成本地检查的 e33e5f93d856027569969056e9d4c25e92e51212 相同，未受影响的门禁复用原结果；最终提交仍需 PR #82 的 CI 通过。

- #83 共用 RPC 及运行时回归：01f25a9a571c3b59024dfbef7e2e96f7605f4a55 上通过 11 项原有 STS 用例及固定官方 JSON helper 对比。这是代理自动化证据，历史真实调用提交及人工体验限制不变。
