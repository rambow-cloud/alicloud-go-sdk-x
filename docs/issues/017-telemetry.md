# [Feature]: Add optional OpenTelemetry operation and attempt instrumentation

### Affected areas

- telemetry, middleware

### Dependencies

- #3, #11, #13

### Problem and scope

- Complete the shared runtime foundation before product generation.
- This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Expose an opt-in middleware adapter with injected TracerProvider.
- [ ] Create one logical-operation span and child attempt spans with request metadata.
- [ ] Never set global providers/exporters or record credentials, raw bodies or query values by default.
- [ ] Keep the core import graph standard-library-only and test spans with an in-memory exporter.

### Documentation and verification

- English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.
