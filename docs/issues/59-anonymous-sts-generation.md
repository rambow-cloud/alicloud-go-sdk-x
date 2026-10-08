# [Feature]: Generate anonymous STS OIDC and SAML RPC operations from Darabonba

### Problem and evidence

- Pinned STS main.tea:294/426 use authType=Anonymous and doRPCRequest.
- Coverage reports DSL_PROTOCOL_PROFILE for AssumeRoleWithOIDC/SAML.
- Current signed runtime must not be bypassed by globally permitting missing providers.

### Scope and dependencies

- Member of #57; accepted #35/#36/#37/#38 baseline.
- Independent of requestless operation work.
- Normalize constant doRPCRequest handoff; represent explicit operation authentication in IR and safely execute reviewed anonymous RPC over HTTPS.
- Generate both complete operations/models/interfaces/docs.
- Tokens/assertions are supplied explicitly; no implicit discovery or new federation provider helper.
- Preserve signed operations' missing-provider rejection and conservative no-retry defaults.

### Acceptance

- [ ] Both operations lower/emit/compile with exact native fields, complete responses and optional presence semantics.
- [ ] Anonymous calls omit signature/authorization/source security token and never retrieve a credential provider; signed AssumeRole/GetCallerIdentity still require explicit providers.
- [ ] Independent wire/body/response/error/context/ownership/negative protocol fixtures verify the official helper contract; unknown/dynamic auth/handoff fails before writes.
- [ ] Default formatting/errors/middleware/traces protect OIDC tokens, SAML assertions and returned credentials.
- [ ] English docs, paired guides and deterministic offline Examples disclose that successful live federation is NOT RUN, outside v0.1.0 required live scope.

### Verification

- Node 22 contracts before both generation checks, doccheck, formatting, vet, tests/Examples, Linux race and Windows CI.
- Establish exact protocol evidence from pinned official imports before implementing; no unit-test cloud/IdP access.
