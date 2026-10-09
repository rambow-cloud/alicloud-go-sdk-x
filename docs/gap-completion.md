# SDK gap completion

[中文](gap-completion.zh-CN.md)

- Decision date: 2026-10-09.
- The user requested completion of the remaining SDK gaps.
- This route extends the accepted Darabonba pipeline and replaces stale next-step text.
- Baseline: STS 4, ECS 380 and VPC 403 generated actions. Generation, compilation, policy review, offline behavior, live behavior and publication are separate measures.
- Keep one issue and branch per reviewable change. Do not manufacture complete coverage by marking unknown behavior as reviewed.

## Delivery order

1. Correct current route, coverage and acceptance documentation. Keep historical evidence and its revision.
2. Expand native pagination, lifecycle waiters, input constraints and conservative retry policy. Resolve paths against full IR. Record operation-specific evidence. Do not infer safety or pagination from names.
3. Add renewable OIDC/SAML STS adapters and token-file loading. Reuse generated STS clients and the shared credential cache. Reload token material for each refresh.
4. Expand credential sources: ECS metadata, URI, process and CloudSSO. Specify bounded retrieval, refresh, precedence, redaction and configuration before implementation. Long-lived keys still require explicit opt-in.
5. Expand endpoints from pinned official declarations and reviewed regional/network rules. Preserve explicit overrides and unsupported-combination errors.
6. Extend protocol lowering and runtime for ROA, JSON bodies, XML and streaming. Use pinned official products as acceptance fixtures; helpers alone do not satisfy product acceptance. Keep streaming ownership, redirects and retry replayability explicit.
7. Improve documentation coverage. Fill missing descriptions from reviewed sources and keep English and Chinese guides equivalent. Do not claim translations absent from the input.
8. Complete reproducible runtime/dependency comparisons under #20. Report results and versions, including unfavorable results.
9. Complete authorized live gaps under #79 and OAuth rotation evidence. Keep real federation, human usability (#76), release (#61) and indexing separate.

## Acceptance

- Each implementation includes behavior tests, deterministic offline Examples, package docs and paired guides.
- Generator changes include official frontend tests/check, deterministic generation and invalid-input no-write checks.
- Relevant Go changes run formatting, doccheck, vet and tests. CI checks Linux race and Windows.
- Publish a capability inventory with applicable, reviewed, unreviewed and unsupported scopes. Raw action coverage does not imply adapter coverage.
- Live work records account/resource scope before execution. Use existing authorized resources first. Resource provisioning requires a concrete scope; do not create billable resources just to obtain passing evidence.
- An implementation-agent result cannot become independent human usability evidence.
- Release and browser indexing remain incomplete until actually performed.

## Tracking

- New implementation issues are linked from the parent gap-completion issue.
- Existing #20, #76, #79 and #61 are reused. Preserve their original evidence and scope; record any revised scope explicitly.
- The parent stays open until every required child and evidence item is complete.

## Actual issues

- Parent: #87
- #88 Expand reviewed ECS and VPC pagination, lifecycle waiters and retry
- #89 Add renewable OIDC and SAML STS credential providers
- #90 Expand temporary credential sources and profile modes
- #91 Expand official endpoint rules and network selection
- #92 Extend Darabonba generation beyond RPC form requests
- #93 Fill SDK prose gaps and repair current acceptance guidance
- #94 Verify native OAuth rotation and live federation renewal
