## English

### Problem and scope

The user selects STS end-to-end Darabonba delivery as v0.1.0. Current pinned STS discovers four actions but emits only AssumeRole. Preserve the official DSL/parser -> normalized IR -> our Go backend -> shared runtime route, Go 1.27/JSON v2, AWS conventions and bilingual pkg.go.dev support. Authoritative scoped plan: docs/sts-v0.1.0.md. This experimental scope supersedes the broader candidate Beta schedule without claiming overall Beta or full-cloud coverage.

### Acceptance and dependencies

All four pinned actions must be generated, compiled and independently offline-tested/documented. Preserve explicit source providers and generated STS provider/cache composition; retain #51/#53/#55 evidence. Require scoped live AssumeRole renewal and GetCallerIdentity, external consumer/official-v2 comparison and independent-user task evidence, real source-update rehearsal, immutable release/licenses/migration notes and same-version pkg.go.dev browser evidence. Successful OIDC/SAML live federation and federation provider helpers are explicitly future scope; record NOT RUN, never claim live success. No STS pagination is invented.

Child tasks will be linked after creation; parent remains open until every required release case passes. Child membership is not a dependency on this parent. This planning task creates no tag/cloud resources.

### Verification and documentation

Node 22 frontend checks/tests before Go generation/product checks, doccheck, vet, tests/Examples, Linux race and Windows CI. Record discovered/lowered/emitted/compiled/offline/live/published separately; required SKIP/NOT RUN keeps the gate open. Maintain docs/sts-v0.1.0.md, AGENTS.md, development/acceptance/release guides and paired issue specs. Maintain one status label and Project Status per issue.

## 中文

用户指定 v0.1.0 为 STS 完整 Darabonba 链路；当前四操作仅生成 AssumeRole。保留官方 DSL/parser→IR→本项目 Go 后端→共享 runtime、Go 1.27/JSON v2、AWS 范式及双语文档。四操作生成/编译/离线契约必需；保留 #51/#53/#55 provider/cache/真实续期证据，补 GetCallerIdentity 真实读取、固定消费者/官方 v2 对比及独立用户任务、真实来源升级、不可变发布/迁移/许可及同版本 pkg.go.dev。OIDC/SAML 真实联邦及 provider helper 预先列为后续范围，NOT RUN 不写成通过。不造 STS 分页，不宣称整体 Beta；子项归属不依赖父项，发布验收前父项保持打开。验证及双语更新/状态维护如上；本规划不创建 tag/云资源。

### Delivery checklist / 交付清单

- [ ] #58 [Feature]: Generate requestless STS GetCallerIdentity from official Darabonba DSL
- [ ] #59 [Feature]: Generate anonymous STS OIDC and SAML RPC operations from Darabonba
- [ ] #60 [Maintenance]: Accept all four generated STS operations and rehearse source updates
- [ ] #61 [Release]: Publish the scoped v0.1.0 STS release and verify pkg.go.dev

Milestone: https://github.com/rambow-cloud/alicloud-go-sdk-x/milestone/4
Project: https://github.com/orgs/rambow-cloud/projects/3

Accepted baseline: #51, #53, #55 (delivered; no repeat required without relevant changes).
中文：已有交付证据纳入，不无理由重跑。
