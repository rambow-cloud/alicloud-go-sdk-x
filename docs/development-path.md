# Development path

[中文](development-path.zh-CN.md)

- #92 next delivers [OSS shared runtime](oss-runtime.md): OSS4, bucket hosts, XML request MD5, typed XML responses and structured errors. Production source/IR/emission follows; ListBuckets remains an explicit conflict.

- #92 next uses [native XML root discovery](native-xml-traits.md): complete hash-bound helper facts complement official DSL, with explicit conflict reports before production XML/Gateway integration. No inferred roots or generated OSS coverage claim.

- The [shared internal XML codec](xml-model-codec.md) is merged under #92. Root/model fixtures precede OSS Gateway/signing/IR integration; no public OSS client is accepted yet.

- #92 follows [FC binary generation](fc-binary-generation.md): schema v6, openapi-http-v1 and 73 generated FC actions. The [JSON/none baseline](fc-roa-product.md) remains historical evidence. XML, unbounded request streams and live acceptance remain separate.
- Before binary emission, implement [bounded response streaming](response-streaming.md) in the shared runtime. Its reader lifetime and publication gates do not establish generated binary coverage.

- Current: #85 is merged. Follow [gap completion #87](gap-completion.md) for the next implementation stages.

## One service path

- #92 checksum arithmetic follows [OSS checksums](oss-checksums.md). Fixed byte vectors and redacted verification precede operation/stream integrity policy; no automatic checksum behavior is enabled by this foundation.

