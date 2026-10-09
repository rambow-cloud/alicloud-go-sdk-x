# [Release]: Deliver v0.1.0 after STS, ECS and VPC acceptance

### Current user-approved route (2026-10-09)

- Complete #60 through truthfully recorded implementation-agent consumer acceptance. Independent human usability moves to optional #76 and remains NOT RUN.
- Then complete ECS #74 and VPC #75 before #61 publication/indexing.
- Follow docs/sts-ecs-vpc-path.md and AGENTS.md. Preserve pinned complete official Darabonba/parser -> IR -> shared Go backend/runtime, Go 1.27/JSON v2 and paired docs.
- Generation counts, selected-field live checks and optional skipped cases are distinct from complete required product acceptance.

### Delivery checklist

- [x] #58 requestless GetCallerIdentity generation
- [x] #59 anonymous OIDC/SAML generation
- [x] #68 native default configuration/Profile/OAuth
- [x] #72 STS provider and Darabonba reuse review
- [ ] #60 agent STS consumer closeout and source/technical evidence
- [ ] #74 ECS product acceptance
- [ ] #75 VPC product acceptance
- [ ] #61 immutable publication and same-version pkg.go.dev browser evidence

- #47 remains the scoped ECS/VPC live-evidence track; review PR #48 separately. No new live call is implied by this route.
- #76 is a separate optional usability follow-up, not a release blocker.
- Parent/milestone stay open until required product and publication/indexing cases pass; parent membership does not block children.
- Project 3 and milestone v0.1.0 track the required work.

<details>
<summary>Historical STS-only schedule and evidence (superseded where it conflicts with the current route)</summary>

### Problem and scope

- The user selects STS end-to-end Darabonba delivery as v0.1.0.
- The planning-time baseline discovered four actions but emitted only AssumeRole.
- Preserve the official DSL/parser -> normalized IR -> our Go backend -> shared runtime route, Go 1.27/JSON v2, AWS conventions and bilingual pkg.go.dev support.
- Authoritative scoped plan: docs/sts-v0.1.0.md.
- This experimental scope supersedes the broader candidate Beta schedule without claiming overall Beta or full-cloud coverage.

### Acceptance and dependencies

- All four pinned actions must be generated, compiled and independently offline-tested/documented.
- Preserve explicit source providers and generated STS provider/cache composition; retain #51/#53/#55 evidence.
- Require scoped live AssumeRole renewal and GetCallerIdentity, external consumer/official-v2 comparison and independent-user task evidence, real source-update rehearsal, immutable release/licenses/migration notes and same-version pkg.go.dev browser evidence.
- Successful OIDC/SAML live federation and federation provider helpers are explicitly future scope; record NOT RUN, never claim live success.
- No STS pagination is invented.

- Child tasks will be linked after creation; parent remains open until every required release case passes.
- Child membership is not a dependency on this parent.
- This planning task creates no tag/cloud resources.

### Verification and documentation

- Node 22 frontend checks/tests before Go generation/product checks, doccheck, vet, tests/Examples, Linux race and Windows CI.
- Record discovered/lowered/emitted/compiled/offline/live/published separately; required SKIP/NOT RUN keeps the gate open.
- Maintain docs/sts-v0.1.0.md, AGENTS.md, development/acceptance/release guides and paired issue specs.
- Maintain one status label and Project Status per issue.

### Delivery checklist

- [x] #58 [Feature]: Generate requestless STS GetCallerIdentity from official Darabonba DSL
- [x] #59 [Feature]: Generate anonymous STS OIDC and SAML RPC operations from Darabonba
- [ ] #60 [Maintenance]: Accept all four generated STS operations and rehearse source updates
- [ ] #61 [Release]: Publish the scoped v0.1.0 STS release and verify pkg.go.dev

- Milestone: https://github.com/rambow-cloud/alicloud-go-sdk-x/milestone/4 Project: https://github.com/orgs/rambow-cloud/projects/3

- Accepted baseline: #51, #53, #55 (delivered; no repeat required without relevant changes).

### Planning delivery

- Planning PR #62 merged at 03d3adab31dab5ec04ab4690fad810f3aea099e2.
- Milestone, linked public Project, native children/dependencies, baseline evidence, paired route/specs and AGENTS constraints delivered.
- Local bilingual/format/doccheck/vet/Go gates and exact-head CI passed; PASS CI on 99f6e4fd86bfaee44239ffd08e9c32dc9c35823b: Linux Go 1.27 race, Windows Go 1.27 and automation; linked-issue also passed. [CI run](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37764613292).
- Parent remains open: child implementation, acceptance and release/indexing are still pending.

## Current delivery

- #58/#59 are closed; PR #63/#64 merged.
- Technical #60 PR #65 merged at d9e56dd24a633ce1e2df241307a64dc25d6c0658 with all exact-head Linux race/Windows/consumer/rehearsal checks passing.
- All four current STS actions emit/compile/offline-test; authorized GetCallerIdentity live comparison passes.
- Independent docs-only developer results remain NOT RUN; #60 stays open.
- Release preparation PR #66 is under review; #61 publication/indexing remains blocked by required #60 evidence.
- Milestone/Project remain open.

## Release preparation delivered

- PR #66 merged at 3e6a5a720e9ba6afd5c995102b7b7b56a7f955b3; exact head 40443a47edd0c8c544c2ef7485723d1fd2f6a649 passed Linux race, Windows, automation, linked-issue, consumer and real-source rehearsal CI.
- Notes, licensing/migration guidance, read-only guard and same-version browser checklist are on main.
- Independent developer records remain NOT RUN; #60/#61 and milestone stay open, blocked on actual acceptance/publication/indexing.
- No tag or indexing claim.

## User-approved native configuration correction

- [x] #68 default configuration and native CLI Profile/OAuth credentials.

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
