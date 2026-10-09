# Full-DSL SDK live acceptance

[中文](product-live-validation.zh-CN.md)

- Issue: [#47](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/47).
- Historical read-only evidence, reviewed for integration on 2026-10-09.
- This static-snapshot run does not replace the later native Profile/OAuth evidence or product acceptance #74/#75.

### Scope and sequence

This acceptance plan is established before execution against the full-DSL `service/ecs`
and `service/vpc` packages on baseline `dde481ce4da680f964ee3bbd1458048e6ed981bc`.
The earlier [live record](live-validation.md) covers the bounded `services/` bridge;
it does not establish live acceptance of these newer packages. Dependencies #33 and
#44 are complete. Establish a separate GitHub issue, then execute its disposable
harness on an issue branch. Runtime or generator defects require a documented issue
scope before changing production code.

Use only the user-authorized local Aliyun CLI profile `oss-sftp` and its configured
region. Run official CLI references first, permitting its existing OAuth refresh flow,
then explicitly inject its valid local credential snapshot into the SDK static provider.
Check expiration; do not print credentials or pass them as command arguments. This
does not add native profile loading or OAuth renewal to the SDK.

### Acceptance criteria

- Compare ECS DescribeRegions, DescribeImages, DescribeInstances in native token and
  page modes, DescribeInstanceStatus, and VPC DescribeVpcs with identical CLI parameters.
- Use small reviewed page sizes and at most two pages per traversal. Compare selected
  typed fields, optional presence, native cursor/page metadata and HTTP/request-ID/attempt
  metadata. Exclude per-call request IDs and timestamps from equality comparisons.
  Distinguish decode/protocol failures, authorization failures and possible live drift.
- Verify input ownership and pre-canceled context handling. If an existing Running
  instance occurs in the bounded status sample, verify Wait and WaitForOutput with
  that identity and a bounded timeout; otherwise record SKIP. Do not create or modify
  resources to obtain additional pages or state transitions.
- Keep the disposable harness and raw CLI responses in ignored `.git/` storage. Publish
  only sanitized counts, pass/fail/skip results, versions, source commit and limitations;
  omit credentials, account/resource identifiers, names and raw bodies.
- Publish paired evidence and an issue-linked PR. Explicitly report single-page
  exhaustion and skipped checks. Do not claim all 579 operations, live retry failures,
  credential renewal, tracing, STS AssumeRole or waiter transitions from these reads.

### Browser verification

SDK/CLI execution is separate from OpenAPI Explorer browser evidence. For optional
manual inspection, sign in to the same account and select the profile region:

- [DescribeRegions](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeRegions)
- [DescribeImages](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeImages)
- [DescribeInstances](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstances)
- [DescribeInstanceStatus](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstanceStatus)
- [DescribeVpcs](https://api.aliyun.com/api/Vpc/2016-04-28/DescribeVpcs)

Inspect CLI Example, matching RegionId and native token/page parameters, successful
response containers and cursor metadata. Browser results are not part of automated
SDK/CLI acceptance and are recorded only if independently provided by the user.

### Evidence

Executed on 2026-10-08 against SDK source commit
`dde481ce4da680f964ee3bbd1458048e6ed981bc`, Go 1.27.1 windows/amd64, Aliyun CLI 3.4.11,
and the explicitly selected `oss-sftp` OAuth profile in cn-hangzhou. Only documentation
changed during this acceptance; the disposable harness imports `service/ecs` and
`service/vpc`, not the older `services/` bridge. SDK credential injection uses an
expiration-checked local static snapshot; no public credential provider was added.

The initial harness selected the old `sftp-oss` profile, whose cached credentials
were expired. CLI client initialization failed while refreshing its OAuth token
(HTTP 400), before any ECS response. Inspection identified the user's newly signed-in
profile as `oss-sftp`; the plan, issue and harness were corrected before executing
successful references. This was a local acceptance profile selection correction,
not an SDK production fix. Credential values and raw diagnostics are excluded.

| Check                                 | Result  | Evidence and limit                                                                                                                                  |
| ------------------------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| ECS DescribeRegions                   | PASS    | 33 regions; selected wire fields and scalar presence match CLI                                                                                      |
| ECS DescribeImages page paginator     | PASS    | ImageOwnerAlias=system, PageSize=2; two pages, four distinct image identities, more=true; native page metadata and selected scalar fields match CLI |
| ECS DescribeInstances token paginator | PASS    | MaxResults=1; one empty terminal page; no live token continuation covered                                                                           |
| ECS DescribeInstances page paginator  | PASS    | PageSize=1; one empty terminal page                                                                                                                 |
| ECS DescribeInstanceStatus paginator  | PASS    | PageSize=1; one empty terminal page                                                                                                                 |
| VPC DescribeVpcs paginator            | PASS    | PageSize=1; one empty terminal page                                                                                                                 |
| Request and paginator ownership       | PASS    | Serialized caller inputs unchanged through all successful reads/pages                                                                               |
| Pre-canceled context                  | PASS    | ECS/VPC operations and the image paginator preserve errors.Is(context.Canceled); paginator cursor remains unconsumed                                |
| ECS Wait / WaitForOutput              | SKIP    | No Running identity in the bounded status sample                                                                                                    |
| STS AssumeRole                        | SKIP    | No explicit target role supplied                                                                                                                    |
| OpenAPI Explorer browser UI           | NOT RUN | SDK/CLI evidence only; optional manual links above                                                                                                  |

Every successful SDK response had HTTP 200, a nonempty request ID matching its typed
RequestId, and exactly one attempt. Equality checks used independently decoded CLI
wire fields and JSON projections of typed SDK outputs, preserving presence of selected
scalar fields; per-call request IDs were checked internally rather than compared
across requests. Item projections were sorted within a page; opaque tokens were
compared by continuation presence. CLI and SDK would each use their own token for
the next page, but this account returned no instance continuation to exercise that path.
The harness bounded each paginator to two pages and checked native continuation and
duplicate identities without exhausting the public image catalog.

This establishes real page-number continuation for DescribeImages and successful
empty-result decoding/exhaustion for the sampled account reads. It does not establish
nonempty instance/VPC model coverage, instance token continuation, waiter polling or
transitions, automatic retries, credential renewal, tracing, other regions, all generated
operations, or pkg.go.dev publication. Raw references and harness remain ignored under
`.git/product-live-validation/`; no resources were created or modified.
