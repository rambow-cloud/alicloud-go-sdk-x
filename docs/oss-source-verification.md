# OSS source verification

[中文](oss-source-verification.zh-CN.md)

- Issue [#92](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/92), reviewed on 2026-10-10.
- Compare the eight XML-shape blockers in [the staged coverage report](research/oss-semantic-coverage.json).
- DSL revision: `ec489e5c3deae95496daae2b41503ac58b221adb`.
- Native reference: `alibabacloud-gateway-oss-util` v0.0.6. Hashes remain in [the semantic pins](../metadata/oss-semantic-pins.json).
- Official documentation was read using the web tool. Explorer's public metadata API was then fetched for all eight operations: HTTP 200, with JSON pointers and response hashes in [the evidence record](acceptance/oss-source-live.json). API metadata is distinct from browser UI evidence.
- Browser schema inspection and browser live requests: **NOT RUN**. The web tool could not read the interactive pages; no browser-control tool was available.
- Native CLI verification used `aliyun` 3.4.11 / `ossutil` 2.4.0 and the existing `oss-sftp` OAuth profile. One existing bucket was found in cn-beijing. All requests were read-only, with zero retries and bounded timeouts.
- SDK raw-wire verification: **PASS** for ListBuckets, GetBucketInfo and GetBucketReplicationLocation after the candidate [#126](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/126) fix. Each returned HTTP 200/application/xml in one attempt. This uses the shared low-level runtime, not a generated OSS client.
- Before #126, the signer rejected the period in native `STS.` temporary AccessKey IDs before HTTP. The evidence binds the baseline revision and corrected signer blob. CLI and raw SDK response structures agree.
- No source bytes, approvals, models or coverage counts changed. Discovery remains 90 operations, with 16 lowered and 74 unsupported. OSS Go emission remains pending.

## Documentation findings and remaining checks

| Operation | Official reference evidence | Live check status |
| --- | --- | --- |
| GetBucketCors | [GetBucketCors](https://help.aliyun.com/zh/oss/developer-reference/getbucketcors) describes header entries. [PutBucketCors](https://help.aliyun.com/zh/oss/developer-reference/putbucketcors) shows repeated AllowedHeader elements in request syntax. This supports a collection model; it is not a live response. | **SKIP**: HTTP 404 / NoSuchCORSConfiguration. No multi-header rule exists to inspect. |
| GetBucketInfo | [GetBucketInfo](https://help.aliyun.com/zh/oss/developer-reference/getbucketinfo) uses `BucketPolicy/LogBucket` and `LogPrefix`; the pinned DSL uses `TargetBucket` and `TargetPrefix`. | **PASS for names**: CLI and raw SDK XML contain LogBucket/LogPrefix, both empty, with no TargetBucket/TargetPrefix. Nonempty logging behavior remains NOT RUN. |
| GetBucketInventory | [GetBucketInventory](https://help.aliyun.com/en/oss/developer-reference/getbucketinventory) defines SSE-OSS as a container and shows an empty element. | **NOT RUN**: no inventory ID found. Empty model/string equivalence needs an explicit serialization decision; do not infer it from an empty list. |
| ListBucketInventory | [ListBucketInventory](https://help.aliyun.com/zh/oss/developer-reference/listbucketinventory) also defines SSE-OSS as a container. | **SKIP**: HTTP 404 / NoSuchInventory. No SSE-OSS configuration exists to inspect. |
| GetBucketReplicationLocation | [GetBucketReplicationLocation](https://help.aliyun.com/zh/oss/developer-reference/getbucketreplicationlocation) shows only one Type per TransferTypes in its example. | **PASS for repetition**: CLI and raw SDK XML contain 21 TransferTypes entries: eight with one Type and thirteen with two. A scalar loses returned data. |
| GetBucketReplicationProgress | [GetBucketReplicationProgress](https://help.aliyun.com/zh/oss/developer-reference/getbucketreplicationprogress) requires `rule-id` and shows one Rule. | **NOT RUN**: discovery GetBucketReplication returned HTTP 404 / NoSuchReplicationConfiguration. No rule ID was invented. |
| GetBucketWebsite | [GetBucketWebsite](https://help.aliyun.com/zh/oss/developer-reference/getbucketwebsite) labels HttpStatus as a string; the reviewed page does not specify IndexDocument.Type. | **SKIP**: HTTP 404 / NoSuchWebsiteConfiguration. Numeric XML text alone cannot determine a Go integer contract. |
| ListBuckets | [ListBuckets](https://help.aliyun.com/en/oss/developer-reference/listbuckets) defines ListAllMyBucketsResult, capitalized fields and Buckets/Bucket wrapping. | **PASS for wrapping/case**: CLI and raw SDK XML agree. One bucket was returned; pagination fields were absent. Live continuation remains NOT RUN. Service endpoint/header lowering still needs implementation. |

## Explorer metadata findings

- Fetch the definitions through `https://api.aliyun.com/meta/v1/products/Oss/versions/2019-05-17/apis/<Operation>/api.json?language=EN_US`. Exact URLs, hashes and JSON pointers are in the evidence record. No browser login is needed for this public metadata interface.
- GetBucketCors: AllowedHeader is declared `type: string` with `items.type: string` and `items.extendType: "true"`. This is an inconsistent/extended representation; it is not a plain array declaration. Normalize and review it before approving a cardinality override.
- GetBucketInfo: inline BucketPolicy properties are LogBucket/LogPrefix, while its `$ref` points to LoggingEnabled with TargetBucket/TargetPrefix/LoggingRole. Preserve both facts; raw XML and documentation support the inline names in this response.
- GetBucketReplicationLocation: TransferTypes.Type is an array of strings. Raw XML confirms repetition.
- GetBucketReplicationProgress: Rule is an array in metadata. No configured rule was available for live verification.
- Both inventory operations refer to SSEOSS declared as `type: string` with empty `properties`. Documentation instead describes an empty container. The live mapping remains unverified.
- GetBucketWebsite: metadata declares HttpStatus and Type as integers; DSL declares strings, and the help page labels HttpStatus a string. Preserve this three-source conflict.
- ListBuckets: metadata agrees with the observed capitalized root and Buckets/Bucket wrapper. It is not evidence for complete bucket field coverage or pagination continuation.
- Metadata can corroborate or expose conflicts. It does not replace the complete official DSL/parser route or authorize silent model changes.

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
| GetBucketCors | [Open](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketCors) | Bucket with a multi-header CORS rule |
| GetBucketInfo | [Open](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketInfo) | Bucket with logging |
| GetBucketInventory | [Open](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketInventory) | Bucket and SSE-OSS inventory ID |
| ListBucketInventory | [Open](https://api.aliyun.com/api/Oss/2019-05-17/ListBucketInventory) | Bucket containing an SSE-OSS inventory |
| GetBucketReplicationLocation | [Open](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketReplicationLocation) | Bucket |
| GetBucketReplicationProgress | [Open](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketReplicationProgress) | Bucket and replication rule ID |
| GetBucketWebsite | [Open](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketWebsite) | Bucket with website configuration |
| ListBuckets | [Open](https://api.aliyun.com/api/Oss/2019-05-17/ListBuckets) | Account with existing buckets |

## Decision gate

- These findings identify candidate corrections. They do not approve runtime behavior or unblock generation.
- Any correction needs a source-bound decision, paired rationale and independent offline XML fixtures before regeneration.
- Preserve the complete DSL and original source bytes. Model overrides and serialization mappings must be explicit and reusable.
- The other 66 unsupported operations are currently blocked by method/body profiles, dynamic paths, missing body targets or service-level binding. Explorer evidence does not implement those generator capabilities.

## Verification

- This branch adds documentation and redacted evidence only. Run the documentation language/local-link check and `git diff --check`.
- Code verification for the signer belongs to #126. Raw private responses remain outside tracked files. No configuration was created or changed to obtain missing evidence.
