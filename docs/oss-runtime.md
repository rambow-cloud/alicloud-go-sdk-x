# OSS shared runtime

[中文](oss-runtime.zh-CN.md)

- Issue #92. Connect the accepted XML/OSS4 helpers to shared HTTP execution before product emission.
- Evidence: GatewayOSS 0.0.42 main.tea SHA256 `71caf417a396688b8a4a38b0e0f431d5adfb7ae043d98fa504f07ac1b319f44c`: XML request/MD5 lines 92–99, path/query split 125–146, XML errors 153–181, bucket host 305–334. The [signer](oss4-signing.md) and [XML codec](xml-model-codec.md) retain their evidence.

## Scope

- Explicit OSS4 authentication and buffered XML success mode. XML success requires a typed operation codec; no JSON fallback or inferred root.
- Bucket identity is separate from the decoded object path. Resolve a regional service origin, then prepend a validated bucket once. Empty bucket means a service-level request.
- Require an explicit region and an HTTPS DNS origin. Reject IP origins, already-prefixed bucket origins and mismatched middleware authorities. CNAME, path-style, access-point, CloudBox and user-ID gateway addressing remain unsupported.
- The generator must put DSL subresources in Request.Query. An OSS Path containing `?` or `#` requires an explicit matching RawPath encoding; never interpret object bytes as query syntax.
- Reviewed nonempty XML request bytes get application/xml and a fresh Content-MD5 after Finalize middleware on each attempt. Requests remain owned buffers bounded to eight MiB.
- Reuse providers, signing, middleware, opt-in retry, timeout, limits and response-stream ownership. No unbounded uploads or automatic response integrity checks.
- Bounded Error-root XML exposes RequestId/EC with OSS header fallbacks. Empty errors use the HTTP status as a string code; malformed errors use InvalidErrorResponse. Default error text excludes message, EC and body bytes.
- XML namespace matching remains exact. Native local-name matching and ListBuckets case/wrapper normalization need reviewed generator policy.

## Acceptance established before code

- Independent offline HTTP fixtures: bucket/service hosts, escaped object paths/subresources, STS headers, absence of ACS headers, post-middleware XML MD5, typed output, malformed XML/errors, response bounds and ownership.
- Retry fixtures: no stale output publication, credentials/digest refreshed per attempt, arbitrary writes remain non-idempotent.
- Invalid mode/bucket/region/origin/authority fails before transport. Invalid initial bucket/region/origin fails before credentials.
- Cancellation remains errors.Is-compatible; response streams are lazy and close once. Existing ACS3/anonymous fixtures remain accepted.
- Deterministic external Example, English exported comments and paired guides.
- Run Node 22 frontend check/tests, sdkgen check/product-check, formatting, doccheck, vet and Go tests once after implementation; review final diff and final-head CI.
- No production OSS registration, IR change, service/oss generation, ListBuckets approval or live OSS call in this stage. Keep #92 open.

## Next stage

- Pin the complete OSS product/import closure in production; project supported Gateway initialization, hostMap and exact XML facts into IR.
- Emit service/oss through the generic backend with deterministic unsupported reasons and renamed synthetic-product compilation proof. Resolve ListBuckets only from reviewed wire evidence.

## Usage and defaults

- Set Operation.Authentication to AuthenticationOSS4 and supply Config.Region plus an unprefixed BaseEndpoint or a custom endpoint resolver. No OSS default endpoint catalog is registered yet.
- Request.Bucket selects virtual-host routing. Put object bytes in Path/RawPath and subresource keys in Query.
- RequestBodyXML treats supplied bytes as already encoded; the operation codec owns root/type validation. ResponseBodyXML requires InvokeModel and Codec.Decode; the built-in internal XML codec validates exact expanded root names.
- See the account-free [runtime Example](../oss_example_test.go). Its small explicit codec demonstrates the extension seam; production generation will use the shared validated XML codec.
- Existing defaults remain: no automatic retries, thirty-second total timeout, eight-MiB request/response bounds. Custom response bounds do not enlarge the internal XML codec's eight-MiB document limit.
- CNAME/path-style addressing, presigning, V1/V2, unbounded upload readers, response CRC/MD5 policy and cloud acceptance are not implemented by this stage.

## Local verification (2026-10-10)

- Node 22.21.1: frontend check passed; 110 frontend tests passed. Production discovery remains four products, without OSS.
- Go 1.27.1: formatting, sdkgen check/product-check, doccheck (17 public packages), vet and language/link policy (263 project files) passed.
- Initial full Go test run: all packages except internal/codegen passed. Its isolated compiler fixture omitted the new checksum/XML dependencies. Fixed both fixture and STS rehearsal dependency copies; four affected frontend rehearsal tests and the complete internal/codegen suite passed (125.278 s).
- Review added final-host DNS length validation and mixed-case managed-header coverage; affected root vet/tests passed (3.189 s). No cloud calls or source/IR regeneration occurred.
- Final-head CI and merge evidence belong to the linked PR/issue; local verification does not claim them.
