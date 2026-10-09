# [Maintenance]: Accept generated VPC product before v0.1.0 release

### Problem and route

- The user revised delivery on 2026-10-09: finish #60, then ECS/VPC acceptance, then #61 publication.
- Full-DSL service/vpc emits 296 actions. Generation and compilation are not product-wide live acceptance.
- Follow docs/sts-ecs-vpc-path.md, docs/product-acceptance.md and the official Darabonba/parser -> complete IR -> shared Go backend/runtime route.
- Member of #57; blocked by #60 and scoped live evidence #47. Parent membership is not a blocking dependency.

### Scope

- Pin complete VPC 2016-04-28 DSL/imports, semantic parser and source-bound policy; account for complete operations/models and separate coverage stages.
- Accept DescribeVpcs traversal and business-consumer workloads using its native page-number paginator.
- Review optional/nested response fields, source ownership, shared credentials/cache/Profile, retry, endpoints, errors and middleware/OTel.
- Unlisted actions remain unreviewed for adapters/retry. No invented NextToken or waiter, and no policy inferred from operation names.

### Acceptance

- [ ] Complete pinned RPC inventory regenerates deterministically, compiles and runs offline Examples; unsupported selected behavior fails before writes.
- [ ] A pinned isolated consumer traverses VPCs using HasMorePages/NextPage and a small operation mock; classify errors/cancellation without SDK internals.
- [ ] Multi-page/empty/short pages, totals/bounds, optional presence, malformed responses and stable paging state after errors/cancel pass independent fixtures.
- [ ] Native page wire fields, nested response containers, operation options and input/config ownership remain exact; no token pagination is fabricated.
- [ ] Reviewed read retry and conservative write behavior, signed endpoints, shared STS/Profile credentials, structured Metadata and secret-safe middleware/OTel pass consumer contracts.
- [ ] Review/integrate #47 evidence. Record actual continuation separately from empty one-page exhaustion, selected fields separately from whole models, and CLI evidence separately from Explorer browser evidence. Declare mandatory live cases before execution.
- [ ] Paired product report/usage/coverage/license guidance and runnable Examples map to applicable AC/UX criteria. Waiter is explicitly not applicable to this scoped workload.

### Verification and limits

- Node 22 frontend checks/tests before both sdkgen checks, doccheck/vet/Go tests/Examples, isolated consumer, Linux race and Windows CI.
- No cloud resource creation/write, IdP setup or performance benchmark is authorized by this planning issue.
- Required NOT RUN/FAIL remains blocking; no all-296-action live or full-cloud claim.
- Publication #61 is separate and waits for this issue and ECS acceptance.

### Execution matrix established 2026-10-09

- Follow docs/vpc-product-acceptance.md and its paired Chinese guide. Implementation-agent acceptance, public isolated consumer and no network/account.
- Required offline cases: native multi-page/empty/short traversal, boundaries, failure/cancel stability, input/options ownership, nested absence/false/int64, malformed responses, Profile/shared generated STS cache, reviewed read retry, conservative write behavior, endpoint/error/Metadata/OTel and requestless smoke.
- Required live case reuses #47 successful selected-field DescribeVpcs empty terminal read. Its sampled contract/runtime/policy are unchanged. Nonempty live continuation is optional/excluded for scoped v0.1.0 because the account sample has no VPCs; keep NOT RUN, require offline continuation and create no resources. Broader Beta live requirements are not completed.
- Reuse unchanged ECS-stage frontend/generator/root gates, run new consumer/format/language gates and final-head Linux race/Windows/automation. Refresh all shared-workload evidence pins before release readiness.
- No invented NextToken/waiter, source change, cloud call, benchmark or publication.
