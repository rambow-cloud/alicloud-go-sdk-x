# Native XML root discovery

[中文](native-xml-traits.zh-CN.md)

- Issue #92; follows the shared XML codec and fixed OSS/helper comparison.
- Keep complete official DSL and the semantic parser authoritative for operation/model discovery. Native helper declarations supply additional serialization facts; they must not replace DSL schemas or select hand-authored models.
- Add a standard-library Go AST reader for complete helper registry/model files. Read source as data; never execute it. Bind inputs to exact module version, revision, checksum and file hashes.
- Discover every registered action and explicit XML root tag automatically. Distinguish structured roots from scalar roots. Preserve source coordinates, XML/JSON names and native type facts; never derive roots from action/type names.
- Reject malformed/dynamic/duplicate registry bindings and ambiguous root declarations. Unsupported model shapes retain explicit reasons. Extraction returns complete results or an error, without writing SDK files.
- Compare the inventory against all 79 XML response declarations in the complete pinned OSS DSL. Normalize structured wrappers before field comparison; record missing or divergent roots. This is root discovery, not complete nested-model validation or generated OSS acceptance.
- Acceptance: independent renamed synthetic registry/model fixtures, source-hash rejection, cancellation, deterministic JSON inventory/Example and the complete fixed-corpus comparison. Run frontend/check, applicable Go/format/docs and exact-head CI gates.
- Reuse verified source already downloaded by the isolated tools/ossxml module. No native implementation/prose is copied into the runtime, no dependencies are added and no cloud calls occur. Production pin/IR/Gateway/signing/namespace/checksum integration remains separate.

## Verification

- Implementation and fixed-corpus comparison pending.
