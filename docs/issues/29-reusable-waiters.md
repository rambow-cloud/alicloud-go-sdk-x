# [Feature]: Generate reusable waiters with Wait and WaitForOutput options

- GitHub issue: #29.

### Problem and evidence

- Review of abcf961 found behavior defects and incomplete AWS Go SDK v2 design alignment.
- Design and acceptance were written before code in docs/aws-style-remediation.md.

### Scope and dependencies

- Bind only API/options at waiter construction; accept owned input on each Wait/WaitForOutput.
- Generate dedicated options, ClientOptions and an overridable reviewed acceptor.
- Wait returns error, WaitForOutput returns success output.
- Preserve bounded all-ID/page checks, context errors and immutable concurrent waiter use.

- Dependency: #28.

### Acceptance criteria

- [ ] Wait and WaitForOutput examples, two independent input sets and concurrent waits, option overrides/defaults, poll ownership, acceptor customization, missing/unknown/duplicate states, cancellation, timeout and error wrapping. Update every caller, public docs and equivalent bilingual migration guidance.
- [ ] English-primary Go docs, deterministic offline Examples, equivalent bilingual docs and migration notes; offline regeneration, doccheck, language, vet, tests, Linux race and Windows CI.
- [ ] Record exact commit and CI evidence before closing.

### Affected areas

- waiter, tools, ecs
