# OSS semantic IR

[中文](oss-semantic-ir.zh-CN.md)

- Issue: #92. This stage follows the merged [OSS runtime](oss-runtime.md).
- Pin the complete official OSS DSL and imports at the existing source revision. Preserve source and notice bytes.
- Use the official semantic parser. Discover every operation and reachable model; do not maintain an operation allowlist.
- Recognize the exact Gateway initializer, bucket host map, query guards, protocol constants and execute handoff.
- Compare XML response roots and nested field names, types and cardinality against hash-bound native Go AST facts. Report differences with both source coordinates; never silently fix them.
- Produce reusable operation IR for bodyless XML reads with reviewed root semantics. Retain all other operations with deterministic reasons.
- This is a frontend acceptance stage. Its staged OSS source is not in the production product list until the Go backend consumes these traits. Four existing products must retain their owned outputs.
- Next: generic Go XML emission, renamed-product compile/runtime proof, service/oss package docs and offline Examples. Then reviewed request-body, pagination and integrity policies. No cloud acceptance is claimed here.

## Acceptance

- Full official OSS discovery and deterministic semantic projection run without network access.
- A renamed synthetic product uses the same lowerer and emits the same protocol traits without OSS operation names.
- Mutation tests reject changed Gateway initialization, host/query/body/handoff programs and nested XML differences.
- Unknown roots, wrapper differences, unsupported bodies and unsafe shapes have explicit reasons. Selected unsupported operations fail before output writes.
- Node 22 frontend checks/tests precede Go generation checks, doccheck, vet, tests and formatting. CI checks the final PR head.
- Paired docs distinguish discovery, lowering, Go emission, compilation and live acceptance.

## Offline workflow and current scope

- `node tools/darabonba/oss-discover.cjs generate` writes the staged IR and coverage under `docs/research/`. `check` verifies byte equality; `report` prints coverage.
- Append operation names to assert support, for example `node tools/darabonba/oss-discover.cjs generate GetBucketAcl`. Selecting GetBucketCors or an unknown action fails before either report is written.
- The source lock uses `stagedProducts.oss`; production `products` remains STS/ECS/VPC/FC. Native facts, registry/model hashes, Gateway import/hash and complete product hash are pinned in `metadata/oss-semantic-pins.json`.
- Current result: 90 discovered, 303 declared models, 16 lowered, 74 unsupported. Go emission, compilation and live coverage are not assessed for OSS. The existing Go backend rejects this staged XML IR.
- Lowered reads include GetBucketAcl, GetBucketLocation, ListObjects and ListObjectsV2. Native roots preserve local-name namespace matching; scalar root text has an explicit field binding. Static subresources are separate from pathname and input queries.
- Supported input facades retain bucket, headers and native query fields. Output facades retain body fields plus transport metadata; HTTP header/status facade fields are intentionally separate from body schema. Retry, paginators, waiters and checksums are not inferred from names or fields.
- Current boundaries: GET/XML, no request body, one explicit bucket parameter, static path/subresources, scalar query values and map headers. Structured root wrappers, dynamic paths, service-level routing, typed header programs, other bodies/methods and shape differences retain reasons.
- Existing production sources and their local import maps are unchanged. Adding staged source changes the corpus digest; policy/translation bindings and generated source indexes are refreshed. Historical cloud/release evidence keeps its original pins.

## Source differences

- ListBuckets has lowercase DSL body members and a flat bucket array, while native XML uses capitalized members and a Buckets/Bucket wrapper.
- CORSRule.AllowedHeader is a scalar in this DSL and repeated in the pinned native model.
- New DSL fields absent from the older native helper remain in discovery. They must not be discarded to make shapes match.
- Source comparison is static evidence. Explorer browser or real cloud evidence must be recorded separately before approving corrections.

## Local verification (2026-10-10)

- Node 22.21.1: frontend/production/staged checks passed. Initial full suite: 115/117 passed; module order and STS rehearsal registrations failed. Fixed bytewise module order, staged import maps and narrowed rehearsal registration; all five importer tests and four rehearsal tests passed.
- Final scoped OSS suite: seven passed, one Windows symlink test skipped. Linux CI runs that test. Renamed-product, nested differences, mutation and no-write cases passed.
- Go 1.27.1: sdkgen check/product-check, doccheck (17 public packages), vet, full tests and formatting passed. Complete codegen tests passed in 127.800 s, including rejection of staged XML IR.
- Automation: 26 tests passed. Language/local links passed before the final paired verification-note update; final-head CI records the final documentation result.
- Corpus digest: `145082e7c8d725925ed509b430a6cc2e302d3f2fc97ae4986ea8f114fbc3924c`. Existing service Go APIs and runtime behavior did not change.
- PR/issue records hold final-head CI and merge evidence. No OSS cloud call, public client or release/indexing acceptance occurred.
