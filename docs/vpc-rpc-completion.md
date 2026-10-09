# Complete VPC RPC generation

[中文](vpc-rpc-completion.zh-CN.md)

## Scope and sequence

- Issue: [#85](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/85). Depends on #83; blocks #61.
- Extend the shared official Darabonba parser -> IR -> Go pipeline for the seven remaining pinned VPC actions.
- Four VPN writes split fields between URL query and form body. Recognize the exact bodyFlat/query/parseToMap program; preserve field locations and one-based array indexes.
- DescribeVpnGatewayAvailableZones uses whole-model query conversion and GET. Preserve its native method and wire names.
- GrantInstanceToVbr and RevokeInstanceFromVbr use explicit simple array encoding. Keep structured inputs and join string elements with commas before URL encoding.
- Do not infer these rules from operation names. Reject unsupported programs and ambiguous bindings before writes.
- Product IR and lock use schema v3 for the new binding locations, encodings and HTTP methods. Older backends reject incompatible IR before writes.
- Keep retry opt-in, source bytes and sparse policies unchanged. Preserve caller input ownership and context cancellation.

## Acceptance

- Pinned VPC: 403 discovered/lowered/emitted/compiled. ECS remains 380; STS remains four actions.
- Official-source and renamed-fixture tests prove shared lowering. Negative tests cover invalid transforms, body programs, methods and duplicate bindings.
- Signed offline requests check query/body separation, form content type, payload signing, GET, simple arrays, nil/empty/zero values, nested copies and cancellation before HTTP.
- Compare form/simple helpers with the pinned official implementation in the isolated consumer module. No account or network is needed.
- Add deterministic external Examples, English Go comments and paired product guidance.
- Run Node 22 checks/tests first, then regeneration/product-check, doccheck, vet, Go tests and formatting. CI retains Linux race and Windows checks.
- Refresh all isolated consumer records at one clean final workload after CI configuration is fixed. Preserve historical live evidence.
- Generated and compiled inventory is not all-action live acceptance. Release/indexing remains #61.

## Status

- VPC: 403 discovered/lowered/emitted, 1,728 complete models; no remaining pinned lowering gaps. ECS: 380/2,053 models. STS: four/19 models.
- Node 22 frontend/IR checks and 84 tests PASS. Product-check, doccheck (15 packages), vet and root tests/Examples PASS. Additional schema/location rejection tests PASS after their final test edits.
- Existing public model field types are preserved: 2,053 ECS, 19 STS and 1,671 VPC models. Upstream source and capability policies are unchanged.
- Formatting, 25 automation tests and paired language/local-link checks PASS. Historical live scope is unchanged.
- Consumer workload and final PR CI evidence will be recorded after the clean implementation commit.
