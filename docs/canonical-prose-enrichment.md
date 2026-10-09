# Optional canonical field prose

[中文](canonical-prose-enrichment.zh-CN.md)

- Issue #93. Keep official Darabonba discovery, models, wire behavior and capability policy authoritative.
- Missing field business prose can use licensed English descriptions from the pinned `aliyun/aliyun-openapi-meta` revision `51286a65c79d008436eb314e636f9c9ad4b1ca08`.
- The metadata repository targets CLI builds and declares an unstable structure. Its prose is informational, not SDK validation or runtime policy.
- Establish this plan before source import or implementation.

## Route

- Fetch matching STS/ECS/VPC metadata automatically from the existing complete IR. Preserve source/license bytes and hashes under `sources/openapi-meta/prose/`; do not modify existing normalization fixtures.
- Run representation normalization first: exact wire names, indexed inputs and itemName response wrappers. Record unsupported metadata and missing mappings with deterministic reasons.
- Match each candidate to the complete DSL model field and native wire type. Add descriptions only where upstream English prose is missing; never override existing prose or alter models.
- Produce deterministic source-bound prose projections and coverage with pinned file hashes and JSON pointers. No per-API authoring prerequisite or handwritten field selection.
- Consume projections as optional documentation enrichment. Absence must preserve complete service generation. Drift, invalid field references or modified sources fail before writes.
- Generate English comments, paired usage/source coverage and Apache attribution. Do not turn upstream account/resource examples into runnable tests.

## Acceptance

- Normalization/mapping fixtures cover arrays, wrappers, indexed inputs, ambiguous conflicts and unsupported shapes.
- Verify exact source text/hash, missing languages and coverage reasons. Conflicting mappings are rejected or explicitly excluded; no guessed descriptions.
- Prove executable operations/model definitions/capability policy remain unchanged; comments and documentation may change.
- Optional-corpus absence, source drift and invalid projections have no partial writes.
- Node 22 projection check/tests, sdkgen check/product-check, doccheck, vet, formatting, Go tests and final-head CI.
- Report actual additional fields and remaining gaps after generation, rather than promising full business-prose coverage.

## Current normalized corpus

- Imported 778 matching metadata files; 9 ECS actions are absent at the pinned revision. Original source bytes total about 7.54 MB.
- Accepted additional English field prose: STS 11, ECS 253, VPC 11. Remaining fields without original DSL or accepted metadata English: STS 12, ECS 5139, VPC 3970. Go contracts still cover every exported field.
- STS: all 4 metadata operations normalize. ECS: 352 normalize, 19 have unsupported representations and 9 are absent. VPC: 399 normalize, 4 have unsupported representations.
- Exact-path/type differences and conflicting prose stay excluded in `metadata/prose/<product>.json`; they do not alter SDK requests. Source examples and backend annotations are not promoted to runtime policy.
- Regenerate after IR changes with `npm --prefix tools/darabonba run discover`, or `node tools/darabonba/prose.cjs generate` after standalone discovery. `npm run check` verifies prose projections offline.
- Projections bind the complete IR hash, corpus manifest hash, original file hash, native field source, text hash and JSON pointer. Generated notices retain the additional Apache source attribution.

## Verification

- Node 22.21.1: frontend/prose checks and all 92 tests passed.
- sdkgen check/product-check, doccheck for 16 public packages, vet, formatting and paired language/link checks passed.
- All Go packages passed; the full generator suite completed in 102.383 seconds.
- Independent tests compare model ASTs and every non-documentation artifact with enrichment disabled. Executable operations, native models and capabilities are unchanged.
- Source drift, unknown fields, altered source text and unlisted source files fail before changing existing generated output. Indexed/wrapper fixtures, conflicting prose and corpus absence pass.
- Final-head Linux race and Windows CI are required before merge. #93 stays open for remaining business-prose gaps.
