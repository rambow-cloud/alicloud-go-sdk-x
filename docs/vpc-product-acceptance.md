# VPC product acceptance

[中文](vpc-product-acceptance.zh-CN.md)

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

## Verification and result

- Reuse accepted ECS-stage frontend/generator/root checks only where the source/runtime is unchanged. Run new isolated VPC tests/vet/run, formatting/language checks and final-head Linux race/Windows/automation.
- Refresh STS and ECS machine evidence against the final shared consumer/CI revision; changed workloads must not pass the release guard using stale pins.
- Record actual test events, versions, revisions, coverage and limits in acceptance/vpc-product-result.json.
- Current result: NOT RUN.

## Criterion mapping

- AC-01/02/03/07/08/10: public construction, native page models, business/mock traversal, optional fields, ownership, endpoints and errors.
- AC-04: not applicable to DescribeVpcs; no waiter added.
- AC-05/06/09: shared Profile/STS cache, reviewed retry and injected OTel.
- AC-11/12: package docs/Examples/licenses, paired consumer guidance, complete inventory and deterministic generation.
- UX-01/02/04/05: agent execution only. Publication/indexing belongs to #61.
