# Implement the shared HTTP runtime

- GitHub issue: #3.

- Depends on standalone signing, middleware, endpoints and structured errors.
- Propagate context; copy requests/config; bound response bodies; decode JSON v2; integrate per-attempt signing and retry without a circular signer dependency.

- Acceptance includes implementation, meaningful offline behavior tests, English Go comments, executable Examples and equivalent bilingual guides.
- GitHub remains the source of execution status.
