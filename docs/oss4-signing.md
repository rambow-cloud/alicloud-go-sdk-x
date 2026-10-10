# Internal OSS V4 signing

[中文](oss4-signing.zh-CN.md)

## Temporary credential correction (#126)

- Native OAuth verification found temporary AccessKey IDs with the `STS.` prefix. The original region validator rejected their period before transport.
- Validate credential identifiers separately: permit ASCII letters, digits, hyphen, underscore and period. Reject whitespace, controls, non-ASCII text and scope/header separators. Region scope validation remains unchanged.
- Verify the existing independent STS vector with a synthetic dotted identifier. Add runtime temporary-provider integration and invalid-identifier rejection fixtures. Use a dotted synthetic identifier in the existing account-free OSS Example.
- Authorized live checks use the existing native temporary profile and GET requests only. No credentials, resource identifiers or bodies are published. SDK wire evidence does not imply generated `service/oss` acceptance.
- Run doccheck, vet, full Go tests and formatting after implementation; final-head CI is required before merge.

- Local verification on 2026-10-10: root runtime tests, the independent STS signature vector, rejection fixtures, deterministic Example, doccheck (17 packages), vet and formatting passed. The initial full Go run failed only two codegen fixtures because the fresh worktree lacked the pinned Node parser. After installing lockfile dependencies with npm ci, both fixtures passed (37.426 seconds); all other packages passed in the original run.
- Read-only live verification after the fix: ListBuckets, GetBucketInfo and GetBucketReplicationLocation each returned HTTP 200, application/xml and an OSS request ID in one attempt through the native oss-sftp temporary profile. Before the fix all three failed before HTTP with the invalid OSS V4 inputs error. Raw response structures agree with official CLI results. Public OSS generation and browser UI checks remain separate.

- Issue #92, internal signing stage after bounded response streams and FC binary generation.
- Follow the [official V4 specification](https://www.alibabacloud.com/help/en/oss/developer-reference/recommend-to-use-signature-version-4) and reviewed GatewayOSS 0.0.42 source. This is an independent standard-library implementation; do not copy the native SDK or add runtime dependencies.
- Use explicit bucket, region and time. Canonical resource paths include the bucket and preserve object-key slashes. Sort encoded query names; empty values use a bare subresource name. Do not reuse ACS3 framing.
- Sign native content-type/content-md5 and x-oss headers, including managed date/token/UNSIGNED-PAYLOAD. Optional additional headers require explicit registration. Preserve request bodies without reading or closing them.
- Normalize names, reject ambiguous repeated headers/query keys, invalid path/query bytes, injection, incomplete scope and mixed ACS3 inputs. Validate on owned copies; failures do not mutate requests.
- Acceptance: the official signing-key/canonical-digest/signature vector, independent synthetic path/query/header/STS/body-ownership fixtures, cancellation and no-mutation rejection, deterministic external Example, paired docs and Go/CI gates.
- This stage adds no public operation/authentication mode, OSS endpoint builder or generator. XML trait/IR/Gateway integration, request replay/checksum behavior and live acceptance remain pending. No cloud calls or resources.

## Verification

- Node 22 frontend/IR/prose check passed. Go 1.27.1/Windows doccheck, full vet and root tests passed (codegen 118.271 seconds).
- Three complete-input native vectors and the document's canonical digest/provided-key MAC passed. Body ownership, STS replacement, Unicode/slashes, empty subresources, additional headers and safe rejection have offline fixtures.
- Final review checked native path normalization and tightened hop-header validation; affected signing tests cover these corrections. Go-created Unicode URL hints are normalized, while mismatched decoded/raw paths fail. Final-head Linux race/Windows CI remains pending. No runtime dependency or cloud calls.

## Source decisions

- [Fixed vector facts](../metadata/oss4-signature-evidence.json) bind GatewayOSS archive/source hashes and the native SDK revision/files. Only synthetic vector parameters and expected signatures are reused.
- The document's canonical digest and MAC using its supplied derived key match. Deriving a key from the displayed `yourAccessKeySecret` gives a different key. This displayed value cannot establish end-to-end vector acceptance; preserve the discrepancy and use complete native basic/STS/additional-header vectors for derivation.
- Request headers are normalized to lowercase for canonicalization. Authorization uses Gateway-style comma spacing; the native test signature values match independently of that formatting.
- Duplicate scalar headers/query keys, presigned auth query fields and unexpanded `x-oss-meta-*` are rejected. Gateway/body preparation must resolve those transformations before signing. Additional headers must be present, unique and outside the mandatory set; no inferred registration.
- `UNSIGNED-PAYLOAD` authenticates the reviewed OSS V4 request shape, not body integrity. XML MD5 and streaming CRC/replay remain separate work.

## Internal contract

- `SignOSS4(ctx, request, credentials, options)` accepts an owned HTTPS virtual-host/service request. Supply exact bucket/region and nonzero time; no discovery occurs. The URL path excludes the virtual-host bucket.
- Success publishes cloned URL/headers. Other request fields, Body and GetBody are unchanged. Failure/cancellation changes no request fields; expiry remains identifiable with `errors.Is`.
- Header/query values are scalar in this reviewed profile. Native content-type/content-md5/x-oss headers sign automatically. Optional names require explicit selection; hop headers are unsupported. Host uses the effective Request.Host/URL.Host; signed content-length requires a positive known body length and no transfer encoding.
- V1/V2, presigned URLs, CloudBox, endpoint construction and checksum calculation are unsupported. This internal helper has no shared mutable state; callers must own requests during signing.
- Run `go test ./internal/signing` for synthetic vectors and the deterministic external Example. Tests send no HTTP requests and print no authorization/credential values.
