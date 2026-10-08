# [Feature]: Implement replaceable endpoint resolution with explicit rules

### Affected areas

- endpoints

### Dependencies

- None

### Problem and scope

- Complete the shared runtime foundation before product generation.
- This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Expose a context-aware resolver and function adapter; explicit endpoint overrides take precedence.
- [ ] Validate HTTPS and reject userinfo, fragments and unexpected endpoint query strings.
- [ ] Use documented ECS regional and STS endpoints; unsupported service/region combinations fail explicitly.
- [ ] Test resolver cancellation, overrides, rule ownership and request signing against the resolved host.

### Documentation and verification

- English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.
