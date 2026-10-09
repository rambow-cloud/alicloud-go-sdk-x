# [Maintenance]: Accept all four generated STS operations and rehearse source updates

### Current scope: user-approved 2026-10-09 closeout

- Complete #60 with implementation-agent consumer acceptance. Record the reviewer type truthfully; this is not independent human UX evidence.
- Independent human developer tasks move to #76, which does not block #60 or the revised v0.1.0 schedule. Keep the human JSON record NOT RUN.
- Finish ECS #74 and VPC #75 before release/indexing #61. STS closeout creates no tag or release.
- Follow docs/sts-ecs-vpc-path.md. Complete official Darabonba/parser -> IR -> our Go backend/runtime remains authoritative.

### Remaining acceptance

- [ ] Run the latest pinned standalone consumer and official STS v2.1.0 comparison; record actual profile/identity, native provider/cache, mock/errors/cancellation and anonymous field/authentication results.
- [ ] Extend external consumer coverage for OAuth refresh-token rotation, persistence/reconstruction, source credential isolation and explicit long-lived opt-in using synthetic profiles and HTTP.
- [ ] Publish paired agent closeout evidence and machine-readable tasks with exact SDK/workload revisions, commands, versions and real results. Automated test duration is not human task time.
- [ ] Preserve the four-action offline/compile coverage, scoped live identity/renewal and real source-rehearsal evidence. No duplicate cloud run unless relevant behavior changes.
- [ ] Update the read-only release guard so STS agent acceptance cannot bypass unfinished ECS/VPC gates. Preserve safe diagnostics, source-change detection and native OAuth requirements.
- [ ] Update AGENTS and paired route/consumer/acceptance/release guides, actual dependencies, milestone/Project and status labels.
- [ ] Final-head automation, Linux Go 1.27 race and Windows Go 1.27 CI pass before completion.

### Evidence already delivered

- #58/#59 generated all four pinned STS actions and complete models.
- #51/#53/#55 provide credential composition and scoped real renewal; #65/#67 provide consumer/official-v2/source/live-identity evidence.
- #68/#69 provide native configuration/Profile/OAuth and scoped live exchange/persistence evidence.
- #72/#73 remove provider ownership/validation coupling and prove renamed DSL parser/emitter/runtime reuse.
- Successful live OIDC/SAML and natural/live OAuth rotation remain explicitly scoped NOT RUN; agent fixtures do not promote them to live PASS.

<details>
<summary>Historical scope and delivery evidence (superseded where it conflicts with the current user decision)</summary>

### Problem and scope

- Emission alone does not establish a usable modern STS SDK.
- Member of #57; depends on #58, #59, accepted #51/#53/#55.
- Do not recreate #55 IAM resources or repeat unchanged natural-renewal evidence.

### Acceptance

- [ ] Pin external consumer and official-v2 versions/Go/OS/workloads; demonstrate GetCallerIdentity, native AssumeRole and provider/cache consumption without response translators or refresh loops, structured errors and small mocks. Record independent Go-user docs-only task success/time/obstacles; no unmeasured superiority claim.
- [x] Record four actions' discovered/lowered/emitted/compiled/offline status and scoped live results separately. Verify authorized generated GetCallerIdentity against safe CLI identity assertions without publishing account IDs/ARNs/credentials. #55 remains scoped renewal evidence unless relevant changes invalidate it. No new cloud/IdP creation authorized by this issue.
- [x] OIDC/SAML have independent complete offline protocol and negative fixtures; successful live federation explicitly NOT RUN/outside required live scope, documented before execution.
- [x] Rehearse one real pinned upstream source revision in an isolated workspace: inventory/binding/model/auth/doc/license drift, reviewed compatible/incompatible changes, deterministic regenerate and safe failure/rollback. Do not silently update production source.
- [ ] Paired acceptance report, consumer guidance and source-update evidence map to AC-01/02/05/06/07/08/09/10/11/12 and scoped UX-04/05. Required FAIL/SKIP/NOT RUN keeps acceptance open.

### Verification

- Node frontend/test gates, both sdkgen checks, external module build/tests, doccheck/vet/tests/Examples, Linux race/Windows; scoped safe live harness only with available explicit authorization.
- Explorer/browser evidence is distinct.

## Source-drift finding and scoped guard

- The real 2025-06-30 STS source assigns @signatureAlgorithm=v2 in its product initializer.
- Per-operation AK/callApi constants alone do not establish ACS3.
- This issue also adds a source-aware signed-auth initializer guard: unsupported/dynamic explicit algorithms must remain discoverable with a reason and selected operations must fail before writes.
- Accepted current sources and anonymous handoff behavior are preserved; fixtures compare actual initializer/endpoint/handoff changes separately from prose coordinates.
- Verify official-parser negative discovery and generation rejection.

## Technical delivery; required gate open

- PR #65 merged at d9e56dd24a633ce1e2df241307a64dc25d6c0658.
- PASS exact-head Linux race/Windows/all CI at d83f911a01682fdd7427174452bda13d75ad16f2: https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37772643940/job/113295727091 https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37772643934/job/113295726772 https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37772643940/job/113295727353 https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37772643940/job/113295727245

- Independent docs-only developer acceptance remains NOT RUN; user arranging another Go developer.
- Fixed consumer-kit commit: d83f911a01682fdd7427174452bda13d75ad16f2. #60 stays open and #61 remains blocked.

## Status correction

- GitHub interpreted a negated closing phrase in PR #65 as an automatic closing keyword.
- Reopened: technical delivery is merged, independent docs-only developer acceptance remains NOT RUN.
- PR #65/#66 references are corrected; publication remains blocked.

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
