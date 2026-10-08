# [Feature]: Add an explicit composable credential chain

### Affected areas

- credentials

### Dependencies

- None

### Problem and scope

- Complete the shared runtime foundation before product generation.
- This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Distinguish an absent credential source from an incomplete or invalid source.
- [ ] Skip only absent sources and stop on configuration, retrieval or cancellation errors.
- [ ] Support environment and explicitly supplied providers without implicit metadata or process execution.
- [ ] Test provider precedence, cancellation and errors without leaking secrets.

### Documentation and verification

- English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.
