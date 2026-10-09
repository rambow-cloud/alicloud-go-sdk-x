# Source normalization

[中文](source-normalization.zh-CN.md)

- Current client paths follow [service consolidation #81](service-consolidation.md): the old services/ packages and Go emitter are removed; generate/check use complete products. Earlier bridge workflows below retain historical evidence only.

- Stage [#34](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/34) implements the first stage of the [authoritative roadmap](product-generator-roadmap.md), on top of the #31 five-operation compatibility bridge.
- Complete official DSL and its semantic parser remain primary.
- Optional pinned CLI metadata uses a versioned adapter and retains provenance, English/Chinese descriptions and CLI/backend attributes separately from wire properties.
- No new runtime dependency or public API is introduced.

- Canonical ECS DescribeImages/DescribeRegions/DescribeInstances fixtures, version.json and Apache-2.0 LICENSE are pinned in [sources/openapi-meta](../sources/openapi-meta/README.md).
- The source lock verifies revision URLs, SHA-256, paths and inventory before bridge projection writes.
- The Node frontend checks present optional enrichment against the selected DSL inputs/protocol; Go independently checks all selected snapshots against the projection, including indexed leaves and explicit compatibility exceptions.
- Go generation consumes the pinned DSL projection without Node or canonical fixtures.

| Source representation                                                 | Normalized meaning                                                                           |
| --------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| name: region_id; raw_name: RegionId; options: --biz-region-id         | Exact wire RegionId; CLI names/options retained as attributes                                |
| bool/int; location; required; param_style: repeatList; element.fields | Boolean/integer, query binding, source requiredness and repeated structural models           |
| Key/Value and deprecated key/value                                    | Distinct case-sensitive members, never merged or renamed                                     |
| Filter.1.Key/Value through Filter.4.Key/Value                         | Eight literal aliases of the DSL Filter array's Key/Value leaves; no inferred maximum length |
| Images array with itemName: Image                                     | Images object with Image array; nested wrappers restored recursively                         |
| backendName/nullToEmpty/valueMapping                                  | Preserved source annotations, never automatic SDK transforms                                 |
| Top-level GET\|POST versus operation.method POST                      | Allowed CLI methods retained separately; actual recipe POST is cross-checked                 |
| operation_type: read                                                  | Source annotation; does not prove retry safety                                               |

- Index binding accepts positive decimal positions without leading zeros, bounded structural nesting, and exact member case.
- Direct and flattened bindings competing for the same root, missing/forged projection aliases, type/case drift and indexed requiredness disagreement fail before outputs.
- A required indexed root without API requiredness evidence also fails.
- Explicit API-required/DSL-optional approvals remain unchanged; DSL optionality is not proof that the service accepts omission.

- Source review found a real compatibility exception: DescribeInstances.Tag metadata contains optional Tag[].key/value absent from the current DSL.
- The paired decision record and machine policy explicitly list those two optional string metadata-only paths.
- They remain outside the current public Go subset.
- New missing fields, changed case, or newly required members are rejected; no case-folding or inferred wire alias is permitted.
- Filter is removed from the DSL-only approval inventory because its eight bindings now match.

- The legacy bridge still emits five operations with its existing selection policy.
- Full operation/model discovery without per-operation snapshots is stage #35, followed by batch emission #36.
- Normalization does not claim those stages are complete or that every canonical response field has been live verified.
- Selected response/model fields still pass the existing Go cross-checks; real DescribeImages nested wrappers are compared to official parser output in adapter tests.

- From the repository root after installing tools:

```sh
node tools/darabonba/frontend.cjs generate
go run ./internal/cmd/sdkgen generate
node tools/darabonba/frontend.cjs check
cd tools/darabonba
npm test
cd ../..
go run ./internal/cmd/sdkgen check
go run ./internal/cmd/doccheck
go vet ./...
go test ./...
```

- Generation/check/test commands are offline.
- Tests cover real case/style/requiredness/ wrapper data, malformed metadata and source tampering, array positions beyond the declared sample, and write-before-validation prevention.
- The existing cross-backend test checks all generated Go file bytes, so no public Example or package contract changes are expected.
- Keep Go 1.27, direct JSON v2 and English-primary comments.

- Browser/live status: **NOT RUN** for this stage.
- For further evidence, open [DescribeImages](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeImages) and inspect the request Tag/Filter fields and response Images.Image wrapper; open [DescribeInstances](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstances) to inspect Filter and native pagination.
- Record UI hints separately from CLI local validation and actual authorized HTTP responses in the paired decision document.
- Prior #30 reads are historical and do not verify these normalization changes.
