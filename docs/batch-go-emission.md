# Batch Go emission

[中文](batch-go-emission.zh-CN.md)

- Current generation follows [RPC expansion #83](dsl-rpc-expansion.md): ECS 380/380; VPC 396/403. Earlier counts and consumer records below describe their accepted revisions.

- Stage [#36](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/36) consumes the complete hash-pinned `models/*/ir.json`, following #35 / PR #40.
- This specification precedes implementation.
- Official DSL and its semantic IR are sufficient; no legacy metadata, per-operation field/model selection or documentation overlay is required.

- The pre-implementation specification is commit `ce18878`.
- Current emission is ECS 283/380 actions and 1,453 models, VPC 295/403 and 1,240 models, STS 1/4 and five models: 579 operations and 2,698 models in total.
- Each emitted operation has an external deterministic Example.
- Local full Go tests include independent temporary- module compilation and signed HTTP/middleware contracts; doccheck passed all 16 public packages.
- Linux race and Windows acceptance are recorded separately in the linked PR.
- Renderer reports intentionally leave compilation/live/policy unassessed rather than embedding machine-dependent successful-build claims into deterministic source output.

- From the repository root:

```sh
go run ./internal/cmd/sdkgen product-generate
go run ./internal/cmd/sdkgen product-check
go run ./internal/cmd/sdkgen product-check -operations ecs/DescribeImages,sts/AssumeRole
```

- Generate supported products into `service/<product>` (singular), following AWS-style service imports.
- #81 removes the earlier plural services/ packages and their Go emitter. Only service/ clients are supported; see [migration](service-consolidation.md).
- The old selective models are removed; callers must use complete native models.
- Migration is explicit: change the import, use optional scalar pointers, preserve native response containers, and pass DSL string fields such as JSON-encoded IDs as strings.
- No upstream source compatibility is claimed.
- New product capability adapters belong to #37.

- The backend emits every lowered operation, its complete reachable models, Input/Output, context-first method, service Options/NewFromConfig and small operation API interface.
- Input aliases represent request roots; Output represents the HTTP JSON response body plus runtime Metadata.
- Response-envelope models are also retained as DSL types; they are not the return envelope.
- Optional scalar/model fields use pointers; nil omits them, non-nil scalar pointers preserve explicit zero/false/empty.
- Arrays/maps retain their element types, numeric widths and exact wire case.
- Initialisms and anonymous names are deterministic; collisions fail rather than dropping fields.
- DSL optionality is not API requiredness.
- This stage introduces no guessed required-field constraints or retry safety.

- Use a private standard-library RPC helper for input snapshots and recursive query encoding.
- Snapshot all pointers, slices and maps before Initialize; hooks receive owned models.
- Query serialization preserves dotted member names and one-based array indexes, JSON string fields remain strings, and nil values are omitted.
- Context, signing, endpoint resolution, hooks, response limits, structured errors and retries remain the accepted shared runtime.
- All new operations conservatively disallow retry until reviewed #37 policy exists.
- Credentials and bodies are not logged.

- Encoding reference: the [official Go Query implementation](https://github.com/alibabacloud-go/openapi-util/blob/master/service/service.go) recursively flattens object members and uses one-based repeated indexes.
- It was inspected on 2026-10-07 as behavior evidence; it is not a pinned build/runtime dependency.
- The backend has independent standard-library code and offline contracts.

- `sdkgen product-generate` and read-only `product-check` operate offline.
- Explicit `-operations product/Name,...` asserts support without silently narrowing the generated product set.
- Unsupported or unknown selections, corrupted hashes, invalid shapes, protocols or naming collisions fail before writes.
- Preflight every owned output and reject symlinks/unmarked files; legacy and product generation have separate ownership.
- Publish a deterministic emission report with discovered/lowered/emitted counts and unsupported reasons.
- Compilation and live validation are separate recorded evidence, not inferred from successful rendering.

- Acceptance: deterministic regeneration; complete real-product model/operation counts; isolated temporary-module compilation; pre-write failures and stale-file reconciliation; signed offline HTTP tests for presence, repeated/nested query fields, case, full response containers, copying under middleware, call-option isolation, cancellation and errors; small-interface mocks; English Go docs and deterministic external operation Examples; corresponding Chinese guides; Go 1.27/JSON v2, frontend checks/tests, format/doccheck, both generation checks, vet and tests, Linux race and Windows CI.
- No live calls are required and no full ECS/VPC/STS coverage is claimed while unsupported operations remain.

- Pagination provenance: existing shared engine was committed directly on main as `61d2581` (issue #7); native generated adapters/options as `89e1d07` (issue #28), without a dedicated historical PR.
- New product-wide paginator/waiter policies and adapters will be reviewed in the PR for #37, after this emission PR.
- Policies name input/output cursors, collection paths, size/defaults and termination rules, rather than guessing capabilities from field or operation names.
