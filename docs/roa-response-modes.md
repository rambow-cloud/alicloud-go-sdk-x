# ROA response modes

[中文](roa-response-modes.zh-CN.md)

- Issue #92, shared-runtime stage. This does not add generated FC, XML or streaming coverage.
- Evidence: pinned OpenApi 0.3.23 `main.tea`, lines 964–1036. Official error decoding accepts `Code`/`code`, `Message`/`message` and `RequestId`/`requestId`. FC 2023-03-30 includes JSON and `none` response declarations.
- Keep JSON as the zero-value response mode. Accept only explicit reviewed `none` mode for operations without a typed success body; do not infer it from HTTP status or method.
- A successful `none` response is read within the configured size limit and discarded. Publish a fresh zero-valued output and header request ID. Never require JSON or invoke the model decoder for this success case.
- Error responses still produce structured JSON service errors. Prefer an explicitly present uppercase member; otherwise use its lowercase counterpart. Keep strict JSON parsing; do not enable general case-insensitive decoding.
- Read errors, size limits, cancellation, response closure, middleware and conservative retry rules still apply. Unsupported modes fail before credential retrieval or HTTP.

## Verification plan

- Independent fixtures: empty 204, non-JSON 200, malformed/oversized/read-failing responses, error member precedence, strict duplicate JSON members and unknown mode rejection.
- Check fresh output publication, header metadata, body closure, custom decoder exclusion and context cancellation.
- Add a deterministic external Example and English package/symbol comments.
- Run formatting, doccheck, vet and Go tests; require final-head Linux race and Windows CI.
- Full official product lowering, typed body/header/path emission, XML and stream ownership remain open under #92.

## Local result

- Go 1.27.1: full package tests, vet and doccheck passed. The root tests/vet were rerun after separating success metadata from error members; JSON business fields named code/message keep their native types.
- The offline Example, size/read failures, closure, error precedence, no-request rejection and reviewed retry fixtures passed.
- Formatting and paired documentation checks passed. Final-head CI remains a merge requirement.
