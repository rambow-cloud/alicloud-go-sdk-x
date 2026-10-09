# ECS product acceptance

[中文](ecs-product-acceptance.zh-CN.md)

- Current generation follows [RPC expansion #83](dsl-rpc-expansion.md): ECS 380/380; VPC 396/403. Earlier counts and consumer records below describe their accepted revisions.

## Scope established before execution

- Issue #74; prerequisite STS #60 is complete. Integrate historical live evidence #47 separately.
- Keep official pinned Darabonba/parser -> complete IR -> shared backend/runtime. No source or policy update is planned.
- Inventory: 380 discovered, 283 lowered/emitted RPC actions, 1,453 models. Unsupported actions keep explicit reasons. Compilation is not all-action behavioral coverage.
- Consumer module: examples/productacceptance; Go 1.27, JSON v2, public imports only, synthetic fixtures and no network.
- Reviewer: implementation agent. No human task-time, benchmark, broader Beta or SDK compatibility claim.

## Required matrix

| Case         | Required proof                                                                                                                                        |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| generation   | Official frontend check/tests; deterministic bridge/product checks; all generated packages compile; unsupported selected behavior fails before writes |
| consumer     | DescribeRegions, image traversal, instance token/page traversal, small mocks, native Profile and cached generated STS composition                     |
| pagination   | Empty/short/full pages, totals, repeated tokens, mixed modes, ownership, optional presence, malformed responses and stable retry/cancel cursors       |
| waiter       | Wait and WaitForOutput, all IDs, missing/partial/duplicate/unknown observations, transitions, errors, deadlines and concurrent reuse                  |
| retry-errors | Reviewed read retries, non-retrying writes, options/endpoints, structured errors/Metadata, cancellation and secret-safe injected OTel                 |
| live         | Existing #47 SDK/CLI selected-field reads and actual image page continuation; check sampled contracts are unchanged                                   |
| docs         | Runnable consumer/Examples, paired guidance, doccheck, provenance/licenses and explicit limits                                                        |

- Required live cases reuse the successful read-only DescribeRegions, DescribeImages two-page continuation, DescribeInstances token/page terminal reads and DescribeInstanceStatus terminal read from 2026-10-08.
- Live instance token continuation and waiter transitions are optional/excluded for this scoped v0.1.0 acceptance: the authorized account sample contained no instances. They remain NOT RUN/SKIP, not PASS.
- Token continuation and waiter transition/error contracts are mandatory offline cases. No instance is created or modified to obtain live coverage.
- This scope does not satisfy the broader Beta requirement for live token/waiter transitions in product-acceptance.md. Explorer browser verification remains NOT RUN.
- Reuse unchanged STS/Profile/source-update evidence. Publication and same-version pkg.go.dev indexing belong to #61.

## Verification and result

- Run Node 22 frontend check/tests before Go gates.
- Run both sdkgen checks, doccheck, vet, root tests/Examples, isolated consumer tests and formatting once for this change.
- CI runs Linux race, Windows and automation on the final PR head.
- Record actual revision, versions, test names and status in acceptance/ecs-product-result.json. Evidence-only follow-up does not change the tested workload.
- Initial #74 result: PASS, 10 external consumer tests and 14 subtests at ac2d9b6521af81a0fc628b7e3a0d209270420b3a; actual timings and Go/OS pins are in [the machine record](acceptance/ecs-product-result.json).
- PASS: Node 22 frontend check/57 tests, 25 automation tests, both generator checks, doccheck for 18 public packages, vet, root tests/Examples, consumer vet/run, formatting and paired language/local links.
- #47 evidence is integrated in PR #48. The sampled ECS implementation/runtime/policy is unchanged from its live baseline. Existing live status remains scoped as declared above.
- Final-head Linux race, Windows, automation and linked-issue checks must pass before merging the closing PR. CI links and merge revision are recorded on #74 and the PR.
- This is implementation-agent acceptance. Publication/indexing and broader live/human gaps remain separate.

## Criterion mapping

- AC-01/02/07/08: public construction/options, owned models, exact RPC fields, nested presence, endpoints and structured errors.
- AC-03/04/10: consumer traversal and independent page/token/waiter fixtures; broader live transition requirements remain excluded above.
- AC-05/06/09: native temporary Profile, cached generated STS provider, reviewed retry and injected secret-safe OTel.
- AC-11/12: generated package docs/Examples/licenses, paired consumer guide, complete inventory and deterministic/safe generation checks.
- UX-01/02/03/04/05: agent-executed consumer tasks only; no independent human usability or performance result.

- Shared-workload refresh for #75: the same ten tests and 14 subtests PASS at 85795cf3cf1604afe59b0e8af03d5df0c9d3ba38. The machine record now pins this revision; initial #74 evidence above is historical. Live gaps remain tracked by [#79](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/79).

- #81 service consolidation refresh: the ten consumer tests and 14 subtests PASS at e33e5f93d856027569969056e9d4c25e92e51212; root cross-capability contracts now use generated clients. Machine evidence is updated; live limitations remain unchanged.

- #81 guide correction: discovery now precedes Go emission in the tool commands. All 29 consumer cases and 24 product subtests PASS at 29d468ad5f8999e92b4e60e4140cb8432efb8a13; machine records use this pin. Go/runtime/source/policy behavior is unchanged from the locally checked e33e5f93d856027569969056e9d4c25e92e51212 workload, so its unaffected gates are reused. Final-head CI remains required on PR #82.

- #83 expansion/regression: 380 ECS actions compile; the existing 10 consumer cases and 14 subtests PASS at 01f25a9a571c3b59024dfbef7e2e96f7605f4a55. Additional root JSON/ownership/cancellation tests and Example PASS. Earlier live evidence stays pinned to its original revision; no all-action live claim.
