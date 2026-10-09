# Darabonba build tools

[中文](README.zh-CN.md)

- The [product roadmap](../../docs/product-generator-roadmap.md) overrides conflicting older per-operation prerequisites. #31 remains a five-operation bridge; #34 adds [source normalization](../../docs/source-normalization.md), and #35 discovers complete products without per-operation snapshots/overlays.
- Its offline command and coverage/IR contract are in [product discovery](../../docs/product-discovery.md).
- The #36 [batch Go backend](../../docs/batch-go-emission.md) consumes this complete IR without legacy overlays.
- Run `go run ./internal/cmd/sdkgen product-generate` and read-only `go run ./internal/cmd/sdkgen product-check` from the root.
- Outputs use `service/`; the old `services/` bridge is removed. `sdkgen generate/check` are aliases for product generation/check. #37 reads optional source-bound `policies/<product>.json` for [reviewed capabilities](../../docs/capability-policy.md), without requiring per-operation entries for full emission.
- Product reports include policy hashes and individual reviewed/unreviewed status.

- #38 [documentation automation](../../docs/product-documentation.md) consumes official parser descriptions/annotations, generates English Go comments and paired usage/source indexes, and reports missing prose/Chinese translations.
- Original example values never enter Go Examples.
- Full Apache terms and source/transformation notices accompany output.

- Issue [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31) replaces the metadata-only production frontend with real official product DSL.
- Requires Node 22 and Go 1.27.
- The official parser 2.2.1 performs syntax and imported-module semantic analysis.
- Official repo-client 1.0.3 resolves explicit network imports; tar 7.5.22 extracts checked archives.
- Exact packages and transitive integrity hashes are locked in package-lock.json.
- Node/Tea dependencies serve build tooling only, outside go.mod.

- From the repository root:

```sh
cd tools/darabonba
npm ci --ignore-scripts --no-audit --no-fund
cd ../..
node tools/darabonba/frontend.cjs generate
go run ./internal/cmd/sdkgen generate
node tools/darabonba/frontend.cjs check
node tools/darabonba/discovery.cjs generate
node tools/darabonba/discovery.cjs report ecs
node tools/darabonba/discovery.cjs check
npm --prefix tools/darabonba test
go run ./internal/cmd/sdkgen check
```

- Installing tools requires the package registry; after installation, projection, generation, checks and tests use local files.
- Normal Go consumers need no Node.
- Go generation consumes checksum-pinned projections, verifies the source lock and cross-checks metadata/policy; CI independently re-parses the DSL on Linux and Windows.
- Both frontends preflight every product before output writes.
- OS write failures can leave a partial update; resolve the failure and regenerate.

- Only explicit import accesses upstream services:

```sh
node tools/darabonba/import.cjs
node tools/darabonba/import-canonical.cjs
```

- The product commit is a constant in import.cjs.
- Imported Teafile wildcard specs are preserved, while manifest.json and .libraries.json lock actual versions and local paths.
- Import may resolve newer registry modules; review source/module/license changes and update metadata/darabonba-decisions.json under an issue before projection.
- The importer writes source artifacts as it fetches them; network failure may require repairing the source lock.
- Generation never fetches or refreshes dependencies.

- Recognized functions are async operationWithOptions(request, runtime), guarded direct query bindings, OpenApiRequest query encoding, constant OpenApi.Params and callApi handoff.
- They lower to our runtime instead of emitting the imported Tea program.
- Selected models support bounded named/anonymous objects, arrays and scalars.
- Unknown behavior, language overrides, new model attributes, incomplete bindings, dynamic parameters and unsupported selected shapes fail.
- Native pagination, waiters, Go names, presence, redaction and idempotency remain reviewed overlay/runtime policy.
- This is a bounded frontend for five operations, not a general Darabonba compiler.

- Read [migration](../../docs/darabonba-migration.md), [decisions](../../docs/darabonba-decisions.md) and [sources/licenses](../../sources/darabonba/README.md) before adding operations.
