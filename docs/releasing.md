# Releasing and pkg.go.dev / 发布与索引

[English](#english) | [中文](#中文)

## English

The first release target is [v0.1.0 STS](sts-v0.1.0.md), parent #57 / release #61,
milestone v0.1.0 / Project 3. Apply its explicitly scoped experimental gate before
publication; the older multi-product candidate Beta schedule is not a prerequisite
for this version. Its broader acceptance remains open. Include all four pinned
STS actions' actual coverage and OIDC/SAML successful live federation NOT RUN in
paired release notes. Other packages stay available with their separate acceptance
limits. This planning update does not publish a tag, release or indexing claim.

Canonical public module: github.com/rambow-cloud/alicloud-go-sdk-x, Go 1.27.
Original runtime/tooling use MIT; generated product definitions/prose retain Apache-2.0,
with package LICENSE/NOTICE and pinned-source references. Inspect both before release.
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

首版目标为 [v0.1.0 STS](sts-v0.1.0.md)，父项 #57/发布 #61、同名 milestone/Project 3。
发布前遵循其明确限定实验门槛，旧多产品候选 Beta 排期不是此版本前提，更广验收仍开放。
双语版本说明记录四个固定 STS 操作的实际覆盖及 OIDC/SAML 真实成功联邦 NOT RUN，
其他包保留且披露独立验收限制。本规划不发布 tag/release，不宣称已索引。

公开 module 为 github.com/rambow-cloud/alicloud-go-sdk-x，Go 1.27；原创 runtime/工具使用
MIT，生成产品定义/说明保留 Apache-2.0、包内 LICENSE/NOTICE 和固定来源引用，发布前
核对两者。本次工作不创建发布标签。
CI 和基础验收通过后，只发布明确记载的操作范围，并同步中英文版本说明。实验 API 使用 v0；
索引标签不能改写，必要时用新版本及 retract。v1 承诺兼容，后续主版本遵守 Go 版本路径规则。

在浏览器打开上述 pkg.go.dev 地址，缺失时点击 Request，检查版本、许可证、概述、导出字段文档及 Examples，
再检查 credentials 和其他公共包。badge 是索引入口，不是已索引证据；不用本地 curl/DNS 验证页面。
索引失败先诊断公开访问、module 路径、Go 支持与许可证，再重试；参考来源与英文章节相同。

## v0.1.0 execution / v0.1.0 执行

### English

Use the prepared [paired release notes](releases/v0.1.0.md) and [exact release/indexing checklist](sts-v010-release-checklist.md). Required independent acceptance remains NOT RUN; no tag or indexing is claimed. The read-only release guard and same-version browser URLs are recorded there. Existing authorization covers completion after gates pass; missing human evidence is not an invitation to reconfirm publication permission.

### 中文

使用已准备的[双语版本说明](releases/v0.1.0.md)及[准确发布/索引清单](sts-v010-release-checklist.md)。必需独立验收仍 NOT RUN，不宣称 tag/索引；清单附只读发布门禁及准确同版本浏览器地址。已有授权覆盖门槛通过后的发布；缺人验收不是再次请求发布许可。
