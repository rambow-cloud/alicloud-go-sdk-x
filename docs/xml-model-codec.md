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

- Local checks and final-head CI are pending implementation.
