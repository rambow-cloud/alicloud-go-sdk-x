# Optional field prose corpus

[中文](README.zh-CN.md)

- Issue #93; [design and acceptance](../../../docs/canonical-prose-enrichment.md).
- Pinned `aliyun/aliyun-openapi-meta` revision `51286a65c79d008436eb314e636f9c9ad4b1ca08`, Apache-2.0. Preserve source and LICENSE bytes.
- Automatic match against all 787 STS/ECS/VPC DSL actions: 778 metadata files, 9 absent at the pinned revision. Missing actions are recorded in manifest.json.
- This is optional informational prose. Complete DSL discovery, native models and runtime policy remain independent.
- The upstream structure is unstable and intended for CLI builds. Normalize indexed inputs and itemName wrappers before matching; mismatches are reported, never silently used.
- Original normalization fixtures in the parent directory remain unchanged.
- `node tools/darabonba/prose.cjs generate` writes deterministic local prose projections; `check` verifies them without network access.
- Source hashes, field coordinates, JSON pointers and excluded mappings are retained. Never use upstream account/resource examples as runnable SDK examples.
- Do not refresh the revision during ordinary generation. Import/source changes require issue review.
