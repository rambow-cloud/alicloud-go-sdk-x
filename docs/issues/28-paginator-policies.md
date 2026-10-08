# [Feature]: Align paginator options and generate multiple native pagination policies

- GitHub issue: #28.

### Problem and evidence

- Review of abcf961 found behavior defects and incomplete AWS Go SDK v2 design alignment.
- Design and acceptance were written before code in docs/aws-style-remediation.md.

### Scope and dependencies

- Return the fetched page before stopping a repeated continuation; expose a safe default and explicit duplicate-stop opt-out.
- Use overflow-safe arithmetic in both page profiles.
- Generate dedicated paginator options and per-fetch NextPage options.
- Support multiple paginator/waiter declarations and token-only/page-only/dual profiles, validating exact selected schema fields.

- Dependency: #27.

### Acceptance criteria

- [ ] Current-page delivery on repeated/cyclic tokens, empty token pages, cursor preservation on errors/cancellation, overflow boundaries, per-page overrides and input ownership. Isolated generated token-only client and multiple policies compile offline; incompatible/missing/conflicting declarations fail before writes.
- [ ] English-primary Go docs, deterministic offline Examples, equivalent bilingual docs and migration notes; offline regeneration, doccheck, language, vet, tests, Linux race and Windows CI.
- [ ] Record exact commit and CI evidence before closing.

### Affected areas

- core, tools, ecs, vpc
