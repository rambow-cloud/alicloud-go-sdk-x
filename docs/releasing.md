# Releasing and pkg.go.dev / 发布与索引

[English](#english) | [中文](#中文)

## English

Canonical public module: github.com/rambow-cloud/alicloud-go-sdk-x, MIT license, Go 1.27.
This work does not create a release tag. Release only the documented operation coverage
after CI and foundation acceptance; update equivalent English/Chinese release notes.
Use v0 for experimental APIs. Never overwrite indexed tags; use a new version and retract
when necessary. v1 promises compatibility; later major modules use Go's versioned paths.

Open https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x in a browser. If absent,
click Request. Inspect version, license, overview, exported field docs and external Examples,
then inspect credentials and the other public packages. The badge is an indexing entry point,
not evidence that indexing occurred. Do not use local curl/DNS to verify the page.
If indexing fails, diagnose public access, module path, Go support and license before retrying.
Sources: https://pkg.go.dev/about#adding-a-package and https://pkg.go.dev/license-policy.

## 中文

公开 module 为 github.com/rambow-cloud/alicloud-go-sdk-x，MIT 许可证、Go 1.27。本次工作不创建发布标签。
CI 和基础验收通过后，只发布明确记载的操作范围，并同步中英文版本说明。实验 API 使用 v0；
索引标签不能改写，必要时用新版本及 retract。v1 承诺兼容，后续主版本遵守 Go 版本路径规则。

在浏览器打开上述 pkg.go.dev 地址，缺失时点击 Request，检查版本、许可证、概述、导出字段文档及 Examples，
再检查 credentials 和其他公共包。badge 是索引入口，不是已索引证据；不用本地 curl/DNS 验证页面。
索引失败先诊断公开访问、module 路径、Go 支持与许可证，再重试；参考来源与英文章节相同。
