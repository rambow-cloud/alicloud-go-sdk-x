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

- [Complete comparison](research/oss-native-xml-roots.json): 90 DSL operations, 79 XML response declarations, 683 native models and 82 explicit native roots. 36 response bodies match root-level wire fields; 42 DSL response facades declare no body field; one (`ListBuckets`) has conflicting names. These counts do not prove nested types or runtime behavior.
- `ListBuckets`: DSL fields `buckets`, `isTruncated`, `marker`, `maxKeys`, `nextMarker`, `owner`, `prefix` differ from native `Buckets`, `IsTruncated`, `Marker`, `MaxKeys`, `NextMarker`, `Owner`, `Prefix`. The native `Buckets` is an additional model wrapper while DSL `buckets` is an array. Root discovery does not approve case conversion or collection flattening. Future lowering requires a separately reviewed source decision/fixture; Explorer evidence is NOT RUN.
- Native tags without a namespace URI explicitly mean local-name matching, as in native Go XML binding. Tags with an absolute URI require exact matching. The shared XML codec currently requires an exact namespace; integrating native local matching remains separate.
- Static registry profile: a global empty `make(map[string]reflect.Type)` with literal-key, empty-model `reflect.TypeOf` assignments directly in `init`. Reflect import aliases are supported. Reject duplicate/dynamic/conditional bindings, shadowed symbols, aliases, whole-map replacement, indirect mutation and model-file registry references. Unsupported root tags/types retain reasons.
- Input plan: [exact helper pins](../metadata/native-helper-pins/oss.json), module v0.0.6 at `cd82cbd16bcb3f0988e125ee63679e887867f225`. Verify original complete file hashes before extraction. No native source is vendored and no cloud API is called.
- Reproduce against the verified private complete OSS source candidate and downloaded native module (paths are explicit; the command does not fetch or generate SDK files):

```powershell
node tools/darabonba/oss-xml-roots.cjs --source-root <complete-verified-source-root> --registry <native-client.go> --models <native-structs.go>
```

- Synthetic tests cover renamed roots, scalar/structured binding, source coordinates, unsupported tags, static ambiguity, drift, cancellation, deterministic output, ownership and short writes. The account-free `ExampleParseXML` uses independent source fixtures.
- PASS: Node 22 frontend/IR/prose check and all 110 frontend tests; doccheck (17 public packages, JSON v2 and stdlib core), product-check, full vet and Go tests/Examples (codegen 125.035s), tracked Go formatting, paired-language/local-link check and diff check. Final reader reproduces the recorded fixed-corpus JSON exactly. Exact-head CI remains a separate PR gate.
