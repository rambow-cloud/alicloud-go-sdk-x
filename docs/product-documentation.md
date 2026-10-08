# Licensed product documentation

[中文](product-documentation.zh-CN.md)

- Stage #38 follows #37 / PR #42 on `issue/38-licensed-product-docs`, stacked on `issue/37-capability-policies`.
- The original dependency stack is now [integrated into main](generator-integration.md).
- This specification preceded implementation.
- Keep the existing supported RPC profile and operation/model counts; broader protocols require separate scoped work.

- Use the official semantic parser's field description strings and operation annotation tokens from checksum-pinned product DSL.
- Extend IR with description/summary records and source coordinates; do not parse Tea source using a second compiler or require per-operation documentation overlays.
- Upstream example values remain coordinates only: never copy account IDs, credentials, resource IDs, CLI commands or policy examples into executable Go Examples.
- Existing invocation/capability Examples stay deterministic, offline and explicitly illustrative, rather than valid cloud request sets.

- Emit English-primary native Go paragraphs on operations and fields, retaining our ownership, pointer, retry and cancellation contracts.
- Convert Markdown links/emphasis, headings and lists into readable Go prose; preserve safe HTTPS links, discard HTML/ unsafe URL markup, normalize controls/newlines, and escape compiler directives.
- Every source line is emitted inside // comments; upstream prose cannot create code or become validation, requiredness, deprecation or retry policy.
- Mark descriptions as upstream service documentation distinct from SDK contracts and link to pinned source lines.
- Missing/empty/non-English-only descriptions are reported, never invented or translated.

- Generate paired English/Chinese guides with the same usage/contracts, operation/source indexes and documentation coverage.
- Semantic prose is available in English in the fixed corpus; no pinned Chinese translation is assumed.
- Both language guides point to the same licensed source/Go documentation and state this language limitation.
- Do not label English excerpts as translated Chinese.
- Retain machine-readable operation/field documentation coverage, missing descriptions and source attribution independently of emission, reviewed capability and live acceptance.

- Preserve upstream source/license bytes, attach copyright, Apache-2.0 attribution, prominent transformation notices and full license terms to product output.
- Root MIT continues to cover original runtime/tooling; upstream-derived definitions/prose are not relabeled MIT.
- No imported module implementation is emitted.
- Package notices and source-lock references make redistributed generated packages self-describing.

- Acceptance: official parser extraction with real corpus and missing/empty annotations; deterministic output, source coordinates, English prose, safe Markdown/HTML/directive handling, no promoted source examples or behavioral changes, preserved notices and accurate language/description counts.
- Existing protected-output/reconciliation tests also cover generated docs/licenses.
- Run frontend/discovery checks and relevant Node tests before Go gates; both generation checks, full Go tests including isolated product compilation/Examples, doccheck, vet, tracked formatting, bilingual/whitespace checks, then Linux race/Windows CI.
- No live calls, release, pkg.go.dev indexing or merges.

- Evidence: pinned [product corpus](https://github.com/aliyun/alibabacloud-sdk/tree/ec489e5c3deae95496daae2b41503ac58b221adb), [upstream license](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/LICENSE), and [Apache-2.0 terms](https://www.apache.org/licenses/LICENSE-2.0.txt), inspected 2026-10-07.
