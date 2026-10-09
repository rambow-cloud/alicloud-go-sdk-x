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

- Implemented offline: schema v6 / openapi-http-v1; 73 discovered, lowered, emitted and compiled FC actions; 335 models and 73 runnable operation Examples.
- Source/license bytes and STS/ECS/VPC executable Go outputs remain unchanged. Their IR/prose hashes were regenerated for the schema change.
- Node 22: frontend check and all 105 tests passed; the additional changed-binary no-write fixture passed separately (106 covered).
- Go doccheck (17 public packages), vet, generation check, formatting and language/local-link checks passed. Initial full tests passed outside codegen; its renamed-fixture source coordinate failure was fixed. The affected codegen tests, including full isolated compilation and invalid binary IR cases, passed in 31.550 seconds.
- Reuse fixture reparses renamed product/action/header-model/request/response fields with the official parser, then compiles and runs the generated client. Invalid bindings, transforms, signature algorithms and selected source drift fail before writes.
- Field prose: 1123 of 1605 fields have upstream English; all 1605 have Go contract comments. Missing business prose remains #93.
- Final-head Linux race/Windows CI is pending. No FC live calls or new resources. XML, unbounded request streams and OSS signing remain open under #92.

## Protocol decisions

- Evidence: [machine-readable review](../metadata/binary-protocol-evidence.json), bound to the production source manifest.
- [Util v2.0.8](https://github.com/alibabacloud-go/tea-utils/blob/6bfaf5ce0320c26217c94983f97bbeac4b2435d6/service/service.go#L264) returns string headers unchanged. JSON quoting would change the wire value.
- [Tea v1.2.2](https://github.com/alibabacloud-go/tea/blob/bb7ae5fd3e46dbac00db758063074c89b93cba26/tea/tea.go#L406) provides the reviewed scalar response-header representation: lowercase names and first values. The generated output preserves it; the low-level StreamingOutput retains all http.Header values.
- Native source hashes and exact revisions are recorded. No native Go implementation or runtime dependency is copied. These facts do not establish all-release or live parity.
