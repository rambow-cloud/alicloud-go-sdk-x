# Live profile validation

[中文](live-validation.zh-CN.md)

- This path is written before running live acceptance against the current generated SDK.
- Use the explicitly selected local Aliyun CLI profile and its region.
- Compare ECS DescribeRegions, DescribeInstances (native token and legacy pages), DescribeInstanceStatus and VPC DescribeVpcs with the same read-only parameters through Alibaba Cloud CLI.
- OpenAPI Explorer supplies CLI examples on each API debugging page; CLI comparison is not a claim that the Explorer browser UI was exercised.

- Run official CLI references first so OAuth may refresh its cached temporary credentials, then read the selected profile locally into an explicit static SDK provider.
- Validate the credential expiration; never print/copy credentials into source, reports or CLI arguments.
- This is an opt-in acceptance harness, not implicit SDK profile discovery or a general OAuth provider.
- Profiles requiring other renewal flows are outside this path.

- Compare selected output projections and pagination metadata after excluding request IDs and timestamps.
- Bound traversal to at most two pages per policy.
- Live resources can change between calls: distinguish a comparison mismatch from an SDK protocol error.
- If a Running ECS identity exists, verify Wait and WaitForOutput using that identity; otherwise record the waiter check as skipped.
- STS AssumeRole requires an explicit target role and is not exercised without one.
- Retry failures, mock helpers, credential renewal and tracing continue to rely on the existing offline tests; live reads alone cannot establish their acceptance.
- No cloud resources are created or modified.

- Keep raw CLI responses and the disposable Go harness under ignored .git/ storage.
- Publish only sanitized pass/fail/skip evidence, SDK commit and CLI version, not resource IDs, names, account IDs or response bodies.
- Use an issue established before the harness.

- Browser verification: sign in to the same account, select the same region and open:

- [DescribeRegions](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeRegions)
- [DescribeInstances](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstances)
- [DescribeInstanceStatus](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstanceStatus)
- [DescribeVpcs](https://api.aliyun.com/api/Vpc/2016-04-28/DescribeVpcs)

- Use CLI Example to inspect the reference commands.
- Debug with RegionId set to the CLI region: token mode uses MaxResults=10, page mode uses PageNumber=1/PageSize=10 (no token parameters).
- Check HTTP success, response containers, token/page metadata and selected fields.
- Browser checks must be recorded separately from SDK/CLI execution.

- Source: [official CLI/Explorer guide](https://help.aliyun.com/en/openapi/developer-reference/new-cli-guide).

### Recorded acceptance

- Issue [#30](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/30), 2026-10-07, SDK commit `618303bf2091afc37c1713025ffa6e9248dd0b8c`, Aliyun CLI 3.4.11, cn-hangzhou.
- Initial OAuth/STS cache was expired; the first CLI reference failed authentication.
- After the user renewed the profile, all five CLI references and corresponding SDK reads succeeded.
- Credentials, resource identities and raw responses are excluded from evidence.

| Check                                 | Result  | Evidence / limit                                                                  |
| ------------------------------------- | ------- | --------------------------------------------------------------------------------- |
| ECS DescribeRegions                   | PASS    | selected region fields match CLI; successful HTTP/request-ID/one-attempt metadata |
| ECS DescribeInstances token paginator | PASS    | selected instances and continuation presence match; one terminal page             |
| ECS DescribeInstances page paginator  | PASS    | selected instances and native page metadata match; one terminal page              |
| ECS DescribeInstanceStatus paginator  | PASS    | selected statuses and page metadata match; one terminal page                      |
| VPC DescribeVpcs paginator            | PASS    | selected VPC projections and page metadata match; one terminal page               |
| ECS Wait / WaitForOutput              | SKIP    | no Running identity in the bounded status sample                                  |
| STS AssumeRole                        | SKIP    | no explicit target role supplied                                                  |
| OpenAPI Explorer browser UI           | NOT RUN | no browser tool/session available; manual links and steps above                   |

- The live service ended each paginator on the first page, so multi-page continuation, duplicate tokens, overflow, retries and wait behavior remain covered by offline regression fixtures rather than this account's live run.
- No write/resource lifecycle calls were made.
- This demonstrates the current SDK's explicit temporary credential injection; it does not add OAuth renewal or native Profile loading to the public credential chain.
