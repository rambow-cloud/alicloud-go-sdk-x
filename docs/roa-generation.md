# ROA generation stages

[中文](roa-generation.zh-CN.md)

- Issue: #92, partial delivery. RPC generation stays supported.
- First implement and verify exact escaped resource paths in the shared runtime.
- Then lower official ROA parameters, headers and JSON bodies into complete IR and emit typed calls.
- [Explicit response modes](roa-response-modes.md) and [FC JSON/none generation](fc-roa-product.md) extend the path stage. XML, streaming ownership and additional signing profiles remain separate #92 stages.

## Path contract

- `Request.Path` is decoded. Optional `Request.RawPath` preserves escapes inside a parameter segment.
- RawPath must be absolute, have valid escapes and decode exactly to Path. Invalid pairs fail before credentials or transport.
- Signing and HTTP transport use the same canonical escaped path. An encoded slash stays inside its parameter segment.
- The internal path builder accepts named placeholders and an exact parameter set. It escapes each value as a single RFC3986 segment; it does not clean paths or reinterpret values as URLs.
- Input maps are read without mutation. Empty/missing parameters, malformed templates, invalid UTF-8 and canceled contexts fail.
- Deterministic offline examples and wire tests verify spaces, slashes, percent signs and Unicode, plus rejection without transport.

## Official product research

- FC source: [pinned complete DSL](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/fc-20230330/main.tea).
- It uses multiple path parameters, a separate headers map, request-model JSON bodies, shared response models and `none` response bodies.
- `ap-southeast-5` contains trailing whitespace in the pinned endpoint map. Keep upstream bytes unchanged. Record a source-bound normalization decision before accepting this entry.
- FC discovery retains 73 operations and emits 72 offline; binary InvokeFunction remains unsupported. No live FC acceptance.
