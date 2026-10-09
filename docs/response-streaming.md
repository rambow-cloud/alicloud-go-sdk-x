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

## Usage

- Configure a reviewed operation with `ResponseBody: alicloud.ResponseBodyStream`.
- For `Invoke`, pass `*alicloud.StreamingOutput`. Check the call error before reading or closing `Body`.
- Read the body once and handle the read error separately from the call error. Use `defer output.Body.Close()` after a successful call.
- `Headers` is a copied map. `StatusCode` and returned `Metadata` describe the initial successful response, not completion of the read.
- For `InvokeModel`, provide `Encode` and `DecodeStream`; `Decode` is not required in stream mode. The output reader field may have any name.
- Reusing an output does not close any older caller-owned reader. Close the previous stream first.
- Run [ExampleStreamingOutput](../stream_test.go) with `go test . -run ExampleStreamingOutput`. It uses an injected local transport and no account.

## Verification

- Local Go 1.27.1 on Windows: full package tests, documentation check (17 public packages), vet, formatting and paired-language/local-link checks passed.
- Fixtures cover lazy binary reads, exact/overflow limits, EOF, cancellation without reading, blocked reads, timeout, concurrent close, read/close error redaction, structured errors, retry before publication and codec/middleware ownership failures.
- Full generation/isolated compilation tests passed in 112.468 seconds. After adding close-error coverage and increasing deadline-test slack, affected tests passed.
- Final-head Linux race and Windows CI remain pending. No live binary API call was executed.
