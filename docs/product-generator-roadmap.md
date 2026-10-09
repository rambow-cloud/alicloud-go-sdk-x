# Product generator roadmap

[中文](product-generator-roadmap.zh-CN.md)

- Current client paths follow [service consolidation #81](service-consolidation.md): the old services/ packages and Go emitter are removed; generate/check use complete products. Earlier bridge workflows below retain historical evidence only.

## Current delivery route (2026-10-09)

- Follow [STS/ECS/VPC delivery](sts-ecs-vpc-path.md): finish #60, then ECS #74 and VPC #75, then #61 release/indexing.
- #60 now requires truthful implementation-agent consumer acceptance. Independent human usability moves to optional follow-up #76; its record remains NOT RUN and does not block this release.
- Existing STS-only first-release scheduling and required-human #60 gates below are historical and superseded by this decision.
- Preserve actual technical/live/source evidence. Agent test duration is not human task time.
- Publication remains pending until both product acceptance issues pass. No tag is created by this change.

- First release scheduling follows [v0.1.0 STS](sts-v0.1.0.md), parent #57 and milestone v0.1.0 / Project 3.
- Extend the accepted pipeline with #58 requestless GetCallerIdentity and #59 Anonymous OIDC/SAML RPC, then #60 acceptance/source-update rehearsal and #61 release.
- Preserve the architecture/stage history below; this is a scoped STS release, not wider product/protocol acceptance.

- Accepted direction, 2026-10-07.
- This roadmap, development-path.md and AGENTS.md supersede conflicting older generation plans.
- Current #31 is a bounded compatibility bridge; product discovery/emission must not require per-operation handcrafted field/model/doc selection.
- Historical acceptance records remain accurate, separate from future goals.

- The source pipeline is complete official DSL -> official semantic parser -> normalized operation/model/binding IR -> our Go backend -> accepted runtime.
- Optional pinned canonical metadata enriches API requiredness, parameter styles, documentation and cross-checks.
- It cannot overwrite exact DSL wire names or dictate CLI response shapes.
- Preserve provenance for every fact; missing enrichment is recorded rather than invented.

- Sources: [SDK DSL](https://github.com/aliyun/alibabacloud-sdk) and [CLI metadata](https://github.com/aliyun/aliyun-openapi-meta).
- The latter declares Apache-2.0 and warns its structure is unstable and currently intended for CLI builds.
- Pin commits and hashes and use versioned adapters.
- Licensed descriptions may seed English Go comments and equivalent Chinese guides with attribution/license preserved.
- Web metadata and imported modules have separate provenance; do not label the entire corpus MIT.

| Stage                | Delivery                                                                                                                       | Acceptance                                                                                                         |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------ |
| 1: normalization     | Exact wire names/case; indexed request models; itemName response wrappers; raw provenance and CLI-only attributes              | Filter model and Filter.1.Key/Value correspond; true name/type/requiredness changes remain visible                 |
| 2: discovery/IR      | Every product operation and reachable model; protocol/bindings/source locations; versioned deterministic IR and coverage       | No per-operation metadata/overlay dependency; discovered/lowered/unsupported counts and reasons                    |
| 3: batch emission    | Complete supported inputs/outputs/models/methods, small interfaces, generic codecs/naming with sparse compatibility exceptions | Compiles; signed-wire/copy/presence tests; deterministic output; selected unsupported behavior fails before writes |
| 4: capability policy | Native token/page pagination, waiter acceptors, idempotency/client tokens, sensitive fields and special validators             | Shared engines and AWS conventions; conservative defaults, no invented pagination or guessed write retry           |
| 5: docs/expansion    | Licensed bilingual doc automation, English Go comments, corresponding guides and offline Examples; further profiles/products   | pkg.go.dev, Linux race/Windows, honest coverage and separate live evidence                                         |

- Stage 1 retains raw_name separately from CLI name/options, case-sensitive Key/key and repeatList element fields.
- Indexes are declared positions, not inferred server limits.
- Response array plus itemName is a representation requiring an object/array wrapper; backendName, nullToEmpty and valueMapping remain source attributes, not automatic SDK transformations.
- Only classify material source conflicts after normalization.

- Stage 2 scans complete pinned DSL independently of legacy selected snapshots/overlays.
- Lowered does not mean generated, compiled or live accepted.
- Enumerate unsupported ROA, body/stream/helper patterns with source locations and reasons.
- An explicit supported subset is reviewable; never silently omit APIs while claiming a complete SDK.

- Stage 3 preserves context, service Options/NewFromConfig, functional call options, wire containers, absence, input copying and cancellation/error behavior.
- RPC comes first.
- Stage 4 overlays describe exceptions/policies rather than every wire field; do not infer retry safety just from names, CLI operation_type or HTTP verbs.

- Every stage includes its own docs/tests/Examples and offline regeneration; stage 5 expands automation rather than postponing docs.
- Do not turn account/resource CLI examples into generated Go tests.
- Browser checks are performed by the user at exact documented links; authorized read-only CLI/HTTP evidence is distinct from offline and UI evidence.

- First milestone: normalize real representations and publish complete pinned ECS inventory/coverage, then batch generate supported RPC operations.
- Track stages under an open roadmap parent with real child issues created before code and dependencies 1 -> 2 -> 3 -> 4 -> 5.
- Record returned issue numbers in docs/issues/README.md.
- Created tracking: parent [#33](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/33), stages #34 (normalization) -> #35 (discovery/IR) -> #36 (emission) -> #37 (policy) -> #38 (docs).
- All five stages and review fix #44 are integrated into main as of 2026-10-08; [review/integration evidence](generator-integration.md) records the accepted pinned RPC scope, merge commits and remaining coverage limits.
- Keep #31 focused on the compatibility bridge and benchmarks separate.
- Each issue uses its own branch; stacked PRs name dependencies and do not imply main contains unmerged work.
- Do not close the parent while batch emission, capability policies or docs remain unfinished.

- #58/PR #63 and #59/PR #64 are merged: all four pinned STS actions emit with native signed/anonymous separation. #60 delivers consumer/real-source/live-identity evidence and the independent developer handoff; keep its UX gate open until the user-arranged Go developer records actual results. #61 publication/indexing follows that required acceptance.
- Historical planning counts below describe the earlier baseline, not current coverage.
- See [acceptance evidence](sts-v010-acceptance-report.md).
