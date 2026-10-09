# RPC DSL expansion

[中文](dsl-rpc-expansion.zh-CN.md)

## Problem and scope

- The pinned ECS DSL has 380 actions. Before #83, the backend emitted 283.
- First blockers: 80 duplicate query bindings, 7 deprecated input attributes, and 10 shrink-request signatures.
- Extend the shared official parser -> IR -> Go pipeline. Do not handwrite operation clients or change upstream source bytes.
- Apply the same rules to all pinned products. Keep capability policies sparse and retry opt-in.

## Implementation order

- Normalize identical guarded assignments to the same query key and source field. Reject conflicting sources, guards, aliases, and transforms.
- Accept boolean `deprecated` field attributes. Keep fields and emit `Deprecated:` Go comments with source attribution.
- Recognize the official temporary request -> validated shrink model -> conversion pattern.
- Represent reviewed JSON-string query transforms in IR with their original input field, exact wire key, and encoding.
- Upgrade product IR and its lock to schema v2. Reject incompatible versions before writes. Reports retain their independent v1 schema.
- Keep public inputs structured. Encode transformed fields as one JSON query value; ordinary arrays retain one-based query indexes.
- Preserve nil versus explicit empty values, cancellation, owned inputs, and middleware behavior.

## Acceptance and verification

- Parser tests use actual pinned ECS patterns and reject mutated conversions, styles, bindings, and attributes.
- IR and backend reject unknown encodings before writes. No operation-name special cases.
- Offline signed request tests check duplicate-key normalization, deprecated fields, JSON query payloads, absence/empty values, input ownership, and cancellation.
- Add executable examples and paired generated guidance. Run Node 22 frontend checks/tests before Go checks.
- Regenerate IR and Go; run product-check, doccheck, vet, Go tests, and formatting once after implementation. CI covers Linux race and Windows.
- Recount all products and record any subsequent blockers. Generated/compiled counts do not establish live cloud acceptance.
- No live writes, release tag, or pkg.go.dev publication in this change. Preserve historical acceptance and live pins; refresh affected offline consumer records at one shared workload commit.

## Result

- Issue: [#83](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/83).
- ECS: 380 discovered/lowered/emitted, 2,053 models; no remaining lowering blockers.
- VPC: 403 discovered, 396 lowered/emitted, 1,671 models. Seven actions remain unsupported: four unbound input flows, one query initialization flow, and two non-JSON shrink flows.
- STS: unchanged, four actions and 19 models.
- Node 22 frontend check and 73 tests, product-check, doccheck (15 packages), vet, root tests/Examples: PASS. Original source, sparse policies and STS Go output are unchanged.
- Shared consumer workload: `01f25a9a571c3b59024dfbef7e2e96f7605f4a55`.
- Isolated modules: both vet and account-free programs PASS. STS: 11 existing cases plus one pinned official JSON-helper comparison; ECS: 10 cases/14 subtests; VPC: 8 cases/10 subtests. Updated records share the workload pin and preserve historical live evidence.
- Review: index JSON encodings once and require IR schema v2. Reject old locks and mixed-version IR before writes. Typed nil JSON fields follow the upstream isUnset guard, while nested null and empty values remain intact.
- After these changes, Node checks/73 tests, product-check, doccheck, vet and root tests/Examples PASS. Generated Go and guides retain identical bytes.
- Tracked Go formatting PASS. Final-head CI is required before merge; inspect [PR #84 checks](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/84/checks) for current results.
- CI follow-up: [run 37913736842](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37913736842) hit Go's default 10-minute package deadline while the isolated compiler test was running. The earlier race suite passed in 335 seconds. Set an explicit 20-minute CI package deadline for the expanded inventory; retain race checks, every test and the compiler subprocess's 180-second deadline.
- These counts describe pinned generation, not complete live ECS/VPC acceptance.
