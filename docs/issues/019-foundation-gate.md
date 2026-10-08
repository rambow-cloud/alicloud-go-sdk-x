# [Maintenance]: Validate the unified runtime foundation before enabling generator development

### Affected areas

- core, tools

### Dependencies

- #3, #4, #5, #6, #7, #9, #10, #11, #12, #13, #14, #15, #16, #17, #18

### Problem and scope

- Complete the shared runtime foundation before product generation.
- This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Verify all eleven requested foundation capabilities with a handwritten ECS/STS reference.
- [ ] Run formatting, bilingual/API documentation checks, vet, Go tests and Linux race CI.
- [ ] Publish equivalent English/Chinese usage guides and a supported-operation matrix.
- [ ] Keep generator issue #8 blocked until the foundation acceptance criteria pass; do not claim full service coverage or live-cloud validation.

### Documentation and verification

- English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.
