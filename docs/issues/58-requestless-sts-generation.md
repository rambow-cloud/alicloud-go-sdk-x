# [Feature]: Generate requestless STS GetCallerIdentity from official Darabonba DSL

### Problem and evidence

- Pinned STS main.tea:500 exposes getCallerIdentityWithOptions(runtime) without a request-model parameter.
- Coverage reports DSL_OPERATION_SIGNATURE.

### Scope and dependencies

- Member of #57; accepted generator baseline #35/#36/#38.
- Support this reviewed signature in semantic projection/normalization, complete IR and Go emission without per-operation field overlays or handwritten generated clients.
- Preserve AWS-style context/input/options API conventions and define/document the empty input representation.
- Keep exact response fields, signed AK protocol behavior, mock interface and Metadata.
- No dependency on the parent issue's closure.

### Acceptance

- [ ] GetCallerIdentity lowers/emits deterministically with complete reachable models.
- [ ] Independent requestless/invalid-signature, wire/signing/empty-query/response/absence/error/context/ownership contracts pass; unsupported signatures fail before writes.
- [ ] Generated external Example, English symbol docs and paired guide/source/coverage report are complete.
- [ ] Existing ECS/VPC/STS outputs and signed-operation credential requirements are preserved.

### Verification

- Node 22 frontend checks/tests; sdkgen check and product-check; doccheck, format, vet, Go tests, Linux race/Windows CI.
- No live cloud calls in unit tests.
