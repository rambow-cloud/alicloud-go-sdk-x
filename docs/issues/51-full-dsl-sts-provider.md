# Full-DSL STS credential provider composition

### Affected areas

- credentials, sts, testing

### Dependencies

- #49 defines product AC-05/UX-04.
- Foundation/cache and full-DSL generation are already accepted; documentation PR will be the explicit stack base.

### Problem and evidence

- feature/stscreds.NewAssumeRoleProvider imports services/sts and its typed operation interface/options.
- A service/sts.Client cannot satisfy this older interface, even though it provides the complete generated AssumeRole API.
- Consumers must manually translate the new response into credentials.Provider, defeating AC-05/UX-04.

### Scope

- Add NewAssumeRoleProviderFromClient for service/sts.AssumeRoleAPI, keeping the existing constructor and returned AssumeRoleProvider contract.
- Copy input pointers and option registrations at construction and each retrieval.
- Validate required role/session fields and existing reviewed helper rules without changing generated operation validation.
- Convert the exact full-DSL response to a credential snapshot, with valid nonempty keys/token and future RFC3339 expiration.
- Reuse credentials.Cache; preserve source credentials, structured errors and cancellation; sanitize invalid expiration diagnostics.
- Do not edit generated files, add dependencies or infer retry safety.

### Acceptance criteria

- [ ] New full-DSL constructor, small mock seam, redacted formatting and zero/nil safety; old constructor/examples preserved.
- [ ] Pointer presence/optional fields and duration width retained; constructor/API/callback mutations do not persist; concurrent retrieval supported with safe extensions.
- [ ] Reject nil/typed-nil API, nil options, invalid required/helper fields, missing/blank keys/token and missing/malformed/expired expiration; diagnostics omit raw values.
- [ ] errors.Is/As preserve API failures and pre/post-call cancellation. Canceled callers do not cancel shared cache refresh.
- [ ] Offline generated STS -> provider -> cache -> generated ECS request verifies original source signing, assumed-role signing, native fields and expiry rotation. Default/non-idempotent token issuance remains non-retrying.
- [ ] Runnable external Example, doc.go/Go docs and equivalent English/Chinese provider guide and acceptance evidence.
- [ ] doccheck, formatting, vet, full Go tests and product-check pass; Linux race/Windows CI recorded separately.

### Limits

- This issue establishes offline composition.
- Live role/refresh, native Profile/OAuth loading, independent UX testing, Beta completion and publication remain open acceptance cases.
- No real role or cloud write is needed.