- Follow [service consolidation #81](service-consolidation.md) before #61: remove the bridge, migrate foundation/provider contracts to full-DSL clients and refresh consumer evidence. This overrides earlier bridge preservation requirements.

## Scoped ECS and VPC acceptance

- #74/#75 use the pre-execution matrices in [ECS acceptance](ecs-product-acceptance.md) and [VPC acceptance](vpc-product-acceptance.md).
- Generated RPC inventory after #85 is 380 ECS and 403 VPC actions. #74/#75 consumer acceptance remains pinned historical evidence; expanded generation does not imply all-action live acceptance. Offline consumer contracts and selected-field live reads are separate; unsupported DSL actions retain reasons.
- Instance token/waiter live transitions and nonempty VPC live continuation remain excluded/NOT RUN or SKIP under follow-up #79. No broader Beta or full-cloud acceptance is claimed.
- [Current existing-resource follow-up](live-resource-followup.md) adds nonempty VPC and current-state waiter evidence. Multi-page and transition cases remain NOT RUN.
- Consumer acceptance is implementation-agent execution. Independent human UX and automated test timings remain distinct.
- Preserve STS/ECS/VPC evidence pins against the shared consumer/CI revision. Publication and same-version pkg.go.dev indexing remain #61. No tag is created during product closeout.

- [Full-DSL live evidence](product-live-validation.md): Historical full-DSL ECS/VPC read-only evidence (#47): two image pages; empty instance/VPC pages; live waiter skipped. Product acceptance #74/#75 remains separate.

## Current delivery route (2026-10-09)

- Follow [STS/ECS/VPC delivery](sts-ecs-vpc-path.md): finish #60, then ECS #74 and VPC #75, then #61 release/indexing.
- #60 now requires truthful implementation-agent consumer acceptance. Independent human usability moves to optional follow-up #76; its record remains NOT RUN and does not block this release.
- Existing STS-only first-release scheduling and required-human #60 gates below are historical and superseded by this decision.
- Preserve actual technical/live/source evidence. Agent test duration is not human task time.
- Publication remains pending until both product acceptance issues pass. No tag is created by this change.

## Current status

- #58 and #59: merged.
- The generator supports all four scoped STS actions.
- #68: merged.
- Default configuration and native CLI Profile/OAuth have scoped live evidence.
- #60: implementation-agent acceptance is delivered; independent human usability remains #76 / NOT RUN.
- #61: immutable v0.1.0 publication and all seven same-version pkg.go.dev browser checks passed. [Publication record](releases/v0.1.0-publication.md) distinguishes agent release checks from the user's browser report; broader gaps remain #87.
- #70: split language files, improve Chinese, and use English-only issues.
- This language policy changes documentation format, not SDK or release acceptance.
- #72: [STS reuse review](sts-reuse-review.md). Share provider validation, isolate call options, and prove renamed DSL actions use the same parser/backend before release.

## Previous decisions and acceptance history

- The user's latest 2026-10-08 correction requires [default configuration and native CLI Profile/OAuth #68](default-configuration.md) before v0.1.0 publication.
- Implement the documented loader/profile/cache/refresh contract and evidence, then update #60's consumer handoff before #61.
- This supersedes older blanket no-discovery and deferred native Profile/OAuth scope below.
- Explicit long-lived key opt-in remains required; direct service constructors keep provider-only validation.
- Historical #53/#55 evidence does not prove native OAuth renewal.

- The user-approved first release is now [v0.1.0 STS](sts-v0.1.0.md): milestone v0.1.0, Project 3 and parent #57.
- Next implement #58 requestless GetCallerIdentity, then #59 Anonymous OIDC/SAML RPC; both unblock #60 four-action acceptance/consumer/ source-update rehearsal, followed by #61 release/indexing.
- This scoped experimental schedule supersedes the broader candidate Beta queue below, preserving its criteria and history without claiming that Beta.
- Existing #51/#53/#55 are accepted evidence; no new cloud objects or tag are created by this planning change.

- Future scoped issues follow [product acceptance](product-acceptance.md) (#49): candidate Beta boundaries, AC/UX IDs, distinct offline/live/UX/publication evidence and required-case status govern completion. #51/#53 STS composition/provider guidance are merged; [#55](live-sts-renewal.md) records scoped real-time renewal.
- Next: consumer/official-v2 comparison, remaining authorized live gaps, source-update rehearsal and release.
- Benchmark #20 remains separate; Smithy is an isolated experiment.

- The first follow-up is [full-DSL STS provider composition #51](sts-credentials.md) for AC-05/UX-04; it preserves the reference helper and establishes offline integration, while real role/refresh and independent developer-task evidence remain separate.

- The next user-directed credential step [#53](credentials.md) makes STS provider/cache the primary documented application path.
- Long-lived AK/SK and environment sources require explicit provider registration, following AWS's provider injection convention; Config/Options must never gain bare key fields or implicit environment fallback.
- Preserve custom providers and the existing StaticProvider API.
- Reject nil and typed-nil sources before requests, prove the rule through runtime/generated-client and per-call tests, and ship a complete offline STS-to-ECS Example with paired guidance before resuming UX comparisons.

- The user now authorizes local unsandboxed [live STS renewal](live-sts-renewal.md).
- Establish its issue, provision only the user-authorized dedicated role/source with least-privilege policies and cleanup, then validate forced refresh separately from real-time natural expiry renewal before marking AC-05 live evidence.

- #55 completed three live issuances and four generated ECS reads, including cache reuse, forced refresh and automatic renewal after genuine 900-second expiration.
- All newly created temporary IAM objects were cleaned up.
- This is scoped AC-05 evidence; independent UX and other product/live requirements remain separate.

- Authoritative route revised at the user's direction on 2026-10-07: complete official product DSL -> official Darabonba semantic parser -> normalized operation/model/binding IR -> our Go backend -> existing runtime. [product-generator-roadmap.md](product-generator-roadmap.md) and AGENTS.md govern this route and supersede conflicting older metadata-first, per-operation snapshot/overlay, field-selection and handwritten-doc prerequisites.
- The older stages and issue references below describe historical acceptance only.

- The five-stage implementation and review fix [#44](capability-role-review.md) were integrated into main on 2026-10-08; [review evidence](generator-integration.md) records the accepted scope and merge commits.
- Same-model cursor/waiter aliases fail before writes.
- Further capability or protocol expansion needs separately scoped issues; the accepted pinned RPC implementation is the starting point.

- Execute source normalization -> complete operation/model discovery and IR -> batch Go emission -> sparse capability policy -> documentation automation/profile expansion.
- Every stage includes paired docs, relevant tests, Examples and pkg.go.dev acceptance; documentation automation at the last stage does not postpone earlier documentation.
- Optional pinned canonical metadata enriches/cross-checks DSL after representation normalization.
- New discovery cannot require metadata/overlay entries for each API.

- First milestone: normalize real source representations and publish complete pinned ECS inventory/coverage; then batch generate supported RPC operations. #31/PR #32 are the five-operation compatibility bridge, not a product-wide generator acceptance gate.
- Overlay supplies reviewed compatibility exceptions, pagination/waiters, idempotency, sensitive fields and special validators, rather than re-declaring each model/field.
- Unsupported operations have explicit reasons; selected unsupported behavior fails before writes.
- Normalize Filter indexed bindings and itemName response wrappers before classifying conflicts.
- Browser, CLI-local validation and actual HTTP evidence are distinct.

- Document -> establish/update actual issue dependencies -> one reviewable branch per issue -> implement, validate and record -> linked PR.
- Keep the roadmap parent open until all stages meet acceptance, preserve accepted runtime/AWS calling conventions, Go 1.27/JSON v2 and standard-library core.
- Benchmarks remain separate.

- The goal is a unified Go runtime, validated by a small handwritten ECS/STS reference, followed by product generation.
- Go 1.27 and encoding/json/v2 are required.
- Core runtime imports remain standard-library-only; OpenTelemetry is an optional integration.

| Stage                      | Deliverables                                                                                           | Gate                                                |
| -------------------------- | ------------------------------------------------------------------------------------------------------ | --------------------------------------------------- |
| A: specification           | bilingual docs, English issues, acyclic dependencies, language policy                                  | documented before code                              |
| B: request foundation      | staged middleware, endpoint resolver, ACS3 signing, HTTP execution, structured errors, testing helpers | offline request/response and concurrency contracts  |
| C: resilience and identity | explicit credential chain, expiry-aware cache, bounded retries, STS AssumeRole helper                  | cancellation, freshness, replayability, idempotency |
| D: reference behavior      | handwritten ECS reads, unified pagination, bounded waiters, optional OTel                              | typed usage and cross-capability integration        |
| E: foundation acceptance   | eleven capabilities, runnable examples, paired guides, Linux race and Windows CI                       | implementations and tests, not interface presence   |
| F: generation              | frozen metadata to typed clients/codecs/docs/policies                                                  | #19 passed; follow #21 -> #22 -> #23 under #8       |

- Dependency order: testing/middleware/endpoints/errors and signing -> HTTP runtime -> ECS/STS clients -> paginator/waiter/AssumeRole integration.
- Credential chain/cache and retry policy can be implemented independently before runtime integration.
- OTel depends on middleware and request metadata.
- Signing's standalone acceptance must not depend on HTTP integration; integration belongs to the HTTP issue, removing the previous cycle.

- All eleven requested areas are mandatory: paginator, waiter, retry/backoff, mock interfaces, credential provider, STS helper, endpoint resolver, middleware, OTel, structured errors, testing helpers.
- Public extension contracts are shared; product-specific rules remain typed.

- Initial protocol is ACS3-HMAC-SHA256 for Alibaba Cloud OpenAPI.
- Initial reference coverage: ECS DescribeRegions, DescribeInstances (selected response fields), DescribeInstanceStatus; STS AssumeRole.
- This is not full ECS/STS, OSS/SLS data-plane, or live-cloud acceptance.
- Use explicit endpoint rules; unsupported regions fail rather than guessing domains.

- Preserve the initial no-retry default.
- Standard retry is opt-in, bounded, cancellation-aware, with jitter and Retry-After, and applies only to explicitly idempotent/replayable operations.
- Waiter total expiry is distinct from per-request retry.
- Credential refresh is bounded and shared without allowing one canceled waiter to cancel other callers.
- No implicit process or metadata credentials.

- Every issue includes package docs, deterministic external Examples and paired behavior guides.
- GitHub is the source of execution status. #1/#2 are historical bootstrap completion; #3-#7/#9 are refined foundation tasks; #8 is generation only.
- Benchmark work is separate.
- New foundation issues are listed in the issue index and GitHub milestone; record actual numbers, never assume them.

- Generator architecture, supported profile, provenance and acceptance are defined in [generator.md](generator.md).
- The foundation gate passed on commit 28684e4 before #21.
- The next real-product gate is #24 -> #25; see [generator-expansion.md](generator-expansion.md).
- After that acceptance, resolve the review under #26 -> #27 -> #28 -> #29 before wider profiles: [AWS-style remediation](aws-style-remediation.md).

- The frontend refactor #31 uses official product DSL and the official parser; its ordered delivery and conflict verification are defined in [darabonba-migration.md](darabonba-migration.md).

- Sources: [ACS3](https://help.aliyun.com/zh/sdk/product-overview/v3-request-structure-and-signature), [metadata](https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/), [AWS middleware](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/middleware.html), [AWS testing](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/unit-testing.html).

- #58/PR #63 and #59/PR #64 are merged: all four pinned STS actions emit with native signed/anonymous separation. #60 delivers consumer/real-source/live-identity evidence and the independent developer handoff; keep its UX gate open until the user-arranged Go developer records actual results. #61 publication/indexing follows that required acceptance.
- Historical planning counts below describe the earlier baseline, not current coverage.
- See [acceptance evidence](sts-v010-acceptance-report.md).
