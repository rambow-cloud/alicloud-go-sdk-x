# [Maintenance]: Enforce deterministic regeneration and generator acceptance in CI

- GitHub issue: #23.

# [Maintenance]: Enforce deterministic regeneration and generator acceptance in CI

### Problem and evidence

- Foundation #19 passed.
- The four reference operations remain handwritten.
- Official source: https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/.
- Development path: docs/generator.md; parent #8.

### Scope

- Implement offline write/check commands with an explicit owned file set, full render before mutation, unmarked-file protection, stale-file detection and stable output.
- Add CI regeneration checks, drift/failure tests, bilingual contributor workflow and an acceptance map for parent #8.
- Record Linux race and Windows results before closing completed issues.

### Dependencies

- #22

### Acceptance criteria

- [ ] Check mode detects edited/missing/stale files without mutation; invalid inputs produce no output changes. Run regeneration, documentation/language gates, vet, all tests/Examples and CI once after integration. Keep benchmark #20 separate.
- [ ] Original English-primary code comments and equivalent English/Chinese Markdown updated together.
- [ ] Record exact commit and relevant successful verification before closure.

### Affected areas

- tools
