# Native page cursor widths

[中文](native-page-widths.zh-CN.md)

- Issue: #88. Extend the shared Darabonba policy renderer, not handwritten service methods.
- Review the pinned DSL below before adding policies. These operations query existing attributes or authorization relationships; retry remains caller opt-in.
- Add typed decimal-string and int64 page inputs while preserving native response widths. Do not synthesize NextToken.
- Allow an explicitly reviewed response page name that differs from the input (`Page` versus `PageNumber`). Keep size and total validation; optional omitted page values remain valid, present values must match.
- Parse numeric fields before narrowing. Bound the cursor by the platform int range and native input width; stop before overflow.
- Own caller inputs/options. Failed, canceled and invalid response fetches must preserve the cursor.

## Reviewed operations

| Product | Action | Input page/size | Response page/size/total | Items | Default / maximum |
| --- | --- | --- | --- | --- | --- |
| ECS | DescribeInstanceAutoRenewAttribute | string | int32 | InstanceRenewAttributes.InstanceRenewAttribute | 10 / 100 |
| ECS | DescribeInstanceMaintenanceAttributes | int64 | int32 | MaintenanceAttributes.MaintenanceAttribute | 10 / 100 |
| VPC | DescribeEcGrantRelation | int64 | Page: int32 / int32 / int32 | EcGrantRelations | 10 / 50 |
| VPC | DescribeGrantRulesToEcr | int64 | int32 | EcrGrantRules | 10 / 50 |

- Pinned source `ec489e5c3deae95496daae2b41503ac58b221adb`: [ECS auto-renewal](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L24076), [ECS maintenance](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L24493), [VPC EC grants](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L22219), [VPC ECR grants](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L24469).

## Acceptance established before implementation

- Generated two-page Examples for all four operations, with exact native wire values.
- Independent wire fixtures for defaults, caller-selected starts, string syntax/range failures, wide values without truncation, missing response page, mismatched metadata, terminal/empty pages, cancellation, failure stability and ownership.
- Invalid policy types/paths fail before writes; unchanged operation/model discovery.
- Node 22 frontend tests/check; generator check/product-check; formatting, doccheck, vet, Go tests; final-head Linux race and Windows CI.
- Update paired guides and deterministic policy coverage. Remaining candidates remain unreviewed; no live acceptance is implied.

## Implementation evidence

- Four adapters and their external two-page Examples pass. ECS now has 14 paginators (16/380 reviewed actions); VPC has 15 (15/403 reviewed actions). Waiter counts are unchanged.
- Shared numeric parsing preserves native int64/string input types, int32 responses and the explicit `Page` alias. Invalid policies leave existing outputs unchanged; page policies do not change model emission.
- Node 22.21.1 frontend check and 87 tests, both generator checks, doccheck for 16 public packages, vet, formatting and language/link checks passed.
- All Go packages passed. Two stale test expectations were corrected after the first run; the complete generator package then passed in 84.398 seconds. No runtime/service test failure occurred.
- Live validation is NOT RUN for these four operations. Final-head CI is recorded in the implementation PR before merge; #88 remains open for other candidates.
