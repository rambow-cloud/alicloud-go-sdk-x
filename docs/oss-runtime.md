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
