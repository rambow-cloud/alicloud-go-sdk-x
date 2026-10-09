# FC binary generation

[中文](fc-binary-generation.zh-CN.md)

- Issue: #92. Build on the shared [response stream contract](response-streaming.md).
- Use the pinned complete FC DSL, official semantic parser and the existing batch backend. Recognize the binary program shape; never special-case an operation name.
- Review its header-model merge, guarded string headers, exact path/query, identical body/stream handoff and guarded response conversions. Unknown transforms or trailing statements remain unsupported before writes.
- Preserve native request bytes. The operation input maps DSL `readable` to copied `[]byte` bounded at 8 MiB; this intentional v0 API choice supports signing and replay without caller-reader ownership. It does not claim official source compatibility or unbounded uploads.
- Preserve declared header models and native wire names. Common headers are copied; typed headers override the same native key, matching the DSL. SDK-managed/invalid headers still fail before credentials.
- Output `Body` is an owned `io.ReadCloser`; expose native headers/status and Metadata. The caller closes it; limits, timeout and errors follow the shared runtime. Never retry an invocation or late read under the default conservative policy.
- Pinned OpenApi 0.3.23 lines 901–905 prioritize request.stream, hash its bytes and use application/octet-stream. Binary response handling returns the response body without JSON decoding.
- Verify exact bytes/hash/path/header presence, caller ownership, lazy read/cancel/limits, structured errors and a renamed product/action. Include a generated deterministic Example and no-write rejection fixtures.
- Run Node 22 frontend tests/check before generation and Go doccheck/vet/tests/formatting. Bump the IR schema/profile explicitly for binary types and regenerate owned outputs; preserve source and license bytes.
- No live function invocation is authorized by the existing-resource read-only scope. Report generation, compilation and offline acceptance separately.

## Status

- Route recorded before implementation. Production binary generation remains pending.
