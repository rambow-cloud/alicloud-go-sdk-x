# [Feature]: Add typed serialization middleware and concrete service Options

- GitHub issue: #27.

### Problem and evidence

- Review of abcf961 found behavior defects and incomplete AWS Go SDK v2 design alignment.
- Design and acceptance were written before code in docs/aws-style-remediation.md.

### Scope and dependencies

- Introduce owned model input/output and a Serialize stage; execute Initialize/Serialize/Build once and Finalize/Deserialize per attempt.
- Generate concrete service Options, NewFromConfig, configuration snapshots and isolated per-call configuration including retry overrides.
- Retain New(Config) and wire Invoke; document migration.

- Dependency: #26.

### Acceptance criteria

- [ ] Typed hooks alter copied inputs before signed encoding; output hooks publish atomically. Verify stage order, retry counts, invalid replacement/short circuits, input ownership, configuration snapshot copying, per-call retries/transports/endpoints and existing protocol/STS/OTel fixtures.
- [ ] English-primary Go docs, deterministic offline Examples, equivalent bilingual docs and migration notes; offline regeneration, doccheck, language, vet, tests, Linux race and Windows CI.
- [ ] Record exact commit and CI evidence before closing.

### Affected areas

- core, middleware, tools
