# Internal OSS V4 signing

[中文](oss4-signing.zh-CN.md)

- Issue #92, internal signing stage after bounded response streams and FC binary generation.
- Follow the [official V4 specification](https://www.alibabacloud.com/help/en/oss/developer-reference/recommend-to-use-signature-version-4) and reviewed GatewayOSS 0.0.42 source. This is an independent standard-library implementation; do not copy the native SDK or add runtime dependencies.
- Use explicit bucket, region and time. Canonical resource paths include the bucket and preserve object-key slashes. Sort encoded query names; empty values use a bare subresource name. Do not reuse ACS3 framing.
- Sign native content-type/content-md5 and x-oss headers, including managed date/token/UNSIGNED-PAYLOAD. Optional additional headers require explicit registration. Preserve request bodies without reading or closing them.
- Normalize names, reject ambiguous repeated headers/query keys, invalid path/query bytes, injection, incomplete scope and mixed ACS3 inputs. Validate on owned copies; failures do not mutate requests.
- Acceptance: the official signing-key/canonical-digest/signature vector, independent synthetic path/query/header/STS/body-ownership fixtures, cancellation and no-mutation rejection, deterministic external Example, paired docs and Go/CI gates.
- This stage adds no public operation/authentication mode, OSS endpoint builder or generator. XML trait/IR/Gateway integration, request replay/checksum behavior and live acceptance remain pending. No cloud calls or resources.

## Verification

- Implementation and final-head CI pending.
