# [Feature]: Import pinned OpenAPI protocol metadata and validate generator IR

- GitHub issue: #21.

# [Feature]: Import pinned OpenAPI protocol metadata and validate generator IR

### Problem and evidence

- Foundation #19 passed.
- The four reference operations remain handwritten.
- Official source: https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/.
- Development path: docs/generator.md; parent #8.

### Scope

- Implement a bounded explicit metadata importer, protocol-only snapshots, source/raw/snapshot SHA-256 provenance, strict manifests/overlays and validated RPC IR.
- Preserve selected public APIs through reviewed overlays; reject removed or changed selected fields, unsupported encoding/references and newly required inputs.
- Generation must be offline and independent of Go map iteration order.

### Dependencies

- #19

### Acceptance criteria

- [ ] Offline importer HTTP fixtures, checksum tampering, schema drift, strict overlay errors and a non-ECS synthetic operation. Pin four official ECS/STS operations; supply paired design/usage docs.
- [ ] Original English-primary code comments and equivalent English/Chinese Markdown updated together.
- [ ] Record exact commit and relevant successful verification before closure.

### Affected areas

- tools
