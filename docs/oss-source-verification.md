# OSS source verification

[中文](oss-source-verification.zh-CN.md)

- Issue [#92](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/92), reviewed on 2026-10-10.
- Compare the eight XML-shape blockers in [the staged coverage report](research/oss-semantic-coverage.json).
- DSL revision: `ec489e5c3deae95496daae2b41503ac58b221adb`.
- Native reference: `alibabacloud-gateway-oss-util` v0.0.6. Hashes remain in [the semantic pins](../metadata/oss-semantic-pins.json).
- Official documentation was read using the web tool. It is current documentation, not a version-pinned response or an Explorer result.
- The web tool could not read the GetBucketCors and ListBuckets Explorer pages. No browser-control tool was available. This does not prove that Explorer lacks these APIs.
- Explorer schema inspection: **NOT RUN** for all eight operations.
- Explorer live requests and SDK live requests: **NOT RUN** for all eight operations.
- No source bytes, approvals, models or coverage counts changed. Discovery remains 90 operations, with 16 lowered and 74 unsupported. OSS Go emission remains pending.

## Documentation findings and remaining checks

| Operation | Official reference evidence | Remaining check |
| --- | --- | --- |
| GetBucketCors | [GetBucketCors](https://help.aliyun.com/zh/oss/developer-reference/getbucketcors) describes header entries. [PutBucketCors](https://help.aliyun.com/zh/oss/developer-reference/putbucketcors) shows repeated AllowedHeader elements in request syntax. This supports a collection model; it is not a live response. | Read an existing rule with at least two allowed headers. Check repeated `CORSConfiguration/CORSRule/AllowedHeader` elements. A single entry does not establish maximum cardinality. |
| GetBucketInfo | [GetBucketInfo](https://help.aliyun.com/zh/oss/developer-reference/getbucketinfo) documents `BucketPolicy/LogBucket` and `LogPrefix`; its example uses those names. The pinned DSL uses `TargetBucket` and `TargetPrefix`. | Read a bucket with existing logging configuration. Compare exact child names under `BucketInfo/Bucket/BucketPolicy`. Do not describe this as merely missing newer fields in an old helper. |
| GetBucketInventory | [GetBucketInventory](https://help.aliyun.com/en/oss/developer-reference/getbucketinventory) defines SSE-OSS as a container and shows an empty element. | Read an existing inventory configured with SSE-OSS. Check presence versus absence of the empty element. An empty model and an empty string can represent the same XML; approve an explicit mapping rather than replacing the model with text. |
| ListBucketInventory | [ListBucketInventory](https://help.aliyun.com/zh/oss/developer-reference/listbucketinventory) also defines SSE-OSS as a container. Its list example does not exercise that element. | Read a list containing an existing SSE-OSS inventory. Check `InventoryConfiguration[]/Destination/OSSBucketDestination/Encryption/SSE-OSS`. |
| GetBucketReplicationLocation | [GetBucketReplicationLocation](https://help.aliyun.com/zh/oss/developer-reference/getbucketreplicationlocation) shows multiple LocationTransferType entries, but only one Type inside each TransferTypes container. | Check `LocationTransferType[]/TransferTypes/Type`. The example does not resolve scalar versus repeated Type cardinality. |
| GetBucketReplicationProgress | [GetBucketReplicationProgress](https://help.aliyun.com/zh/oss/developer-reference/getbucketreplicationprogress) requires `rule-id` and shows one Rule. Its prose also contains inconsistent parent naming. | Read an existing replication rule. One returned Rule is compatible with either a singleton model or a one-item collection; it cannot alone establish maximum cardinality. |
| GetBucketWebsite | [GetBucketWebsite](https://help.aliyun.com/zh/oss/developer-reference/getbucketwebsite) labels ErrorDocument.HttpStatus as a string. The reviewed page does not specify IndexDocument.Type. | Check existing website configuration for both fields. Numeric XML text does not prove an integer API type. HttpStatus has documentation support for the DSL string; Type remains unresolved. |
| ListBuckets | [ListBuckets](https://help.aliyun.com/en/oss/developer-reference/listbuckets) defines the root ListAllMyBucketsResult, capitalized fields and the Buckets/Bucket wrapper. It also documents optional pagination elements. | Inspect exact XML wrapping, case and pagination-field presence. A flattened JSON display is not evidence that the wire XML has no wrapper. Service-level endpoint and typed request headers remain separate lowering gaps. |

## Explorer handoff

- Open [OpenAPI Explorer](https://api.aliyun.com/) in an authenticated browser and search for OSS and the operation.
- Candidate operation links are below. They are handoff targets, **not verified routes**. If a link is unavailable or the catalog omits the operation, record that result. Do not replace an OSS data-plane operation with a similarly named control-plane API.
- Use an existing account, bucket and its region. Only perform the GET operations listed here.
- Inventory and replication reads need existing configuration IDs. CORS, logging and website reads need existing configurations. Do not create or change configuration to manufacture evidence.
- Inspect the API schema first, then the response if online debugging is available. Record whether the UI shows original XML, decoded JSON or only a schema/example.
- A 403, a missing-configuration response or an empty result does not verify the disputed field shape.
- Record operation/version, UTC request time, region, HTTP status, response format and disputed XML paths. Retain a redacted structural fragment. Keep credentials, authorization headers, tokens and real account/resource identifiers out of the repository.
- Browser schema evidence, browser live evidence and local SDK/CLI evidence remain separate. If no raw XML is visible, mark raw-wire verification NOT RUN.

| Operation | Candidate Explorer page | Existing resource |
| --- | --- | --- |
| GetBucketCors | [Open](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketCors) | Bucket with a multi-header CORS rule |
| GetBucketInfo | [Open](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketInfo) | Bucket with logging |
| GetBucketInventory | [Open](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketInventory) | Bucket and SSE-OSS inventory ID |
| ListBucketInventory | [Open](https://api.aliyun.com/api/OSS/2019-05-17/ListBucketInventory) | Bucket containing an SSE-OSS inventory |
| GetBucketReplicationLocation | [Open](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketReplicationLocation) | Bucket |
| GetBucketReplicationProgress | [Open](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketReplicationProgress) | Bucket and replication rule ID |
| GetBucketWebsite | [Open](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketWebsite) | Bucket with website configuration |
| ListBuckets | [Open](https://api.aliyun.com/api/OSS/2019-05-17/ListBuckets) | Account with existing buckets |

## Decision gate

- These findings identify candidate corrections. They do not approve runtime behavior or unblock generation.
- Any correction needs a source-bound decision, paired rationale and independent offline XML fixtures before regeneration.
- Preserve the complete DSL and original source bytes. Model overrides and serialization mappings must be explicit and reusable.
- The other 66 unsupported operations are currently blocked by method/body profiles, dynamic paths, missing body targets or service-level binding. Explorer evidence does not implement those generator capabilities.

## Verification

- This change adds documentation only. Run the documentation language/local-link check and `git diff --check`.
- No Go/frontend tests or cloud calls are required for this documentation change.
