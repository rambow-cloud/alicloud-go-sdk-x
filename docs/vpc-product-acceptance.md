# VPC product acceptance

[中文](vpc-product-acceptance.zh-CN.md)

- Current generation follows [RPC expansion #83](dsl-rpc-expansion.md): ECS 380/380; VPC 396/403. Earlier counts and consumer records below describe their accepted revisions.

## Scope established before execution

- Issue #75; #60 and historical live-evidence integration #47 are complete. ECS #74 is reviewed separately.
- Keep pinned official Darabonba/parser -> complete IR -> shared backend/runtime. No source/policy change or token/waiter invention.
- Inventory: 403 discovered, 296 lowered/emitted actions, 1,242 models; 107 unsupported actions retain explicit reasons.
- Extend the isolated public examples/productacceptance module. Synthetic responses, small DescribeVpcsAPI mocks and application-owned OTel use no network/account.
- Reviewer: implementation agent. No independent human timing, performance, whole-cloud or broader Beta claim.

## Required matrix

| Case         | Required proof                                                                                                                                                   |
| ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| generation   | Reuse unchanged official frontend/57 tests and deterministic generators; compile complete supported inventory; unsupported selected behavior fails before writes |
| consumer     | Native DescribeVpcs page traversal and a small business mock; explicit Profile/cached generated STS shared with ECS                                              |
| pagination   | Multi-page/empty/short results, totals/bounds, stable failures/cancellation, model/options ownership, exact page fields and nested optional models               |
| retry-errors | Opt-in reviewed reads, conservative writes, signed endpoint/call overrides, Metadata/errors, in-flight deadline and secret-safe tracing                          |
| live         | Reuse #47 successful selected-field DescribeVpcs empty terminal read; sampled VPC contract/runtime/policy are unchanged                                          |
| docs         | Paired guide/report, runnable consumer/Examples, provenance/licenses, doccheck and explicit limits                                                               |

- Required live case: the accepted 2026-10-08 DescribeVpcs SDK/CLI read and native terminal-page metadata/ownership/cancellation checks.
- Nonempty live VPC continuation is optional/excluded for scoped v0.1.0: the authorized account sample contained no VPCs. Keep NOT RUN; offline continuation is mandatory. No VPC is created to obtain coverage.
- A waiter does not apply to the scoped DescribeVpcs traversal. Do not infer one from a Status field.
- The later requestless ListGeographicSubRegions action is not covered by #47. Add its independent offline smoke case; do not extend historical live coverage.
- Broader Beta live-continuation requirements and Explorer browser evidence remain separate. No new cloud calls, resource writes, source update or release.
- [Follow-up #79](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/79) tracks live instance tokens/waiter transitions and nonempty VPC continuation when suitable authorized existing resources are available. It is outside the scoped v0.1.0 gate.

## Verification and result

- Reuse accepted ECS-stage frontend/generator/root checks only where the source/runtime is unchanged. Run new isolated VPC tests/vet/run, formatting/language checks and final-head Linux race/Windows/automation.
- Refresh STS and ECS machine evidence against the final shared consumer/CI revision; changed workloads must not pass the release guard using stale pins.
- Record actual test events, versions, revisions, coverage and limits in acceptance/vpc-product-result.json.
- PASS: eight external consumer tests and ten subtests at 85795cf3cf1604afe59b0e8af03d5df0c9d3ba38; actual versions/timings are in [the machine record](acceptance/vpc-product-result.json).
- PASS: native traversal, mock business output, bounds/failure stability, fields/options, read retry/conservative write, Profile/shared STS cache, OTel, requestless smoke and cancellation/deadline.
- PASS: consumer vet/run, formatting/language/local links; unchanged frontend/generation/root evidence is reused from #74.
- STS (11 cases), ECS (10 cases/14 subtests) and VPC (eight cases/ten subtests) were rerun on the same clean shared workload revision. Required live evidence remains historical and scoped as declared.
- Final-head Linux race, Windows, automation and linked-issue checks are required before merge; exact CI/merge links are recorded on #75 and its PR.
- This completes scoped agent product acceptance only. #79 live gaps, optional human UX and #61 publication/indexing remain separate.

## Criterion mapping

- AC-01/02/03/07/08/10: public construction, native page models, business/mock traversal, optional fields, ownership, endpoints and errors.
- AC-04: not applicable to DescribeVpcs; no waiter added.
- AC-05/06/09: shared Profile/STS cache, reviewed retry and injected OTel.
- AC-11/12: package docs/Examples/licenses, paired consumer guidance, complete inventory and deterministic generation.
- UX-01/02/04/05: agent execution only. Publication/indexing belongs to #61.

- #81 service consolidation refresh: the eight consumer tests and ten subtests PASS at e33e5f93d856027569969056e9d4c25e92e51212. Machine evidence is updated; historical live samples do not establish nonempty continuation.

- #81 guide correction: discovery now precedes Go emission in the tool commands. All 29 consumer cases and 24 product subtests PASS at 29d468ad5f8999e92b4e60e4140cb8432efb8a13; machine records use this pin. Go/runtime/source/policy behavior is unchanged from the locally checked e33e5f93d856027569969056e9d4c25e92e51212 workload, so its unaffected gates are reused. Final-head CI remains required on PR #82.

- #83 shared lowering regression: 396 VPC actions compile; the existing 8 consumer cases and 10 subtests PASS at 01f25a9a571c3b59024dfbef7e2e96f7605f4a55. Seven unsupported actions retain explicit reasons. Historical live limits remain unchanged.

- #85 completion: VPC reaches 403/403 compiled actions and 1,728 models. Eight existing consumer cases/10 subtests PASS at 0fa3989f5df2e809082126eb9f177a8cc5ec218c. Root signed-wire tests cover all seven added actions; official simple/form helper comparisons pass separately. Historical live limits remain unchanged.
