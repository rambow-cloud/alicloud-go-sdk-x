# [Release]: Publish v0.1.0 after STS, ECS and VPC acceptance

### Current scope and dependencies (2026-10-09)

- The user explicitly defers publication until ECS and VPC are accepted after #60.
- Required native dependencies: #60, ECS #74 and VPC #75. Member of #57; parent membership is not a blocker.
- Optional independent human follow-up #76 is outside this release gate. Its evidence must not be replaced with agent tests.
- Follow docs/sts-ecs-vpc-path.md, docs/releasing.md and docs/sts-v010-release-checklist.md.

### Acceptance

- [ ] All required STS/ECS/VPC task and product evidence passes at reviewed revisions, with explicit workload/live/coverage boundaries.
- [ ] Paired release notes describe actually accepted products, native pagination/waiter policy, credentials/defaults/retry/migration and excluded live federation/other protocols.
- [ ] Go 1.27/JSON v2, MIT runtime and Apache generated LICENSE/NOTICE/source lock, docs/Examples and exact-final-head Linux race/Windows CI pass.
- [ ] The read-only release guard accepts STS agent evidence and both product records without source/workload drift; incomplete product gates remain blocking.
- [ ] Create immutable v0.1.0 tag/release only after required issues close. Never overwrite an indexed tag.
- [ ] Inspect same-version pkg.go.dev STS/ECS/VPC/credentials/helper/config/Profile pages in a browser and record actual version/license/docs/Example evidence.
- [ ] Close #57/milestone only after actual publication/indexing and align status labels/Project.

- Status: deferred/blocked on the required product work. No tag, release or indexing is executed during #60 closeout.

<details>
<summary>Historical STS-only schedule and evidence (superseded where it conflicts with the current route)</summary>

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

## Technical delivery; required gate open

- PR #66 merged at 3e6a5a720e9ba6afd5c995102b7b7b56a7f955b3.
- PASS exact-head Linux race/Windows/all CI at 40443a47edd0c8c544c2ef7485723d1fd2f6a649: https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37774664521/job/113302399799 https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37774663031/job/113302394677 https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37774664382/job/113302398598 https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37775577654/job/113305431471 https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37774664521/job/113302399832 https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37774664521/job/113302399553

- Release notes/checklist/read-only guard are prepared. #60 independent acceptance, actual immutable publication and same-version browser indexing remain required; #61 stays open.

## User-approved native configuration correction

- The 2026-10-08 user correction makes #68 required before updated consumer acceptance and v0.1.0 publication.
- Supersede blanket no-discovery/deferred-native-Profile/OAuth scope while preserving provider-only direct service construction and explicit long-lived AK opt-in.
- Follow docs/default-configuration.md; retain historical #53/#55/#60 evidence, but it does not establish native OAuth refresh.
- Independent developer acceptance must use the new behavior; publication/indexing remain pending.

## Native configuration delivery and current handoff

- #68 / PR #69 are delivered at 4a6aec10c4a34c5b79032ab2161681974a81a12f with passing exact-head CI at ae2d6cba5ae92b3a9a7bbd7b66909066b9aa4803: https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37802808980.
- Native Profile identity/exchange/cache/persistence PASS; live refresh-token rotation and natural OAuth expiry wait NOT RUN.
- Latest independent consumer-kit pin: ae2d6cba5ae92b3a9a7bbd7b66909066b9aa4803.
- Follow docs/sts-consumer-acceptance.md and docs/default-configuration.md including native default loading.
- Independent developer tasks remain NOT RUN; #60 and publication/indexing #61 remain OPEN.
- Milestone v0.1.0 remains OPEN; do not publish before those actual gates pass.

</details>
