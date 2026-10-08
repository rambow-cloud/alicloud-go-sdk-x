# STS reuse review

[中文](sts-reuse-review.zh-CN.md)

- Tracking: [#72](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/72).
- Review the accepted STS baseline. Preserve the public constructors and pinned source files.
- Follow the [development path](development-path.md): document findings, open the issue, then implement and verify.

## Reusable layers

- Official Darabonba DSL and its semantic parser supply operations, models and protocol behavior.
- Complete IR drives `service/sts` generation. All four actions use the same product backend; no handwritten operation bypasses it.
- Scoped actions: `AssumeRole`, `GetCallerIdentity`, `AssumeRoleWithOIDC` and `AssumeRoleWithSAML`.
- Shared runtime supplies signing, endpoint resolution, middleware, errors and retry policy. Issuance remains non-retrying.
- `feature/stscreds` is a handwritten credential adapter over the generated `AssumeRoleAPI`, not a second STS protocol implementation.
- `credentials.Cache` handles shared refresh and expiration. Native Profile role composition uses the generated client and the same adapter/cache.
- Reviewed role validation is shared internal policy. It must not translate native requests through a compatibility wire model.
- `services/sts` remains a bounded compatibility bridge. Its public constructor remains supported; the primary generated path does not execute through it.

## Findings and changes

- **Option ownership:** the bridge provider passes its stored option slice to custom APIs. Copy registrations for every call; test an API that mutates both input and options.
- **Validation coupling:** the generated provider constructs a bridge request just to validate it. Extract the accepted rules into a shared internal validator without changing optional field presence or integer width.
- **Guide reuse:** the anonymous guide template hardcodes STS action names. Derive names from IR; keep STS federation acceptance scope in STS guides only.
- Centralize cancellation and returned credential checks for both provider constructors. Keep request/response adaptation specific to each API type.

## Verification plan

- Reproduce option mutation before the fix. Check sequential and concurrent calls after it.
- Parse renamed signed, requestless and anonymous actions with the official semantic parser. These are synthetic reuse tests, not new service coverage.
- Emit and compile a differently named synthetic product through the shared backend. Check actual guide names and authentication modes.
- Run Node frontend check/tests before Go gates; then regeneration checks, doccheck, vet, package/consumer tests and formatting.
- Record local results below. Linux race and Windows checks must pass for the final PR head.

## Scope limits

- STS has no paginator or waiter. Do not infer these capabilities from names or fields.
- OIDC/SAML raw operations are generated. Renewable federation providers and live federation acceptance need separate scope.
- Retaining the old constructor means `feature/stscreds` still imports the bridge for that API signature. This review removes its use from primary validation, not the compatibility signature.
- Keep existing [live renewal evidence](live-sts-renewal.md). No new cloud calls are required for this review.
- Independent human acceptance (#60) and publication/indexing (#61) remain separate.

## Results

- Fixed all three findings without changing public signatures, source pins, production IR, policies or generated STS Go code.
- The option-mutation regression failed before the fix; sequential and eight concurrent calls now pass. Existing native provider, cache and Profile tests also pass.
- Node 22 frontend check and all 57 tests pass. The renamed fixture was rerun after correcting its logical source path; production source validation was not relaxed.
- SDK regeneration checks, doccheck, vet, Go package tests/Examples, standalone consumer tests and formatting pass locally.
- The isolated module compiles and runs ECS, STS, VPC and the renamed synthetic product with the shared runtime only. Its four calls retain wire fields and authentication; only the two signed calls retrieve source credentials.
- CI remains a merge gate. Record final Linux race/Windows results on issue #72 and its PR; local results do not replace CI.
