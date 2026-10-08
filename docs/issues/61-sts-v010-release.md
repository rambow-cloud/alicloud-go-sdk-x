# [Release]: Publish the scoped v0.1.0 STS release and verify pkg.go.dev

### Scope and dependencies

- Member of #57; blocked by #60.
- Publish a scoped experimental v0.1.0 using docs/sts-v0.1.0.md, not the broader ECS/VPC/STS Beta claim.
- This issue records future release work; the current planning change does not authorize immediate publication or invent a tag/index result.

### Acceptance

- [ ] At the exact candidate commit, all required STS generation/behavior/consumer/maintenance/scoped-live/docs cases pass; required failures/skips remain blocking.
- [ ] English/Chinese release notes explain the four-action pinned STS scope, explicit providers, AWS-style API/migration limits, defaults/retry/redaction and OIDC/SAML live NOT RUN. Existing service/ecs,vpc and services/ compatibility artifacts are available with separate evidence, not promoted to v0.1.0 product acceptance.
- [ ] Review module/Go 1.27/JSON v2, MIT runtime and Apache generated LICENSE/NOTICE; record clean build, Examples/doccheck, Linux race/Windows CI and reviewed source lock.
- [ ] Publish immutable v0.1.0 tag/release through the authorized release workflow; never overwrite tags. Record release/commit evidence.
- [ ] Open same-version pkg.go.dev STS/credentials/stscreds pages in a browser, inspect version/license/exported docs/Examples and record actual indexing evidence. No curl/DNS substitute.
- [ ] Complete parent issue and close milestone only after all required evidence, then align Project statuses/labels.

### Documentation and verification

- Follow paired docs/releasing.md and docs/sts-v0.1.0.md.
- Publication and indexing are separate from local validation; release remains blocked pending acceptance.
