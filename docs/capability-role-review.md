# Capability role validation review

[中文](capability-role-review.zh-CN.md)

- Tracking: [#44](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/44), based on #38 / PR #43 (`issue/38-licensed-product-docs`).
- Refs roadmap #33.

- Review and [main integration](generator-integration.md) completed on 2026-10-08 in PR #45.
- The branch/dependency plan below records the pre-implementation scope.

- Review of the full-DSL PR stack found that `validateCapabilityPolicy` checks paths and scalar types independently but does not distinguish the roles sharing a model.
- For example, changing DescribeImages paginator `page` to `PageSize` leaves both `page` and `size` pointing at the same int32 field.
- The generated constructor and fetcher then overwrite page size with page number.
- A `total` pointing at `PageNumber` or `PageSize` likewise reads the wrong continuation metadata.
- Waiter `page`/`size` aliases overwrite the forced first page; `id`/`state` aliases compare identifiers as states.
- These are invalid policies even though each path exists and has the right type.
- The checked-in seven operation policies use distinct roles; this review does not claim a defect in their emitted paginators or change Alibaba wire semantics.

- Before further expansion, reject aliases within each request, response or collection member model.
- For dual paginators, request page, size and token-limit roles must be distinct; response page, size and total roles must be distinct.
- Waiter request page and size must differ, and member ID and state must differ.
- Identical paths across different models or separate adapters remain valid: a native token may have the same request/response name, and a paginator and waiter may intentionally share page fields.
- Do not invent new field names or infer behavioral roles from types.

- This is a separately tracked review fix based on the documentation stage branch; the five-stage route and its outstanding dependency reviews remain authoritative.
- Document the scope and create the actual issue before implementing regression tests or the fix.
- Verify accepted alias cases fail before rendering, including both product-generate and product-check without modifying existing owned output or deleting stale artifacts.
- Verify existing policies and legal cross-model token reuse still render, with generated SDK artifacts unchanged.
- Run frontend checks/tests, both generation checks, doccheck, vet, full Go tests and formatting once after the fix; Linux race and Windows CI provide platform acceptance.
- No live cloud calls are needed for invalid local policy rejection.

- Local verification, 2026-10-08: all eight invalid-role cases were accepted before the fix.
- After the fix, full Go tests (including isolated product compilation and Examples), official frontend/discovery checks, all 50 Node tests, both generation checks, 16-package doccheck, vet and formatting passed.
- Existing generated SDK artifacts have no drift.
- New tests verify legal cross-model/adapter reuse and read-only/failing-write preservation.
- CI and integration evidence belongs to the linked PR; local results alone do not imply Linux race acceptance or a main merge.
