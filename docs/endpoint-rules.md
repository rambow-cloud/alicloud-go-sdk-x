# Official endpoint rules

[中文](endpoint-rules.zh-CN.md)

- Issue: #91. Endpoint data comes from pinned official Darabonba through the semantic parser.
- Product IR and lock use schema v4. Older versions fail before writes; source bytes and pins are unchanged.
- Generation owns the shared catalog `endpoint/rules.gen.go`, its Apache LICENSE and NOTICE. Original resolver/tooling retain MIT.

## Selection

- `Config.Network` is copied into service and operation `Options`. Empty or `public` uses public rules; custom resolvers receive the network too.
- `BaseEndpoint` always wins, including unknown products, regions or networks. It must be a valid HTTPS origin.
- Public resolution: exact official `endpointMap`, then the source's regional/global construction rule.
- Examples: ECS Hangzhou → `ecs-cn-hangzhou.aliyuncs.com`; VPC Hangzhou → `vpc.aliyuncs.com`; STS ap-south-1 → `sts.aliyuncs.com`.
- Regional fallback accepts one lowercase DNS label. Constructing a hostname does not prove deployment or availability in that region.
- Invalid region labels and unknown products fail with `endpoint.ErrUnsupported`. Resolution performs no HTTP or DNS calls.
- `NewRules` remains an exact-rule resolver. It copies input; empty and `public` share one key. Its zero value supports explicit overrides only.

## Private access

- `Network: "vpc"` requires an exact reviewed private rule. It never falls back to public access.
- Current reviewed private rules: ECS 11, STS 35, VPC 1 (Beijing). Policy files list regions and evidence URLs.
- VPC PrivateLink account configuration may still be required. Resolving an address does not establish connectivity.
- Unlisted private networks, suffixes and region/network combinations require `BaseEndpoint` or custom rules/resolver.
- Intentional difference: a reviewed private rule wins over the DSL's public map. The original map can return a public origin even with a network option set.

## Generator contracts

- The frontend recognizes unique constant rule/map initialization, exact endpoint arguments and explicit origin → map → rule control flow.
- Changed control flow, nonconstant expressions, duplicate mappings and unsupported rules fail projection.
- Go generation validates source coordinates, origins, private profiles, duplicates and evidence before writes.
- Endpoint policy is source-bound; it does not imply retry, pagination or operation review.
- The catalog is emitted deterministically from all product IR and preflighted with product outputs.
- Tests cover special mappings, regional fallback, private/public isolation, override priority, ownership and malformed source/policy rejection. A negative fixture verifies no partial service/endpoint writes.
- Address selection is offline evidence; private live connectivity is not claimed.

## Primary sources

- Pinned `sources/darabonba/products/{ecs,sts,vpc}/main.tea`; initialization/getEndpoint coordinates remain in IR/generated comments.
- [Official construction](https://www.alibabacloud.com/help/en/sdk/developer-reference/endpoint-configuration).
- [ECS private endpoints](https://www.alibabacloud.com/help/en/ecs/developer-reference/call-api-operations-over-the-internal-network).
- [STS endpoints](https://www.alibabacloud.com/help/en/ram/developer-reference/api-sts-2015-04-01-endpoint).
- [VPC PrivateLink](https://www.alibabacloud.com/help/en/vpc/use-privatelink-to-access-vpc-openapi-over-private-network).
