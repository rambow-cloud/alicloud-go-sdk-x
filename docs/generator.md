# Generator development path

[中文](generator.zh-CN.md)

- Current client paths follow [service consolidation #81](service-consolidation.md): the old services/ packages and Go emitter are removed; generate/check use complete products. Earlier bridge workflows below retain historical evidence only.

- The authoritative new direction is [product-generator-roadmap.md](product-generator-roadmap.md).
- It supersedes conflicting instructions below: complete DSL is primary, metadata is optional, operations/models are discovered automatically and overlays are sparse policy/compatibility exceptions.
- Below describes the five-operation bridge and historical acceptance, not future product-generator prerequisites.

- Foundation #19 passed before this work.
- Generator #8 is delivered in three ordered, independently reviewable issues: metadata/IR, emission/integration, then regeneration and acceptance.
- Benchmarks #20 remain separate.
- Issues are created before code.

- Execution dependencies: #19 -> #21 (metadata/IR) -> #22 (emission/integration) -> #23 (regeneration/CI).
- Parent #8 closes only after all three pass acceptance.

- After #31, production uses official product DSL -> official Darabonba semantic parser -> deterministic protocol/binding/model projection + pinned public metadata + reviewed overlay -> validated intermediate representation (IR) -> formatted Go and paired guides.
- See [migration/tools](darabonba-migration.md) and [source decisions](darabonba-decisions.md).
- The first supported profile is RPC over HTTPS, POST `/`, query parameters and a JSON 200 response.
- Initial operations are ECS DescribeRegions, DescribeInstances, DescribeInstanceStatus and STS AssumeRole.
- Unsupported selected shapes fail generation; the generator does not guess ROA, body serialization or endpoints.
- Expansion #25 adds VPC DescribeVpcs with a page-only paginator.
- Expansion #24 supports bounded offline local schema references; see [the schema matrix](generator-expansion.md).

- The importer is an explicit network command.
- Generation and verification are offline.
- Each manifest records product/version, the official source URL, retrieval time, raw SHA-256 and extracted snapshot SHA-256.
- Snapshots retain protocol facts, not upstream descriptions/examples.
- Alibaba's metadata page states no redistribution license for its prose; do not label that prose MIT.
- Our extraction, overlays, templates and generated comments are original project work under LICENSE.
- Keep source attribution with snapshots.

- Overlays select the public API subset and supply English/Chinese field guidance, idiomatic names, explicit idempotency, JSON-string-array conversions, time conversions, sensitive-model redaction, validation hooks and reviewed paginator/waiter rules.
- DSL supplies operation constants, guarded bindings and model shapes; metadata cross-checks supported methods, locations, API requiredness, types and response paths.
- Explicit approved differences preserve the public contract; new differences fail.
- Overlays cannot silently invent a wire field or change its type.
- Unknown JSON members in an overlay/manifest fail; unselected upstream additions are tolerated.
- Removed selected fields, changed types/styles, new required inputs, unsupported selected reference forms, invalid identifiers and missing policy fields fail before output.

- Generated clients delegate execution, signing, retries, credentials, endpoints, middleware, errors and tracing to the existing runtime.
- Paginator/waiter adapters delegate their engines to the shared packages.
- Validation extensions stay handwritten where constraints are prose-only (STS syntax, ECS mutually exclusive paging modes and VPC tag syntax).
- Existing public names and selected fields are preserved.
- Default endpoint rules stay in the foundation resolver; metadata hosts are not automatically trusted or expanded.

- Generation owns an explicit file set, marked `Code generated`.
- It renders all products before writing and refuses to overwrite unmarked files.
- Check mode never writes and reports missing, changed or stale generated files.
- Deterministic output excludes local timestamps/absolute paths; stable sorting and go/format remove map-order differences.
- CI checks regeneration, existing protocol fixtures, cross-capability integration, Examples, public docs, bilingual guides, race tests and Windows portability.

- Commands from the repository root:

```sh
go run ./internal/cmd/sdkgen generate
go run ./internal/cmd/sdkgen check
```

- For source/projection changes, first run the Node installation and frontend commands in [tools/darabonba](../tools/darabonba/README.md).
- Production Go generation requires an official pinned projection; metadata-only Load/Render remains for synthetic backend tests.
- CI checks full official semantic projection independently of Go regeneration.

- Both are offline; `-root PATH` selects another repository root.
- Check never writes.
- Generate repairs owned drift and removes stale files bearing sdkgen's exact marker; unmarked handwritten files remain protected.
- Inputs and ownership are preflighted before mutation.
- An OS I/O failure can leave a partial multi-file update; rerun generate after resolving it.
- Files are replaced individually using temporary files and rename.

- To add an operation: establish an issue and protocol evidence, explicitly import its metadata ([commands](../metadata/README.md)), review the manifest/overlay, select named fields and exact response paths, mark idempotency, supply bilingual guidance and an offline example, then generate and run the existing gates.
- Validators name local handwritten functions of `func(OperationInput) error`; they are compiled by tests, not executed by the generator.
- Policy fields are checked against selected input/output models.
- Schema version 1 supports reviewed `paginators` and `waiters` collections per product.
- Paginator mode `tokens` selects token-only, `pages` selects page-only, and omitted mode selects dual traversal.
- Legacy singular policies remain readable but cannot be mixed with their collection counterparts.
- Generated names must be unique; native policy fields are required.
- Dedicated paginator options and per-page service options are described in [pagination](pagination.md).
- Optional scalar pointers preserve explicit false/empty/zero; location=input models bind repeatList items and are deeply copied.
- Nested response projections use generated JSON v2 methods.
- Local #/components/schemas chains have a 32-node bound; external/dangling/cyclic references and structural siblings fail.
- Composition/maps and general nested query objects are not supported.
- Endpoints continue to use the shared resolver.

- Acceptance mapping: #21 -> metadata_test.go (including synthetic IR and schema drift); #22 -> emit_test.go (isolated synthetic client compilation with GOPROXY=off), existing ECS/STS protocol and paginator/waiter/helper tests plus foundation_test.go; #23 -> generate_test.go, CLI tests, CI check and docs/language gates.
- Expansion #24 -> expansion_test.go (isolated generated compilation, presence/copying, nested JSON and bounded local references); #25 -> services/vpc tests (signed wire, retry ownership, page boundaries, errors and runnable Examples) and label-classifier tests.

- Next profiles (ROA/body encodings, general nested requests and wider service coverage) require separate issues and protocol evidence; this first working generator is not full Alibaba Cloud schema coverage or an API compatibility guarantee before v1.

- Sources: [official metadata guide](https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/), [ACS3](https://help.aliyun.com/zh/sdk/product-overview/v3-request-structure-and-signature).
