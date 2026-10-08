# [Feature]: Implement a shared staged middleware pipeline

### Affected areas

- middleware, core

### Dependencies

- None

### Problem and scope

- Complete the shared runtime foundation before product generation.
- This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Provide Initialize, Build, Finalize and Deserialize stages with named middleware and function adapters.
- [ ] Separate once-per-operation hooks from once-per-attempt hooks; freeze client registrations.
- [ ] Preserve context, short-circuit errors and deterministic ordering; reject duplicate IDs per stage.
- [ ] Test ordering, retries, cancellation and concurrent client use.

### Documentation and verification

- English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.
