# [Feature]: Provide structured operation errors and response metadata

### Affected areas

- core

### Dependencies

- None

### Problem and scope

- Complete the shared runtime foundation before product generation.
- This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Retain APIError compatibility and add Unwrap-capable operation errors.
- [ ] Expose service, operation, request ID, HTTP status and attempt count without default raw-body logging.
- [ ] Preserve errors.Is for cancellation and errors.As for service and transport errors.
- [ ] Handle non-JSON service failures and reject malformed successful JSON using encoding/json/v2.

### Documentation and verification

- English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.
