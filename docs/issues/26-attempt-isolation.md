# [Bug]: Isolate attempt outputs and retry interrupted idempotent response reads

- GitHub issue: #26.

### Problem and evidence

- Review of abcf961 found behavior defects and incomplete AWS Go SDK v2 design alignment.
- Design and acceptance were written before code in docs/aws-style-remediation.md.

### Scope and dependencies

- Clear and commit decoded state per successful attempt.
- Reject successful short circuits without an output.
- Retry EOF/unexpected EOF and explicitly transient body reads only under the bounded idempotency policy.
- Preserve cancellation, JSON failures, atomic output and safe diagnostics.

- Dependency: #25 (completed).

### Acceptance criteria

- [ ] Offline regressions reproduce stale middleware/retry output, incomplete successful short circuits, interrupted body retry/success, no-retry/non-idempotent/canceled/JSON failures, closure of response bodies and fresh metadata.
- [ ] English-primary Go docs, deterministic offline Examples, equivalent bilingual docs and migration notes; offline regeneration, doccheck, language, vet, tests, Linux race and Windows CI.
- [ ] Record exact commit and CI evidence before closing.

### Affected areas

- core, transport
