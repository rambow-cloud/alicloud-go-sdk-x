# v0.1.0 发布清单

[English](sts-v010-release-checklist.md)

- 用户后续 #68 修正还要求在新的 #60 人员交接前完成默认配置/原生 CLI Profile/OAuth 及限定证据；docs/acceptance/profile-oauth-live.json 与历史手动快照分别记录，发布前必需。

- #61 保持开放：#60 必需独立验收 NOT RUN，发布/同版本浏览器索引也 NOT RUN。
- 说明/命令已准备，此变更不创建标签。
- 用户已授权完成发布，等待的是实际验收证据，不再次请求许可。

1. 接收用户安排的独立 Go 开发者完成的[任务报告](sts-independent-result-template.zh-CN.md)，仅将真实结果转写到
   `docs/acceptance/sts-independent-result.json`，保留脱敏源码/耗时/障碍/版本。全部必需
   项通过才完成 #60。
2. 审核合并最终说明/证据，准确 main 候选提交需要 Linux Go 1.27 race、Windows、
   automation/issue 及隔离消费者/真实来源演练通过；核 module/JSON v2、MIT/Apache 和
   固定来源。`node .github/scripts/sts-release-check.cjs`拒绝未完成的人验收或之后的 SDK 变化，不创建标签/发布。
3. 确认远端 v0.1.0 标签不存在、#60 已关闭，记录准确 SHA/CI URLs，验收后才执行共享
   命令创建/推送不可变注释标签并发布。禁止覆盖；标签创建后发布失败须检查远端状态，
   完成同标签发布，不重建或移动。
4. 记录 tag 目标/release URL，在浏览器打开下表中的同版本页面，核 v0.1.0、对应 Apache/MIT、导出 API/字段说明及 Example。缺失则使用 Request
   并记录真实时间/结果；不从本地检查/curl/DNS 宣称索引。尤其确认索引器支持 Go 1.27/
   直接 JSON v2，不支持时保持尚未运行/失败并解决，不静默降低版本。
5. 将实际发布/浏览器证据写入本清单或明确链接的双语报告；全部必需通过后关闭 #61/#57/
   milestone 4 并同步标签/Project 3。合并准备 PR 不等于发布或索引，不关闭 issue。

## 可运行命令与示例

```powershell
git tag -a v0.1.0 <verified-main-commit> -m "v0.1.0: generated STS"
git -c credential.helper= -c 'credential.helper=!gh auth git-credential' -c credential.interactive=false push origin refs/tags/v0.1.0
gh release create v0.1.0 --verify-tag --title "v0.1.0: generated STS" --notes-file docs/releases/v0.1.0.md --prerelease
```

## pkg.go.dev 页面与检查项

| 包             | 对应版本的页面                                                                           | 检查项                                                                     |
| -------------- | ---------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| STS            | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/service/sts          | 版本 v0.1.0、Apache 许可证、四个操作、模型和字段文档、可运行 Example       |
| Credentials    | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/credentials          | 版本、MIT 许可证、显式凭据提供者（含 AnonymousProvider）、Cache 和 Example |
| STS helper     | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/feature/stscreds     | 版本、MIT 许可证、完整 DSL 客户端构造、所有权和续期说明、Example           |
| Config         | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/config               | 版本、MIT 许可证、默认加载器、选项、优先级和离线 Example                   |
| Native Profile | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/feature/profilecreds | 版本、MIT 许可证、原生模式、OAuth 持久化、所有权、错误和离线 Example       |
