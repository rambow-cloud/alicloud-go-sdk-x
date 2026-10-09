# Real-product generator expansion

[中文](generator-expansion.zh-CN.md)

- Current client paths follow [service consolidation #81](service-consolidation.md): the old services/ packages and Go emitter are removed; generate/check use complete products. Earlier bridge workflows below retain historical evidence only.

- Historical #24/#25 expansion plan.
- The [new product route](product-generator-roadmap.md) takes precedence; per-operation metadata is not a future DSL discovery prerequisite.

- After #8/#21–#23 passed, the next gate is a real RPC product outside ECS/STS.
- Use Vpc API 2016-04-28 DescribeVpcs: reviewed read-only/idempotent query, POST `/`, JSON response.
- Pin official operation metadata before implementing the client.
- Generic schema work #24 precedes VPC integration #25 (which depends on #24).
- ROA/body profiles, benchmark #20 and preview release are subsequent stages.

| Shape                   | Previous profile     | Expansion and acceptance                                                                                                           |
| ----------------------- | -------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| Boolean                 | rejected             | response bool; optional request \*bool preserves nil vs false                                                                      |
| Explicit scalar zero    | omitted              | *string/*int/\*int64 preserve empty/zero vs absence; required and bounds checked                                                   |
| repeatList object array | rejected             | named input model, Tag.N.Key/Value, recursive input copying; pointer Value preserves explicit empty string                         |
| Nested response         | top-level paths only | named models can flatten nested paths; JSON v2 methods decode/encode reviewed containers atomically                                |
| Local references        | rejected             | offline #/components/schemas refs, RFC6901 key escapes, bounded chains; dangling/cyclic/external refs and structural siblings fail |
| Pagination              | token + legacy pages | reviewed page-only profile reuses pagination engine; start/default/size/total/error/overflow checks                                |
| Unsupported schema      | error                | composition/maps/general nested query objects remain errors; unselected optional structures do not expand coverage                 |

- Reference resolution is lazy over selected paths, without network or source mutation.
- No arbitrary schema-wide expansion or guessed types.
- Named model dependencies must be acyclic; unsupported or newly required selected model members fail before output.
- Input models are declared explicitly with location=input and a parameter path ending in [] for repeatList items.
- Models stay bound to their reviewed operation/schema path.
- Pointers are used only for explicit presence; ordinary fields retain current APIs.
- Generated copying owns input pointers, slices and nested model members before hooks.
- Scalar bool input must be a pointer because zero and absence differ.

- VPC coverage selects region/VPC/name/resource-group filters, boolean filters, owner ID, page parameters and tags; outputs select identifiers, CIDRs, booleans, tags, IPv6 blocks, vSwitch IDs and page metadata.
- Prose-only tag syntax and positive paging constraints remain a handwritten validator, not a generator service branch.
- A generated page-only DescribeVpcsPaginator uses totals and nonempty pages, with overflow-safe termination.
- The existing shared runtime handles signing, credentials, retry, errors, middleware and tracing unchanged.
- Review the same five public-region VPC endpoints explicitly in the default resolver; unknown regions/partitions still require explicit reviewed rules.

- Validation is offline: generator fixtures exercise all new scalar presence states, named input copying, nested response JSON and local-ref failure modes; generated VPC tests assert actual signed query encoding, body/response handling, caller ownership, bounded retry, page-only traversal, cancellation, errors and endpoint rules.
- Preserve existing ECS/STS protocol/acceptance tests.
- Public doc.go, English symbol comments, deterministic external Examples, bilingual guides and CI regeneration are mandatory.
- No live account, complete VPC coverage or pkg.go.dev indexing is claimed.

- Sources: [DescribeVpcs](https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs), [VPC endpoints](https://help.aliyun.com/zh/vpc/developer-reference/api-vpc-2016-04-28-endpoint), [official metadata](https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/).
