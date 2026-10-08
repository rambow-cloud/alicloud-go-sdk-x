# STS source revision rehearsal

[中文](sts-source-rehearsal.zh-CN.md)

- Issue #60 uses real immutable official revisions, original license bytes and blob/ SHA256 provenance under `tools/darabonba/fixtures/sts-revisions`.
- It compares the 2025-06-30 source to repository revision d2c0338636a58a6cafc5316d2ed5d158f1f5b162.
- This is a bounded STS-source-only rehearsal with accepted pinned imports/parser, not a claim that a newer STS API was just released or that all products were upgraded.
- The candidate STS main/Teafile/catalog match production bytes; historical evolution and this no-change production update are reported separately.

- From the root, after Node 22 frontend checks/tests:

```powershell
node tools/darabonba/rehearse-sts.cjs .git/sts-source-rehearsal
go run ./internal/cmd/sdkgen product-generate -root .git/sts-source-rehearsal/candidate
# Expected failure: stale source-bound policy; no service output may exist.
go run ./internal/cmd/sdkgen product-generate -root .git/sts-source-rehearsal/baseline -operations sts/AssumeRole
# Expected failure: historical v2 initializer; no service output.
Copy-Item .git/sts-source-rehearsal/reviewed-sts-policy.json .git/sts-source-rehearsal/candidate/policies/sts.json
go run ./internal/cmd/sdkgen product-generate -root .git/sts-source-rehearsal/candidate
go run ./internal/cmd/sdkgen product-check -root .git/sts-source-rehearsal/candidate
node tools/darabonba/rehearse-sts.cjs --runtime .git/sts-source-rehearsal/candidate
go -C .git/sts-source-rehearsal/candidate test ./service/sts
```

- Choose a new output directory if it already exists.
- The tool verifies fixtures and accepted candidate bytes before writing, lowers all four actions with the official parser, accounts for complete models, compares inventory/bindings/types/auth/prose/ licenses and checks deterministic IR.
- No production file is changed.
- Imported module versions remain pinned; import resolution updates are outside this rehearsal.

- The policy rebind is explicitly reviewed: only the source-manifest hash changes, with byte-identical candidate STS inputs and unchanged reviewed native field paths.
- Do not silently rebind a changed DSL or invalid policy.
- To compile candidate output, use --runtime to copy the current runtime/credentials/endpoint/middleware/retry/internal signing/rpcmodel sources into a fresh standalone canonical module, add generated STS, sdktest and independent contracts, and run its Examples/tests.
- Record exact commands/result in the acceptance report.
- Compilation and live coverage are not inferred from IR.

- Also record expected-negative stale-policy failure before writes, malformed/unknown authentication selection rejection, and rollback by proving the original source/IR/ policy/generated artifacts' hashes unchanged.
- Synthetic incompatible fixtures remain distinct from the real upstream revision comparison.

- The historical initializer explicitly sets v2; signed actions now report DSL_PRODUCT_AUTH_INITIALIZER rather than being emitted with incompatible ACS3.
- Historical OIDC/SAML callApi remains unsupported.
- Current candidate inherits the ACS3 default and uses anonymous doRPCRequest.
- Native models/fields/types/requiredness and semantic prose are unchanged; prose coordinates and initializer endpoint mappings change.
- Endpoint-map parity across all historical regions is not promoted to accepted coverage; required live region remains cn-hangzhou.
- The runtime-fixture command in the shared snippet automates standalone copying/compilation; both old-policy and historical signed selections must fail before service writes.
