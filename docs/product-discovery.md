# Product discovery and IR

[中文](product-discovery.zh-CN.md)

- Current generation follows [RPC expansion #83](dsl-rpc-expansion.md): ECS 380/380; VPC 396/403. Earlier counts and consumer records below describe their accepted revisions.

- Stage [#35](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/35) follows the [authoritative roadmap](product-generator-roadmap.md) and #34 normalization.
- Branch issue/35-product-discovery is stacked on issue/34-source-normalization / PR #39, which originally depended on #32.
- This specification was committed before implementation; the complete dependency stack is now [integrated into main](generator-integration.md).

- The build-time command uses Node 22, official parser 2.2.1 and the complete pinned product/import corpus.
- It verifies source hashes and import resolution, then performs semantic parsing offline.
- It never reads legacy metadata manifests, decisions, snapshots or overlays; canonical enrichment is optional and not needed for discovery.
- No Go runtime dependency, generated service API or automatic retry policy is added.

- Operation candidates come from DSL WithOptions functions, API declarations and functions constructing OpenApi.Params.
- Constant action names retain their exact case; ambiguous/dynamic actions remain unsupported with source evidence.
- Optional pinned api-info.json is a coverage cross-check: catalog-only and DSL-only entries remain visible rather than limiting discovery to a catalog whitelist.
- Ordinary forwarding functions are not counted as separate APIs.
- Unexpected candidate behavior is recorded; source/parser failures and malformed inventory are fatal before output writes.

- The versioned product IR records provenance, operation declaration locations, protocol constants, request/response roots, reachable named/anonymous models and exact wire members.
- Model references retain identity instead of recursively duplicating models.
- Original numeric type names, optionality, attributes and source coordinates are preserved; descriptions/examples point to the licensed original source rather than being reinterpreted as validators.
- RuntimeOptions/imported transport models remain explicit external references.
- Unreachable product declarations are accounted for.

- Coverage distinguishes discovered, lowered and unsupported operations.
- Lowering is limited to the existing reviewed HTTPS POST RPC direct-query/json-body profile, with every wire input/response shape checked for unsupported constructs.
- ROA, body/stream, helper transforms, unresolved/recursive/inherited wire models or other unsupported patterns have stable reason codes, messages and source locations.
- A lowered record is not a claim of Go emission, compilation, runtime policy or live-cloud acceptance.
- Those later acceptance stages remain unassessed here.

- Repository-root commands:

```sh
node tools/darabonba/discovery.cjs generate
node tools/darabonba/discovery.cjs report ecs
node tools/darabonba/discovery.cjs check
node tools/darabonba/discovery.cjs check --operations ecs/DescribeImages,sts/AssumeRole
```

- Artifacts live under models/{ecs,sts,vpc}/ir.json and coverage.json with a hash-pinned models/manifest.json.
- All products preflight before writes; check never writes or fetches.
- System write failures can leave partial updates.
- Optional --operations selection is a strict acceptance gate: selected unknown/unsupported operations fail before writes, while unselected unsupported entries remain in the full inventory.
- The report prints actual counts/reasons and locations; no invented API coverage.
- The current pinned corpus accounts for ECS 380/283/97, STS 4/1/3 and VPC 403/295/108 discovered/lowered/unsupported operations.
- See [artifact guide](../models/README.md).

- Acceptance: real full-corpus deterministic regeneration and complete accounting; discovery with no legacy metadata/overlays/canonical files; malformed/unsupported protocol, binding, model and source negative tests; strict selection before writes; paired docs and CLI usage examples.
- Run frontend/discovery checks and tests, sdkgen check, doccheck, vet, Go tests, formatting and language checks.
- CI runs Linux race and Windows Go 1.27.
- Public Go packages continue direct JSON v2 and existing offline Examples.
- Browser/live calls are not part of this stage; full product Go emission is #36, capability policies #37 and documentation automation #38.

- Current scope after #85: product IR/lock schema v3, ECS 380/380, VPC 403/403 and STS 4/4. See [VPC RPC completion](vpc-rpc-completion.md); earlier counts and reasons below are historical.
