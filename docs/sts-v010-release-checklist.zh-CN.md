# v0.1.0 发布检查清单

[English](sts-v010-release-checklist.md)

- 用户于 2026-10-09 调整顺序：#60 代理 STS 验收、ECS #74、VPC #75，最后执行 #61 发布与索引。
- 遵循[当前路线](sts-ecs-vpc-path.zh-CN.md)，原先首版只发布 STS、#60 必须等待人工验收的规则已替代。
- 独立人工体验是可选 #76，原记录仍为 NOT RUN；代理测试不能写成人工证据。
- [v0.1.0 发布及七页浏览器核验](releases/v0.1.0-publication.zh-CN.md)已通过，网页证据来自用户报告。STS 收尾阶段没有创建标签或发布；本次发布在 ECS/VPC 验收后执行。

## 必需门禁

- [本轮消费者证据](release-consumer-refresh.zh-CN.md)记录集成后的 SDK 提交，同时保留历史记录。

1. 将实际代理消费者结果写入 acceptance/sts-agent-result.json，固定测试过的 SDK 和任务版本。保留明确范围的真实调用、来源、Profile 证据及限制。
2. 完成 #74、#75 及各自的产品记录。生成数量、空的终止页、跳过的 waiter 检查，不能代替尚未执行的必需验收项。#47/PR #48 单独评审。
3. 最终 main 提交须通过 Linux Go 1.27 race、Windows Go 1.27、自动化及 issue 关联检查。审核 module/JSON v2、来源和策略锁、MIT 运行时与 Apache 生成包 LICENSE/NOTICE，将配对发布说明更新为实际验收范围。
4. 执行只读发布检查。它检查 STS 代理任务、两个产品报告、原生 OAuth/真实调用/来源证据、干净的 main 及验收版本之后的修改；不创建标签或发布。产品证据缺失或 NOT RUN 时继续阻塞。
5. 确认 #60/#74/#75 已关闭、远端不存在 v0.1.0，记录准确候选版本及 CI 链接，然后在 #61 中执行下方命令。

## 门禁通过后的发布命令

```powershell
node .github/scripts/sts-release-check.cjs
git tag -a v0.1.0 <verified-main-commit> -m "v0.1.0: generated STS, ECS and VPC"
git -c credential.helper= -c 'credential.helper=!gh auth git-credential' -c credential.interactive=false push origin refs/tags/v0.1.0
gh release create v0.1.0 --verify-tag --title "v0.1.0: generated STS, ECS and VPC" --notes-file docs/releases/v0.1.0.md --prerelease
```

- 已有完成发布的授权；用户现在要求等 ECS/VPC 完成后再执行。
- 不覆盖版本标签；若标签已创建但发布失败，检查远端状态并完成同一版本。

## 同版本浏览器证据

| 包                   | 准确地址                                                                                 |
| -------------------- | ---------------------------------------------------------------------------------------- |
| service/sts          | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/service/sts          |
| service/ecs          | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/service/ecs          |
| service/vpc          | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/service/vpc          |
| credentials          | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/credentials          |
| feature/stscreds     | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/feature/stscreds     |
| config               | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/config               |
| feature/profilecreds | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/feature/profilecreds |

- 在浏览器打开这些页面，检查 v0.1.0、许可证、完整操作和字段文档、已审核的适配器及可运行 Example。
- 页面不存在时使用 Request，并记录真实索引结果和时间；本地测试、curl 或 DNS 不证明索引完成。
- 确认索引器接受 Go 1.27/直接 JSON v2，不静默降低版本要求。
- 真实发布与索引完成后再结束 #61/#57 和 milestone，并同步 label 与 Project 状态。
