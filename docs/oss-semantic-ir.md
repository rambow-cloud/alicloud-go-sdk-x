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

## Source differences

- ListBuckets has lowercase DSL body members and a flat bucket array, while native XML uses capitalized members and a Buckets/Bucket wrapper.
- CORSRule.AllowedHeader is a scalar in this DSL and repeated in the pinned native model.
- New DSL fields absent from the older native helper remain in discovery. They must not be discarded to make shapes match.
- Source comparison is static evidence. Explorer browser or real cloud evidence must be recorded separately before approving corrections.
