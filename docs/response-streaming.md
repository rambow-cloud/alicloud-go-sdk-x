# Bounded response streaming

[中文](response-streaming.zh-CN.md)

- Issue: #92. Implement this shared runtime contract before binary DSL emission.
- Successful `ResponseBodyStream` calls return an owned `io.ReadCloser` without eagerly reading it. Low-level calls use `StreamingOutput`; model codecs attach the supplied body to an exported top-level `io.ReadCloser` field.
- The caller closes the body. EOF, read failure, overflow, cancellation or timeout also closes the transport body exactly once.
- `Config.Timeout` covers the call and subsequent reads. `MaxResponseBytes` covers delivered bytes; the defaults remain 30 seconds and 8 MiB. A shorter context deadline wins.
- Read has one owner. Close may run concurrently with Read. Preserve context and transport errors through `errors.Is`/`errors.As`; default error formatting omits raw causes.
- `Codec.DecodeStream` attaches the supplied body and copies retained response headers. It must not read, close or replace that body, or retain the response object.
- Successful middleware must preserve the owned body in the final output. Reject missing or substituted bodies before publication. Failed attempts close temporary bodies.
- Reviewed retry policy may retry before publication. Read failures after publication never start another HTTP request.
- Requests remain copied, replayable byte slices limited to 8 MiB. This stage does not add unbounded request streams.
- Verify lifecycle, exact byte limits, cancellation, error redaction, retry and codec/middleware ownership with account-free fixtures. Include a deterministic external Example, public Go docs and the usual Go gates.
- FC binary lowering and OSS XML/signing/checksums require separate parser/IR/backend acceptance. This foundation alone does not claim generated binary coverage or live acceptance.

## Status

- Planned. Verification results will be recorded with the implementation.
