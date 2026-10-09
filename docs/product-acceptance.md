# Product acceptance

[中文](product-acceptance.zh-CN.md)

## Scoped ECS and VPC acceptance

- #74/#75 use the pre-execution matrices in [ECS acceptance](ecs-product-acceptance.md) and [VPC acceptance](vpc-product-acceptance.md).
- Supported RPC inventory is 283 ECS and 296 VPC actions. Offline consumer contracts and selected-field live reads are separate; unsupported DSL actions retain reasons.
- Instance token/waiter live transitions and nonempty VPC live continuation remain excluded/NOT RUN or SKIP under follow-up #79. No broader Beta or full-cloud acceptance is claimed.
- Consumer acceptance is implementation-agent execution. Independent human UX and automated test timings remain distinct.
- Preserve STS/ECS/VPC evidence pins against the shared consumer/CI revision. Publication and same-version pkg.go.dev indexing remain #61. No tag is created during product closeout.

## Current delivery route (2026-10-09)

- Follow [STS/ECS/VPC delivery](sts-ecs-vpc-path.md): finish #60, then ECS #74 and VPC #75, then #61 release/indexing.
- #60 now requires truthful implementation-agent consumer acceptance. Independent human usability moves to optional follow-up #76; its record remains NOT RUN and does not block this release.
- Existing STS-only first-release scheduling and required-human #60 gates below are historical and superseded by this decision.
- Preserve actual technical/live/source evidence. Agent test duration is not human task time.
- Publication remains pending until both product acceptance issues pass. No tag is created by this change.

### Objective and authority

- Acceptance establishes whether developers can replace the official Alibaba Cloud Go SDK v2 for an explicit workload with consistent Go APIs, correct behavior and manageable maintenance.
- The criteria below govern future scoped issues alongside [development-path.md](development-path.md).
- Code generation, compilation, tool selection, stars and contributor counts do not establish product acceptance.

- Keep complete pinned official Darabonba DSL -> official semantic parser -> normalized IR -> our Go backend -> shared runtime.
- Smithy remains an isolated local experiment unless a separate issue/decision changes the route.
- Require Go 1.27+, direct JSON v2, standard-library core imports and optional OTel.
- Issues, PRs and Go comments use English.
- Keep English and Chinese guides in separate, linked files.

### Candidate Beta boundary

- The first version is now the user-selected [v0.1.0 STS](sts-v0.1.0.md), tracked by #57-#61, milestone v0.1.0 and Project 3.
- Its scoped experimental release gate requires all four pinned STS actions generated/compiled/offline-tested, credential composition, scoped live evidence, consumer/maintenance/docs and publication/indexing.
- Successful OIDC/SAML live federation is declared outside that required live scope before execution, never promoted from NOT RUN to PASS.
- This explicit route revision supersedes the broader Beta release prerequisite for v0.1.0 only; the candidate boundary/AC/UX definitions below remain the broader goal.
- Neither release claims broader Beta or unverified ECS/VPC coverage.
- The later user-approved #68 correction adds required native default configuration/Profile/OAuth behavior before publication; the earlier exclusion is superseded by [the new route](default-configuration.md).

- The candidate scope below is the starting acceptance workload, not a release claim.
- Adding/removing required cases needs an issue, evidence and paired updates.
- Passing shared contracts does not extend product coverage automatically.

| Boundary          | Candidate requirement                                                                                                                                  |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Products/packages | ECS 2014-05-26, VPC 2016-04-28, STS 2015-04-01 under `service/`                                                                                        |
| Actions           | ECS DescribeRegions/DescribeImages/DescribeInstances/DescribeInstanceStatus; VPC DescribeVpcs; STS AssumeRole                                          |
| Protocol/region   | Reviewed RPC/ACS3 over HTTPS; cn-hangzhou; other modeled regions retain separate evidence                                                              |
| Capabilities      | Four native paginators; ECS InstanceRunningWaiter; reviewed opt-in read retry; STS provider/cache composition; optional OTel                           |
| Identity          | Explicit static/env providers, explicit chain/cache and AssumeRole using separate source credentials                                                   |
| Exclusions        | Automatic process/metadata discovery, arbitrary write retry, ROA/OSS/streaming, whole-cloud parity; native CLI Profile/OAuth is now required under #68 |

