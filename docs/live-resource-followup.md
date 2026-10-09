# Existing-resource live follow-up

[中文](live-resource-followup.zh-CN.md)

- Issue: #79; parent #87. Partial evidence, 2026-10-09.
- Authorization: query existing resources through the local Profile. No provisioning or state-changing operation.
- SDK source: `1a58fe9f26fb2d4ddab713bd561b5ae65c4610a9`; Go 1.27.1 windows/amd64; Aliyun CLI 3.4.11.
- Native SDK loader: `LoadDefaultConfig` with explicit `oss-sftp` OAuth Profile. No static credential injection.
- Official sources remain those recorded in `sources/darabonba/manifest.json`. The generated ECS/VPC clients and policy-generated paginator/waiter are used directly.

## Scope established before execution

- Query the configured cn-hangzhou region first. Its ECS instance/status and VPC samples are empty.
- Discover regions through generated ECS `DescribeRegions`, then query one instance and one VPC page per region, with bounded contexts.
- If existing VPCs are found, compare a small native page against CLI in the same region and verify input ownership, terminal cursor and response metadata.
- If a VPC is already Available, call `Wait` and reusable `WaitForOutput` with its exact ID. Do not change its state.
- Keep raw responses and disposable harnesses under ignored `.git/`. Publish no credentials, account/resource IDs, names or response bodies.

## Observations

- All 33 discovered regions returned successful ECS/VPC reads.
- No ECS instances were found. Beijing and Zhangjiakou each contain one VPC; the other 31 regions contain none.
- In each nonempty region, generated `DescribeVpcsPaginator` with PageSize=1 returned one terminal page.
- CLI and typed SDK output agree on selected fields and their presence: VpcId, RegionId, CidrBlock, Status, IsDefault, Ipv6CidrBlock and EnableIpv6.
- Native PageNumber, PageSize and TotalCount match. Each compared response has HTTP 200, one attempt and a nonempty request ID matching the typed field.
- Caller paginator/waiter inputs remain unchanged. `Wait` and `WaitForOutput` succeed on each existing Available VPC; the latter returns the requested ID.
- A pre-canceled waiter preserves `errors.Is(context.Canceled)`. This is cancellation evidence with a live-configured client, not a service-side transition.

| Required follow-up | Result | Limit |
|---|---|---|
| Nonempty VPC typed response and native terminal page | PASS | One VPC per region |
| Existing VPC Available waiter | PASS | Immediate successful state; no transition |
| VPC multi-page continuation | NOT RUN | No region contains two VPCs; resources in different regions are not one traversal |
| ECS nonempty token continuation | NOT RUN | No instances found |
| ECS all-ID waiter transition | NOT RUN | No instances; no changes authorized |
| OpenAPI Explorer browser evidence | NOT RUN | SDK/CLI evidence only |
| Federation renewal | NOT RUN | Configuration to be provided separately under #94 |

- #79 remains open. Empty pages and successful current-state reads do not satisfy multi-page or transition acceptance.
- Native OAuth token rotation and natural expiry are separate #94 criteria; this run does not claim them.

## Optional browser inspection

- Open [DescribeVpcs in Explorer](https://api.aliyun.com/api/Vpc/2016-04-28/DescribeVpcs).
- Select the same account, RegionId `cn-beijing` or `cn-zhangjiakou`, PageNumber=1 and PageSize=1.
- Inspect the Vpcs.Vpc container and native page metadata. Do not share identifiers or credentials in public evidence.
- Browser results remain separate until the user supplies an observation.
