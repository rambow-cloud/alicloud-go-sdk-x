## English

### Scope and dependencies

Member of #57; blocked by #60. Publish a scoped experimental v0.1.0 using docs/sts-v0.1.0.md, not the broader ECS/VPC/STS Beta claim. This issue records future release work; the current planning change does not authorize immediate publication or invent a tag/index result.

### Acceptance

- [ ] At the exact candidate commit, all required STS generation/behavior/consumer/maintenance/scoped-live/docs cases pass; required failures/skips remain blocking.
- [ ] English/Chinese release notes explain the four-action pinned STS scope, explicit providers, AWS-style API/migration limits, defaults/retry/redaction and OIDC/SAML live NOT RUN. Existing service/ecs,vpc and services/ compatibility artifacts are available with separate evidence, not promoted to v0.1.0 product acceptance.
- [ ] Review module/Go 1.27/JSON v2, MIT runtime and Apache generated LICENSE/NOTICE; record clean build, Examples/doccheck, Linux race/Windows CI and reviewed source lock.
- [ ] Publish immutable v0.1.0 tag/release through the authorized release workflow; never overwrite tags. Record release/commit evidence.
- [ ] Open same-version pkg.go.dev STS/credentials/stscreds pages in a browser, inspect version/license/exported docs/Examples and record actual indexing evidence. No curl/DNS substitute.
- [ ] Complete parent issue and close milestone only after all required evidence, then align Project statuses/labels.

### Documentation and verification

Follow paired docs/releasing.md and docs/sts-v0.1.0.md. Publication and indexing are separate from local validation; release remains blocked pending acceptance.

## 中文

归属 #57，依赖 #60，限定实验 v0.1.0 不宣称更广 Beta；本规划不立即发布或编造 tag/索引。按精确候选提交核全部 STS 必需验收；双语说明四操作/固定来源/显式 provider/AWS 范式迁移/默认与重试脱敏/OIDC-SAML 真实 NOT RUN，其他包保留但不升级产品验收声明。检查 Go 1.27/JSON v2/许可/来源、Examples/docs/CI，再按授权流程不可变发布，浏览器核同版本 pkg.go.dev。发布/索引证据齐全后才关闭父 issue/milestone 并同步状态，不用 curl/DNS 替代。
