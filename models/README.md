# Product IR artifacts

[中文](README.zh-CN.md)

- Current lowering/emission: ECS 380/380, VPC 396/403, STS 4/4. See [RPC expansion #83](../docs/dsl-rpc-expansion.md). The older table below is historical.
- Schema v1 adds optional binding `encoding: "json"`, boolean field `attributes.deprecated`, and `kind: "json", dslType: "any"` for dynamic values inside reviewed JSON transforms. No unknown transform is accepted.

- Generated build-time artifacts for [#35](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/35).
- Read [product-discovery.md](../docs/product-discovery.md) before consuming them.
- Schema version 1 and profile rpc-query-json-v1 record the pinned official product DSL, parser version, complete source-lock hash, source/API catalog hashes and Apache-2.0 provenance.
- The original license is [LICENSE.upstream](../sources/darabonba/LICENSE.upstream).
- Do not relabel source-derived definitions under the project MIT license.
- Descriptions/ examples reference coordinates in the licensed source; #38 includes parser description/ summary text for English Go comments, with paired guides and documentation coverage.
- Source example values stay absent and prose is never converted to validators.
- Imported module license evidence remains separately recorded in the source lock, including its documented NOASSERTION limitation.

- manifest.json hashes all product ir.json and coverage.json files.
- Regenerate using `node tools/darabonba/discovery.cjs generate` and check using `node tools/darabonba/discovery.cjs check`.
- Do not edit generated records directly.
- Model IDs retain named/anonymous identity; operation reachableModels lists local model references.
- Field required records DSL optionality, not service-side API requiredness; API constraints/policies remain unassessed.
- Imported RuntimeOptions and other module types are explicit external references.
- Protocol/binding evidence includes unsupported constructs, while executable bindings/protocol exist only for lowered operations.
- Numeric DSL type names remain alongside normalized wire types.

| Product        | Discovered | Lowered | Unsupported | Reachable named / declared models | Reachable anonymous models |
| -------------- | ---------: | ------: | ----------: | --------------------------------: | -------------------------: |
| ECS 2014-05-26 |        380 |     283 |          97 |                       1140 / 1150 |                        913 |
| STS 2015-04-01 |          4 |       1 |           3 |                           11 / 11 |                          8 |
| VPC 2016-04-28 |        403 |     295 |         108 |                       1208 / 1212 |                        520 |

- Counts describe revision ec489e5c3deae95496daae2b41503ac58b221adb under the current strict lowering profile.
- IR generation itself emits no Go clients; the separate #36 [batch backend](../docs/batch-go-emission.md) emits 283 ECS, 295 VPC and one STS operation into `service/`, with separate reports in docs/products.
- Discovery coverage intentionally leaves downstream acceptance **not assessed**; compilation evidence belongs to the emission PR, and product capability policy/live acceptance remain separate.
- Existing five-operation Go bridge acceptance is separate.
- Unsupported operations retain stable reasons and source locations.
- Source coordinates use parser line/column conventions; file paths are relative to sources/darabonba.
- Report ECS reasons with `node tools/darabonba/discovery.cjs report ecs`.
