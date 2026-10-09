# [Maintenance]: Accept generated ECS product before v0.1.0 release

### Problem and route

- The user revised delivery on 2026-10-09: finish #60, then ECS/VPC acceptance, then #61 publication.
- Full-DSL service/ecs emits 283 actions. Generation and compilation are not product-wide live acceptance.
- Follow docs/sts-ecs-vpc-path.md, docs/product-acceptance.md and the official Darabonba/parser -> complete IR -> shared Go backend/runtime route.
- Member of #57; blocked by #60 and scoped live evidence #47. Parent membership is not a blocking dependency.

### Scope

- Pin complete ECS 2014-05-26 DSL/imports, parser and source-bound policy. Account for every operation/model and report distinct discovery/lowering/emission/compilation/test/live coverage.
- Accept consumer workloads for DescribeRegions, DescribeImages, DescribeInstances and DescribeInstanceStatus.
- Review existing page-number paginators, dual-mode DescribeInstances and InstanceRunningWaiter against exact IR/wire paths.
- Preserve shared credential/cache/Profile, context/options/mocks, endpoints, errors, middleware/OTel and reviewed retry behavior.
- Changes to policy/runtime/generator need evidence and a scoped implementation issue; no per-operation authoring prerequisite or generated-file edits.

### Acceptance

- [ ] Complete pinned RPC inventory regenerates deterministically, compiles and runs offline Examples. Unsupported shapes retain explicit reasons; no counts are promoted to all-action live coverage.
- [ ] A pinned isolated consumer implements image/instance traversal, mock business logic and waiter use with public APIs and no internal imports.
- [ ] Page/token mode exclusivity, optional fields, empty/short/repeated pages, cursor stability on failure/cancel and caller ownership pass independent fixtures.
- [ ] Wait and WaitForOutput cover all requested IDs, missing/partial/duplicate/unknown states, transitions, errors, deadlines and concurrent reuse.
- [ ] Reviewed read retry, non-retrying writes, endpoint overrides, structured errors/metadata and secret-safe middleware/OTel retain shared contracts.
- [ ] Review/integrate #47 evidence. Record selected-field live reads, actual continuation and authorized waiter cases separately; declare required live cases before execution. Empty terminal pages and SKIP do not prove continuation or state transitions.
- [ ] Paired acceptance report, consumer guide, coverage/limits and pkg.go.dev docs map to AC-01 through AC-12; no unmeasured superiority or broader Beta claim.

### Verification and limits

- Node 22 frontend checks/tests before both sdkgen checks, doccheck/vet/Go tests/Examples, isolated consumer, Linux race and Windows CI.
- Reuse accepted shared/STS/Profile evidence when unchanged. No cloud writes, IdP setup or benchmark execution is authorized by this planning issue.
- A required NOT RUN/FAIL remains blocking. Optional/excluded live cases must be declared before execution.
- Publication #61 is separate and waits for this issue and VPC acceptance.

### Execution matrix established 2026-10-09

- Follow docs/ecs-product-acceptance.md (paired Chinese guide). Implementation-agent consumer acceptance; no independent human claim.
- Required offline cases: full generation checks; public consumer/native Profile/cached generated STS; native page/token boundaries, failure stability and ownership; waiter transitions/failures/deadlines/concurrent reuse; read retry/write conservatism; endpoint/options/errors/Metadata; secret-safe OTel.
- Required live cases reuse accepted #47 selected-field reads and actual DescribeImages two-page continuation. Sampled ECS code/runtime/policy are unchanged.
- Live instance token continuation and waiter transitions are optional/excluded for this scoped v0.1.0: the account sample has no instances. Preserve their NOT RUN/SKIP status. These cases remain mandatory offline; do not create resources. This does not satisfy broader Beta live requirements.
- Run Node frontend gates before Go; both generator checks, doccheck/vet/root/isolated tests, formatting and final-head Linux race/Windows/automation. Record actual pins/test events in docs/acceptance/ecs-product-result.json.
- No new cloud calls, source updates, benchmarks or publication.
