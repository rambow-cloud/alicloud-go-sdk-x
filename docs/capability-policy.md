# Product capability policy

[中文](capability-policy.zh-CN.md)

- Current expansion #88 adds native int64/string token limits, explicit deprecated-input exclusions, multiple lifecycle waiters per action and native scalar-ID waiters. See [reviewed scope](product-capability-expansion.md). Unlisted behavior remains unreviewed.

- Stage #37 follows #36 / PR #41 on `issue/37-capability-policies`, stacked on `issue/36-batch-go-emission`.
- This original dependency stack and follow-up #44 are now [integrated into main](generator-integration.md).
- This specification preceded code.
- Policy files under `policies/` are sparse reviewed exceptions, never a required list of every API/model/field.
- Bind policy to product, API version and source-manifest hash; each operation entry has evidence references.
- Missing policy means unreviewed, no generated capabilities and no Standard retry.
- Unknown/unsupported actions, paths, types, names, modes or constraints fail before writes.
- Optional policy absence must not prevent complete supported Go emission.

- Role validation under [review fix #44](capability-role-review.md) rejects same-model aliases between paginator request page/size/limit, response page/size/total, waiter request page/size and member ID/state.
- Identical paths in different request/response models or separate adapters remain valid.
- Path existence and type matching alone are insufficient evidence for distinct behavioral roles.

- Initial coverage: native paginators for ECS DescribeInstances (token by default, explicit page fields choose pages), DescribeInstanceStatus and DescribeImages, plus VPC DescribeVpcs; ECS InstanceRunningWaiter; reviewed idempotency for these four queries and ECS DescribeRegions; ClientToken generation/validation for ECS AllocateDedicatedHosts; conservative formatting for STS AssumeRole sensitive models.
- The other emitted actions remain unreviewed.
- Per-operation reports distinguish reviewed capability policy, generated adapters, conservative retry and live acceptance.

- Paginator policy identifies wire input/output cursor paths, collection path, limit, page/size/total, adapter defaults and maxima.
- Paths are resolved against the complete IR and emitted as typed accesses with nil guards; no runtime field-name guessing.
- Constructors follow NewOperationPaginator and return errors for invalid options; HasMorePages/NextPage(ctx, operationOptions...) reuse the shared engine.
- Deep-copy input at construction and each fetch; copy options and isolate each page override.
- Errors/cancellation do not advance cursors.
- Duplicate tokens deliver the fetched page then stop by default; disabling protection can permit cycles.
- Empty token pages with continuation still advance; token completion ignores TotalCount.
- Page completion uses a reviewed total count and collection; missing totals, negative metadata, mismatched page numbers and overlarge response sizes fail.
- Empty collections stop.
- Bound page arithmetic to the native int32 field.
- Dual-mode requests reject mixing non-nil token/limit with page/size fields, including explicit empty/zero pointers.
- Status paginator uses an intentional size 50, distinct from the service default 10.
- Other initial defaults are 10; status/VPC maxima are 50, image/instance maxima 100.

- Waiter policy records ID input, native collection/member ID/state, first-page fields, maximum IDs, success/retry states and evidence.
- Reuse shared total-time/backoff engine and reusable Wait/WaitForOutput.
- InstanceRunningWaiter requires 1..50 distinct IDs, forces page one/size 50 and requires every requested ID Running.
- Missing IDs retry; Pending/Starting/Stopping/Stopped retry; duplicate requested IDs, missing/unknown states, nil results and API errors fail.
- Options allow a reviewed acceptor override, but cannot turn failed fetches/cancellation/expiry into success.
- Each invocation and poll has isolated inputs/options; an immutable waiter supports concurrent waits.

- Operation policies explicitly mark safe read replay; Standard remains opt-in.
- Neither CLI operation_type nor HTTP verbs grant retry safety.
- AllocateDedicatedHosts stays non-idempotent for retry purposes even with ClientToken: this stage deliberately does not promote token-bearing writes to automatic retries.
- Generate a cryptographically random ASCII token once on the owned input when absent, before hooks; preserve explicit tokens, validate nonempty ASCII length <=64 before execution and again after hooks.
- Caller input remains unchanged.
- Separate invocations get distinct generated tokens; callers intentionally repeating an operation supply the same explicit token.

- Generate typed public ValidateOperationInput functions for sparse bounds, ASCII/string length, array cardinality/nonempty IDs and mixed-mode exclusions.
- Nil is an empty request; optional API fields are not made required.
- Validate copied input and post-Initialize input before serialization.
- Diagnostics identify paths/rules without including values.
- An explicit zero pointer violates a reviewed positive bound, while unconstrained scalar zero/false/empty continues to serialize.
- Keep requiredness and constraint claims limited to the recorded rules.

- Sensitive model String/GoString return a constant type/redacted marker for both values and pointers, covering ordinary fmt output, including %#v.
- This hides the entire marked model, not a guessed subset.
- JSON serialization and direct field access remain explicit raw-data operations; no logger is added.
- Mark STS request/body/credentials/ response envelope.
- Naming exceptions change only Go names: initially the native DescribeInstanceStatus InstanceId array becomes InstanceIDs.
- Wire names remain exact.

- Acceptance: deterministic independent policy loading and emission; negative schema/ source/type/path/naming cases leave outputs untouched; multi-page token and page mocks, cyclic tokens, empty token pages, failed/canceled fetches, nil containers, metadata checks, input/callback isolation; all-ID waiter states, missing/duplicates, timeout/ errors/acceptor overrides and concurrent waits; signed token/presence/validator checks, opt-in read retry and non-retrying writes, fmt redaction; offline Examples for each adapter/policy, corresponding guides and package comments.
- Run Node frontend checks/ 47 tests, both Go generation checks, formatting/doccheck/vet/full Go tests once after implementation; Linux race and Windows CI; no live cloud calls or full-policy coverage.

- Evidence inspected 2026-10-07: [DescribeInstances](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-describeinstances), [DescribeInstanceStatus](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-describeinstancestatus), [DescribeImages](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-describeimages), [DescribeVpcs](https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs), [AllocateDedicatedHosts](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-allocatededicatedhosts), and pinned DSL/earlier accepted reference policies.
- Web descriptions are separate behavior evidence, not embedded mutable generation inputs.
- Waiter acceptance and conservative write retry are our reviewed SDK policy, not official waiter declarations.
