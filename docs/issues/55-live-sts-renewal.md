# Live STS renewal acceptance

## Problem and evidence

- #51/#53 provide offline full-DSL STS provider/cache integration, but AC-05 live issuance and renewal are unverified.
- The user now authorizes local unsandboxed verification using the existing Aliyun Profile.
- Current default oss-sftp uses OAuth and has no configured target role; the user now authorizes creation of a dedicated validation role.
- CLI preflight PASS confirms an Account identity, which cannot AssumeRole directly.

## Scope and dependencies

- Depends on merged #49/#51/#53.
- Plan: docs/live-sts-renewal.md, baseline f4cebddd3660d0d09d2b56e6319f5895c2f4d4a5.
- Preflight one bounded CLI GetCallerIdentity and use that account Profile only for setup/cleanup.
- Inject a separate explicit RAM-user source into generated STS.
- Use the dedicated authorized role with DurationSeconds=900, real server time, existing role provider/cache and generated consumer reads.
- Provision only a unique RAM user without console login, one temporary AccessKey kept in memory, a policy permitting sts:AssumeRole on one new role, and a role trusting that user with only ecs:DescribeRegions.
- Clean up all newly owned resources/attachments after the run.
- Do not modify existing identities/permissions/resources.
- No implicit SDK Profile/OAuth feature.

## Acceptance criteria

- [ ] Verify the selected source identity/authentication and sufficient source lifetime; stop safely on missing role or invalid login.
- [ ] Perform real AssumeRole and validate nonblank keys/token/future expiry, role identity and structured HTTP metadata without exposing credentials or account/resource identifiers.
- [ ] Prove cache reuse and a separately labeled forced Invalidate refresh with actual issuance counts and in-memory change comparisons.
- [ ] Wait for genuine server-side expiration, then prove automatic cache renewal and successful role-authenticated generated-client reads with no virtual time/manual invalidation.
- [ ] Publish paired sanitized evidence with version/commit/timing/status/limits. Authorization failure, missing prerequisites or skipped required cases do not establish PASS or Beta.

## Verification

- Keep disposable Go 1.27/JSON-v2 harness and reports under ignored .git/.
- Bound network requests/context and CLI retries; progress waits at most 60 seconds.
- Use real clock, explicit source provider, generated service/sts, stscreds and Cache, with the generated AssumeRole response identity plus generated ECS DescribeRegions (GetCallerIdentity is CLI preflight only; it is not emitted by the pinned STS DSL); do not broaden permissions on failure.
- Record STS versus ECS authorization separately.
- Existing offline CI remains the baseline; no production code change is planned.
