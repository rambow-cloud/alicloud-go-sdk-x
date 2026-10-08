# DSL and metadata decisions

[中文](darabonba-decisions.zh-CN.md)

- Issue [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31), 2026-10-07.
- DSL revision ec489e5c3deae95496daae2b41503ac58b221adb; public metadata versions and raw/extracted SHA-256 are recorded in metadata/{ecs,sts,vpc}/manifest.json.
- Machine-readable approvals live in metadata/darabonba-decisions.json, pinned by each product manifest.
- Both frontend and Go checks reject new input coverage or requiredness differences; selected wire type/path/binding differences also fail.
- Approval edits require issue evidence and an equivalent update to this document.
- Approved DSL-only fields must remain optional.
- Requiredness approvals only allow API required / DSL optional; reversing that direction requires a new supported policy.

| Difference                                                                                                        | Reviewed disposition                                                                                                                                                                             | Behavior evidence                                                                            |
| ----------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------- |
| ECS/VPC DSL adds OwnerAccount, OwnerId, ResourceOwnerAccount, ResourceOwnerId absent from public snapshots        | Keep outside the public subset. DSL presence does not grant a supported contract                                                                                                                 | Source comparison; no live owner-override calls                                              |
| DescribeInstances DSL has a repeated Filter model; snapshots expose Filter.1.Key/Value through Filter.4.Key/Value | #34 normalizes eight aliases and removes Filter from DSL-only approvals. Preserve the legacy Go subset; indexes are not inferred server limits                                                   | Both sources checked and offline leaf/type/case/requiredness tests; Filter not tested live   |
| DescribeInstances.Tag metadata contains optional lowercase key/value, absent from current DSL                     | Explicit metadataOnlyFields approval for optional string Tag[].key and Tag[].value; retain exact case and keep outside the public subset. Other missing fields or type/requiredness changes fail | Pinned canonical/public snapshot and semantic DSL comparison; no live deprecated-field calls |
| DescribeInstances, DescribeInstanceStatus, DescribeVpcs DSL marks RegionId optional; metadata requires it         | Retain API requiredness and existing configured-region fallback                                                                                                                                  | Offline missing-region checks and prior #30 authorized reads with configured region          |
| AssumeRole DSL marks RoleArn/RoleSessionName optional; metadata requires both                                     | Retain local required-field validation before HTTP; DSL optionality models unset values, not successful omission                                                                                 | Existing offline STS tests; live AssumeRole skipped without an explicit role                 |
| Metadata allows GET and POST; DSL fixes POST for these operations                                                 | Use DSL POST and preserve current signed RPC encoding                                                                                                                                            | Existing signed-wire fixtures and prior #30 reads; GET not exposed                           |
| DSL response wraps headers/statusCode/body; metadata describes the body                                           | Project body; common runtime provides response metadata and structured errors                                                                                                                    | Existing response/error/middleware tests and prior #30 reads                                 |

- These are deliberate SDK contract decisions, not a claim that hidden fields are invalid or that service-side requiredness has been experimentally proven.
- API requiredness comes from pinned public metadata; it is conservative client validation.
- The full product source/imports pass official semantic analysis; lowering accepts only the documented profile.
- Numeric Go widths, nullable/missing-value behavior, JSON-string arrays, time conversion and API constraints remain reviewed metadata/ overlay policy.
- DSL prose is not automatically translated into validators.

- For a material new conflict: preserve the two source versions, reproduce with the same parameters using Explorer/its CLI example, classify service behavior separately from CLI local validation, record sanitized result and request date, decide the supported contract, then update approvals/tests/docs before regeneration.
- Never resolve a selected incompatible wire type/path by silently choosing one source.

- Open the exact Explorer pages and use the same authorized account/region:

- [DescribeRegions](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeRegions): inspect the public request fields; do not send owner overrides.
- [DescribeInstances](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstances): inspect RegionId and Filter visibility; compare MaxResults=10 with PageNumber=1/PageSize=10 separately.
- [DescribeInstanceStatus](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstanceStatus): inspect RegionId requiredness and page response fields.
- [DescribeVpcs](https://api.aliyun.com/api/Vpc/2016-04-28/DescribeVpcs): inspect RegionId and native page metadata.
- [AssumeRole](https://api.aliyun.com/api/Sts/2015-04-01/AssumeRole): inspect RoleArn/RoleSessionName requiredness; successful debug calls require an explicitly authorized target role.

- Browser verification status: **NOT RUN**.
- UI schema hints, CLI-generated examples, CLI local validation and actual HTTP responses are different evidence.
- Prior live SDK/CLI evidence is recorded in [live-validation.md](live-validation.md); it predates this frontend migration and does not verify owner fields, Filter or omitted required parameters.
- Regenerated Go service output is checked byte-for-byte against the previous metadata backend by TestOfficialDSLJoinsAllProductsWithoutChangingGoContracts.
- Do not report previous runs as new migration or Explorer browser passes.
