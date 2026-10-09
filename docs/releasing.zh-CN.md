# 发布与索引

[English](releasing.md)

## 当前执行路线（2026-10-09）

- 按[STS、ECS 与 VPC 路线](sts-ecs-vpc-path.zh-CN.md)执行：先完成 #60，再完成 ECS #74 和 VPC #75，最后进行 #61 发布与索引。
- #60 改为如实记录的代理消费者验收；独立人工体验移到可选后续项 #76，人工记录仍为 NOT RUN，不阻塞本次发布。
- 下文原先首版只发布 STS、#60 必须等待人工验收的安排属于历史记录，已被本次决策替代。
- 保留实际技术、真实调用和来源演练证据；代理测试耗时不能当作人工任务耗时。
- 两个产品验收 issue 通过后再发布，本变更不创建版本标签。

- 首版目标为 [v0.1.0 STS](sts-v0.1.0.zh-CN.md)，父项 #57/发布 #61、同名 milestone/Project 3。
- 发布前遵循其明确限定实验门槛，旧多产品候选 Beta 排期不是此版本前提，更广验收仍开放。
- 双语版本说明记录四个固定 STS 操作的实际覆盖及 OIDC/SAML 真实成功联邦 NOT RUN， 其他包保留且披露独立验收限制。
- 本规划不发布 tag/release，不表示已索引。

- 公开 module 为 github.com/rambow-cloud/alicloud-go-sdk-x，Go 1.27；原创运行时/工具使用 MIT，生成产品定义/说明保留 Apache-2.0、包内 LICENSE/NOTICE 和固定来源引用，发布前核对两者。
- 本次工作不创建发布标签。
- CI 和基础验收通过后，只发布明确记载的操作范围，并同步中英文版本说明。
- 实验 API 使用 v0； 索引标签不能改写，必要时用新版本及 retract。
- v1 承诺兼容，后续主版本遵守 Go 版本路径规则。

- 在浏览器打开上述 pkg.go.dev 地址，缺失时点击 Request，检查版本、许可证、概述、导出字段文档及 Examples， 再检查 credentials 和其他公共包。
- badge 是索引入口，不是已索引证据；不用本地 curl/DNS 验证页面。
- 索引失败先诊断公开访问、module 路径、Go 支持与许可证，再重试；参考来源与所列参考资料一致。

## v0.1.0 发布步骤

- 使用已准备的[双语版本说明](releases/v0.1.0.zh-CN.md)及[准确发布/索引清单](sts-v010-release-checklist.zh-CN.md)。
- 必需独立验收仍 NOT RUN，不表示 tag/索引；清单附只读发布门禁及准确同版本浏览器地址。
- 已有授权覆盖门槛通过后的发布；缺人验收不是再次请求发布许可。

## 来源与证据链接

- [参考链接](https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x)
- [参考链接](https://pkg.go.dev/about#adding-a-package)
- [参考链接](https://pkg.go.dev/license-policy)
