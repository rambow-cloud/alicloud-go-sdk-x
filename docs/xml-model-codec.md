# Shared XML model codec

[中文](xml-model-codec.zh-CN.md)

- Issue #92, next internal runtime stage after PR #115/#116.
- Reuse the [pinned OSS comparison](../tools/ossxml/README.md) and [complete-source discovery](oss-xml-protocol.md). Complete DSL fields remain authoritative. Native helper declarations supply additional XML root evidence.
- Add an internal standard-library codec. Bind exact root names and scalar-root fields explicitly; never derive them from action names. Use native JSON wire tags for fields, preserving generated DSL models.
- Decode ACL/CORS structured roots without retaining an unwanted wrapper. Preserve LocationConstraint as a scalar-root field. Test renamed synthetic models to prove reuse across products.
- Preserve optional pointers, explicit empty values and repeated elements. Ignore unknown response children for forward compatibility. Reject duplicate singular fields, malformed documents, wrong roots, unsupported types and ambiguous bindings.
- Require one document, UTF-8, bounded bytes/depth/elements and context checks. Reject directives, unsupported processing instructions and mixed model content. Return redacted codec errors; do not publish partially decoded output.
- Encode deterministic, escaped XML from a reviewed root and typed model. Reject unsupported/cyclic models and oversized output before returning bytes. No account, network, credential or cloud-resource calls.
- Acceptance: independent synthetic request/response fixtures, a deterministic external Example, root Go doccheck/vet/tests, formatting, paired documentation and exact-head Linux race/Windows CI.
- This stage adds no public operation, OSS signer or XML generator. Production OSS integration still requires pinned source/helper traits in IR, endpoint/subresource projection, OSS signing and stream/checksum contracts. Do not claim generated OSS or full #92 acceptance.

## Verification

- Node 22 official frontend/complete IR/prose checks passed for the unchanged source corpus.
- Go doccheck (17 public packages), full root tests (codegen 120.640 seconds), formatting and local links passed on Go 1.27.1/Windows.
- Vet initially rejected the intentional duplicate-tag fixture. Dynamic construction preserves that negative test; full vet then passed. The affected XML tests passed after fixture/declaration validation corrections.
- Final-head Linux race/Windows CI remains pending. No generated outputs, source pins, dependencies or live results changed.

## Internal contract

- `internal/xmlmodel.Root` supplies an exact `xml.Name`; `ScalarField` selects a scalar-root field. It is not a public SDK API or a permanent action table.
- `Decode(ctx, bytes, root, &model)` publishes a fresh output only after complete success. `Encode(ctx, root, model)` returns independently owned bytes or no bytes on failure.
- Both use native `json` wire names. Supported fields: strings, booleans, signed integers, finite floats, optional scalar/struct pointers and repeated scalar/struct elements. Maps, bytes, embedded/recursive models, duplicate names and pointer-to-slice shapes are rejected.
- Limits: 8 MiB, 64 element levels and 65,536 elements. Namespaces require absolute URIs; an optional UTF-8 BOM/XML 1.0 declaration is accepted. Namespace attributes are validated; other attributes do not populate model fields.
- Nil optional fields are omitted. Explicit empty scalars remain present. Empty arrays produce no XML elements; nil array items fail. Unknown response children are ignored, but must still form a valid bounded document.
- `ErrInvalid`/`ErrLimit` contain no raw body/value text. Cancellation remains inspectable with `errors.Is`. The codec has no shared mutable state; callers must own model/output values during use.
- Run the independent fixtures and external Example with `go test ./internal/xmlmodel`.
