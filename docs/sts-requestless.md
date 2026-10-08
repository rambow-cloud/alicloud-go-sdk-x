# Requestless STS generation

[中文](sts-requestless.zh-CN.md)

- Issue #58 lowers the pinned official GetCallerIdentity function with only a runtime parameter.
- Its IR request root is `{"kind":"empty"}` with no bindings or fabricated DSL model.
- The Go backend emits `GetCallerIdentityInput struct{}` to preserve `client.GetCallerIdentity(ctx, input, optFns...)` and the small GetCallerIdentityAPI.
- Nil and zero-value input both send no query members.
- The complete native response and Metadata remain typed; pointers distinguish absent and explicit empty fields.

- This is a signed RPC action: explicitly register a source provider, including when the account needs no RAM permission for this operation.
- Construction never reads credentials.
- Context cancellation/deadlines and structured service errors survive; default error text omits raw service messages.
- JSON v2 rejects duplicate names and tolerates unknown response members.
- Requestless functions must use the exact reviewed runtime signature, empty OpenApiRequest, constant protocol and callApi handoff; extra behavior/fields remain unsupported before writes.

- Generate through the official frontend/discovery and sdkgen product-generate, then check both generation paths, doccheck, vet, tests/Examples, Linux race and Windows CI.
- The generated product guide/coverage and ExampleClient_GetCallerIdentity document the operation; independent offline wire/output/error/context tests live under service/sts.
- Empty Go inputs do not inflate official reachable model counts.
- Real identity evidence is accepted separately under #60; no live call is made here.

- The same source pattern also automatically admits pinned VPC ListGeographicSubRegions.
- Its generated complete types/API/Example and coverage are additive; existing operations and retry policies are unchanged.
- It stays unreviewed for optional capabilities and does not expand v0.1.0 STS acceptance.
