# One service package path

[中文](service-consolidation.zh-CN.md)

- Decision: use only `service/<product>` for SDK clients. Remove the five-operation `services/` bridge before v0.1.0.
- This user-approved route replaces earlier requirements to preserve the bridge or its generated output. Preserve historical evidence and upstream source notices.
- This is a breaking v0 change. Old imports and selective response models are removed; callers must use native pointer fields and complete response containers.

## Work

- Migrate the root cross-capability contract test to generated ECS and STS clients.
- Use the generated STS adapter for the role provider. Make `NewAssumeRoleProvider` accept `service/sts` types; keep `NewAssumeRoleProviderFromClient` as a forwarding convenience for current consumers.
- Preserve provider/cache rotation, input ownership, cancellation, pagination, waiter, retry and telemetry checks on the full-DSL path.
- Remove legacy client packages, their emitter and generation entry points. Keep source normalization and metadata cross-check evidence where used by parser tests.
- Route `sdkgen generate/check` to complete product generation so existing commands cannot recreate the bridge. Retain `product-generate/product-check` spellings.
- Update paired guides, generated documentation templates, working agreements and CI.
- Re-run affected consumer acceptance and record its actual revision. Historical live evidence remains historical; no cloud calls, tag or publication are part of this work.

## Verification

- Node 22 frontend checks and tests; language and automation checks.
- Full-DSL generation consistency, package docs, formatting, vet and Go tests.
- Isolated STS and ECS/VPC consumer tests and runnable examples; Linux race and Windows checks in CI.
- No production import of `services/`; no generator command recreates it. Product inventory remains STS 4, ECS 283 and VPC 296 supported operations.
- Run `sdkgen product-check` once: `sdkgen check` invokes the same backend. Do not duplicate equivalent regeneration gates.

## Migration

- Replace `github.com/rambow-cloud/alicloud-go-sdk-x/services/<product>` with `github.com/rambow-cloud/alicloud-go-sdk-x/service/<product>`.
- Use `NewFromConfig`; supply pointers for optional fields and read the native response containers.
- Use `stscreds.NewAssumeRoleProvider` with generated `sts.AssumeRoleInput`. The `FromClient` spelling uses the same implementation.
- The shared runtime, credential providers and cache remain reusable.