- Historical local-profile validation injected a checked snapshot after CLI authentication; it does not prove native loading or renewal. #68 adds native loading and scoped live exchange/persistence evidence under [the new contract](default-configuration.md).
- No cloud resource creation, write operation or target-role use is authorized by this document.

- Application guidance prioritizes renewable STS role providers/cache.
- Long-lived keys require explicit StaticProvider/EnvProvider/profile-provider registration; complete temporary environment credentials and native profiles are discoverable through LoadDefaultConfig.
- Custom providers remain supported.
- Config/Options have no bare credential fields, construction never makes credential HTTP calls, and nil/typed-nil sources fail before requests.
- Present invalid sources stop resolution.
- Historical [#53](credentials.md) rejection/signing/cache evidence remains under AC-01/05/11 and UX-01/04; #68 supersedes its blanket discovery exclusion.

### Required criteria

| ID    | Contract                                                                                                                                                              | Required proof                                                                                                                                                        |
| ----- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| AC-01 | Consistent typed Options/NewFromConfig, context-first operations and call options; documented absence/defaults and ownership                                          | External consumer examples; input/config isolation; cancellation/deadline identity; Windows and Linux race CI                                                         |
| AC-02 | Exact Alibaba action/version/wire names, complete supported models, absence/zero/false/int64 and response containers; JSON v2 strictness with unknown-field tolerance | Independent wire fixtures, signed requests, malformed JSON cases and selected CLI/Explorer comparison; report all-field vs selected-field evidence                    |
| AC-03 | Native HasMorePages/NextPage; no invented token; reject mixed modes; stable cursor on errors/cancel; bounded cycle handling                                           | Multi-page token/page, empty page, repeated token, metadata bounds and ownership cases; live continuation evidence for both candidate modes                           |
| AC-04 | Reusable Wait/WaitForOutput; all requested IDs required for success; bounded/context-aware polling and isolated concurrent waits                                      | Missing/partial/duplicate/unknown states, transitions, errors and expiry; authorized live waiter success and transition evidence                                      |
| AC-05 | Explicit credential precedence, missing vs invalid distinction, shared bounded refresh, valid role keys/token/expiration and separate source identity                 | Full-DSL STS client -> provider -> cache -> generated consumer integration; rotation/concurrency/cancel/invalid-response tests; authorized role/refresh live evidence |
| AC-06 | Opt-in retries only for reviewed replayable/idempotent operations; bounded attempts/time, jitter/Retry-After and fresh per-attempt signing                            | Offline fault injection, canceled backoff, budget and unsafe-write rejection; live fault injection is separate and never implied                                      |
| AC-07 | Deterministic product/region endpoints and explicit override; unsupported combinations fail                                                                           | Rule/HTTPS/signing tests and candidate-region live reads; custom and other-region evidence kept separate                                                              |
| AC-08 | Ordered middleware lifecycle and optional injected OTel; operation/attempt hierarchy, propagation and no secret/raw-body labels                                       | Hook lifecycle and test-exporter integration; no global provider mutation; private error/body/query exclusion                                                         |
| AC-09 | errors.Is/As, typed service/operation errors, request ID/status/attempt metadata and safe default formatting                                                          | Error chains, cancellation, status/decode failures and sensitive fmt/log/span tests; no string matching required in examples                                          |
| AC-10 | Small operation/paginator/waiter mock seams, scripted HTTP and virtual time                                                                                           | Real consumer-style tests with no network/account; test business results and failure handling                                                                         |
| AC-11 | English Go docs and runnable offline Examples, equivalent Chinese guidance, defaults/limits/migration/license notices                                                 | doccheck/Examples; source-prose gaps recorded; tagged release and user browser inspection of the same version on pkg.go.dev                                           |
| AC-12 | Complete discovery, deterministic regeneration, explicit unsupported reasons, reviewed sparse policy and safe source updates                                          | Frontend/check/product-check, invalid selected cases leave outputs untouched, compatible/incompatible drift report and one real upstream revision rehearsal           |

- For AC-11, local docs/Examples/licenses are Beta requirements; tagging and same-version pkg.go.dev inspection belong to the release gate.
- They are not prerequisites for Beta.

- AC-03/04/05 use product-specific reviewed rules, not field-name inference.
- The basic Smithy allStringEquals waiter and stock paginator ownership observed in the PoC do not satisfy these contracts automatically.
- Our policy semantics remain authoritative.

### Developer experience tasks and comparison

| ID    | Consumer task                                                        | Acceptance observation                                                                                                                            |
| ----- | -------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| UX-01 | New module -> explicit credentials -> first ECS read using only docs | Initial target <=15 minutes with Go installed and credentials already authorized; exclude browser/approval wait; record actual time and obstacles |
| UX-02 | Traverse ECS token plus image/VPC page results                       | Zero handwritten cursor advancement/termination logic; caller may use the documented paginator loop                                               |
| UX-03 | Wait for multiple instances                                          | Zero handwritten polling/backoff loop; handle missing IDs, timeout and structured failure                                                         |
| UX-04 | Use renewable AssumeRole credentials                                 | Source STS client, provider, cache and generated consumer compose without a handwritten credential translator or refresh loop                     |
| UX-05 | Test the application and add observability                           | Fake only needed operations; no cloud access in tests; injected tracing and errors.Is/As without SDK internals                                    |

- Use Go developers who did not implement the feature.
- Pin this SDK and official v2 module versions, Go/OS/architecture, API inputs, result projection and workload.
- Record task success, time, application-code lines (excluding generated SDK/vendor/setup), documentation lookups, manual glue and obstacles.
- Show both solutions; do not infer superiority from lines alone.
- No independent-user result has yet been recorded.

- [Benchmark #20](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/20) separately tracks imports/build/binary costs; encoding/decoding latency and allocations require equal fixtures and published methodology.
- Set regression thresholds after obtaining a reproducible baseline, before optimization.
- No speed or percentage claim is accepted without measured evidence.
- Contributor docs, issue/label hygiene, reproduction guidance and update workflow are maintainer acceptance; community popularity is not a release gate.

### Evidence, gates and current gaps

- Every case records criterion/task ID, required scope, SDK/source/official-tool versions, Go/OS/architecture, commit, check command or browser action, fixture origin, result, redacted evidence link, limitations and responsible issue.
- Separate discovered, lowered, emitted, compiled, offline-tested, live-tested and published/indexed evidence.

- Results are PASS, FAIL, SKIP (with reason) or NOT RUN.
- SKIP/NOT RUN is not PASS.
- A skipped required live/UX/publication case keeps its gate open; do not relabel it as optional after execution.
- Expected-negative tests pass only when the required rejection is observed; an audit recording a weakness does not approve that weakness.

| Gate         | Required completion                                                                                                                          |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Foundation   | Eleven shared capabilities, meaningful offline contracts/Examples/doccheck, standard-library core, Linux race and Windows CI                 |
| Product Beta | Candidate AC contracts, all five consumer tasks, scoped live cases and real source-update rehearsal; zero unresolved required failures/skips |
| Release      | Product Beta plus immutable version/release notes, migration/compatibility statement, licenses and same-version pkg.go.dev browser evidence  |

- Historical baseline: [foundation mapping](foundation-acceptance.md) and [integration record](generator-integration.md).
- The latter records 579 emitted actions, four paginators, one waiter and seven reviewed policies; 572 actions remain unreviewed for those capabilities.
- These are recorded snapshots, not a Beta completion count. [Live #47 / PR #48](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/48) records two image pages; instance token/page/status and VPC traversals end on an empty first page, waiter/AssumeRole are skipped and Explorer browser verification is not run.

- STS composition #51 and explicit-provider guidance #53 are merged. [Live STS #55](live-sts-renewal.md) passes scoped issuance/reuse, forced refresh and automatic real-time expiry renewal with generated ECS reads and temporary-IAM cleanup.
- Background/concurrent live refresh, other roles/conditions and native Profile/OAuth renewal are not established by this run.

- Current work queue, in priority order: reproducible consumer workloads and independent UX/official-v2 comparison; authorized token/waiter/ remaining scoped role live gaps; real upstream-update rehearsal; release/indexing.
- Benchmark #20 stays independent.
- The older STS helper only accepts `services/sts`; its foundation acceptance does not prove composition with the new `service/sts` API. #51 adds that composition with local doccheck, product-check, vet, all Go tests/Examples and 50 frontend tests passing. #51 is offline evidence; #55 adds scoped live renewal.
- Independent UX remains open.

- Closing an implementation or this definition issue does not close these overall gates.
- Review all tracked cases at the target release commit.
- This change creates no version tag or switch the production generator; this acceptance definition itself authorizes no live calls.
- The separately user-authorized #55 scope is recorded above.
