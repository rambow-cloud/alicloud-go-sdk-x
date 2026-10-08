# Issue 35: complete product discovery

- Actual issue: [#35](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/35), stage of #33, following #34 normalization.
- Specification [product-discovery.md](../product-discovery.md) was committed at 3e63e39 before implementation.
- Branch issue/35-product-discovery is stacked on #39 / issue/34-source-normalization, which depends on #32.

- Deliverable: automatic complete official DSL candidate discovery and cycle-safe, identity-preserving reachable model IR.
- Protocol/binding evidence, exact members, numeric source types, document coordinates and source locks are retained.
- Optional upstream api-info is a cross-check, never a whitelist.
- No legacy snapshots/overlays/ decisions/canonical prerequisite.
- Strict selection fails before writes; check/report are read-only and all commands are offline.
- Generated artifacts are under models/.

- Verification includes complete corpus accounting/determinism, missing enrichment, unsupported/dynamic/ambiguous behavior, unknown and recursive wire types, duplicate members/catalogs, source tampering, strict selection and read-only checks.
- Run the frontend/discovery tests/checks and all documented Go/doc gates; record actual results in the linked PR.
- Linux race and Windows CI remain required.
- Model graph references are not executable DSL.
- Lowering is distinct from Go emission/compilation/live coverage; batch Go output remains #36.
- Parent #33 and unmerged dependencies stay open.
