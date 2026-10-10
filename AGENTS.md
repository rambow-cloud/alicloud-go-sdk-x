# Project working agreements

- The current #92 stage follows [OSS semantic IR](docs/oss-semantic-ir.md): complete source pinning, Gateway lowering and nested XML checks precede Go emission; frontend acceptance is not public OSS coverage.

- #92 now follows [OSS shared runtime](docs/oss-runtime.md) before product emission: explicit OSS4/body modes, bucket identity, per-attempt XML MD5 and bounded XML errors. No inferred roots or generated/live OSS coverage claim.

- #92 XML root discovery follows [native XML traits](docs/native-xml-traits.md) after the [shared codec](docs/xml-model-codec.md). Official DSL remains authoritative; hash-bound native declarations only supplement explicit serialization facts. Preserve the unresolved ListBuckets case/wrapper conflict and namespace policy; discovery counts are not generated OSS coverage.

[中文](AGENTS.zh-CN.md)

## SDK gap completion (2026-10-09)

- The next #92 scope follows [OSS XML/streaming](docs/oss-xml-protocol.md). Pin complete official product and gateway/helper source first; project exact host, XML root and signing behavior into IR. No ACS3 fallback for OSS, inferred XML roots or claimed streaming without ownership/replay/checksum contracts.

- #92 follows [FC binary generation](docs/fc-binary-generation.md). Current complete IR/lock uses schema v6 (`openapi-http-v1`) with exact path/query/header/body bindings and operation facades. Earlier schemas and the 72-action JSON/none baseline are historical. All 73 pinned FC actions are emitted and verified offline, including binary InvokeFunction. No inferred policy or live acceptance.
- Add pinned products with tools/darabonba/import-product.cjs, reusing all locked transitive imports. Source-bound endpoint exceptions live in metadata/endpoint-source-decisions.json; preserve official bytes and reject unapproved source/coordinate/value drift before writes.

- #91 projects official endpoint initialization into IR schema v4 and generates the shared catalog. Follow [endpoint rules](docs/endpoint-rules.md); private rules need exact reviewed evidence and never fall back to public origins.

- The user requested completion of remaining gaps. Follow [gap completion](docs/gap-completion.md) and parent #87.
- This route replaces stale next-step scheduling. Preserve accepted evidence; track implementation, live behavior, human usability and publication separately.

## One service path

- The user-approved #81 route in [service consolidation](docs/service-consolidation.md) overrides requirements to preserve the old bridge. Only `service/<product>` is supported; no `services/` package or emitter. Keep historical evidence and official source notices.

## Scoped ECS and VPC acceptance

- #85 extends shared RPC generation with explicit query/form locations, native GET and simple string arrays. Product IR and lock use schema v3. Follow [VPC RPC completion](docs/vpc-rpc-completion.md); never infer location, encoding or method from action names.
- #74/#75 use the pre-execution matrices in [ECS acceptance](docs/ecs-product-acceptance.md) and [VPC acceptance](docs/vpc-product-acceptance.md).
- Generated RPC inventory after #85 is 380 ECS and 403 VPC actions. #74/#75 consumer acceptance remains pinned historical evidence; expanded generation does not imply all-action live acceptance. Offline consumer contracts and selected-field live reads are separate; unsupported DSL actions retain reasons.
- Instance token/waiter live transitions and nonempty VPC live continuation remain excluded/NOT RUN or SKIP under follow-up #79. No broader Beta or full-cloud acceptance is claimed.
- Consumer acceptance is implementation-agent execution. Independent human UX and automated test timings remain distinct.
- Preserve STS/ECS/VPC evidence pins against the shared consumer/CI revision. Publication and same-version pkg.go.dev indexing remain #61. No tag is created during product closeout.

## Current delivery and acceptance authority

- The user-approved 2026-10-09 route in [docs/sts-ecs-vpc-path.md](docs/sts-ecs-vpc-path.md) overrides older STS-only release and required-human #60 rules.
- Complete #60 through recorded implementation-agent consumer acceptance. Keep independent human UX in optional #76 and leave its evidence NOT RUN until actual results exist.
- Then complete ECS #74 and VPC #75 before #61. Maintain native dependencies, milestone and Project state; no release during STS closeout.
- Separate agent tests from human task success/timing. Preserve broader Beta UX criteria and accepted live/source evidence.
- Release guard requires current STS agent and ECS/VPC product records; no single-product acceptance can bypass the other gates.

- The 2026-10-09 writing policy replaces all older mixed-language Markdown rules.
- See [documentation style](docs/documentation-style.md).

## Product goal

- Build an independent Alibaba Cloud SDK for Go with idiomatic APIs, small dependencies, predictable request behavior, and documentation available on pkg.go.dev.
- This project is not an official Alibaba Cloud SDK.
- Read `docs/research.md` and `docs/design.md` before changing the public API.

- Follow docs/product-acceptance.md for candidate Beta/release scope, AC/UX criterion IDs and required evidence.
- Record PASS/FAIL/SKIP/NOT RUN per case; skipped required cases keep the gate open.
- Generated/compiled counts, local PoCs and closed implementation issues do not establish Beta/release acceptance.
- Keep Smithy isolated until a separately approved route decision; follow the full-DSL STS composition -> consumer/official-v2 comparison -> authorized live gaps -> source-update rehearsal -> release sequence.

- The user-approved 2026-10-08 first release is v0.1.0 STS, governed by docs/sts-v0.1.0.md, milestone v0.1.0, Project 3 and parent #57.
- This scoped experimental release supersedes conflicting broader Beta scheduling; it does not claim ECS/VPC/STS Beta.
- Execute #58 requestless GetCallerIdentity -> #59 anonymous OIDC/SAML RPC -> #60 four-action acceptance/consumer/source-update rehearsal -> #61 release/indexing. #58/#59 are independent prerequisites for #60.
- Keep full official Darabonba/parser -> IR -> our Go backend/runtime; signed credential rules remain strict, anonymous auth is explicit per operation.
- Preserve #51/#53/#55 evidence and disclose OIDC/SAML live federation as NOT RUN/outside this release's required live scope.
- Maintain milestone membership, native issue hierarchy/dependencies, one status label and matching Project Status; do not claim automatic synchronization.
- Keep parent/milestone open until required release/indexing evidence is delivered.

## Issue-driven development

- Every code or behavior change starts with an issue containing the problem, evidence,
  scope, acceptance criteria, and documentation requirements. Use GitHub issues in
  `rambow-cloud/alicloud-go-sdk-x`; when offline, create a draft under `docs/issues/`
  and synchronize it before opening a PR.
- Keep one independently reviewable issue per branch: `issue/<number>-<slug>`.
- Before implementing, establish how the acceptance criteria will be verified.
- PRs link the issue using `Closes #<number>` for complete work, or `Refs #<number>`
  for partial work, and include behavior, validation, and docs.
- A change is complete only when code, relevant checks, examples, and package docs
  satisfy the issue. Do not close roadmap issues just because an interface exists.
- Never invent issue numbers, completed checks, compatibility, or service coverage.

## Go API rules

- Follow AWS Go SDK v2 calling conventions for service Options, NewFromConfig,
  operation functional options, paginators and reusable Wait/WaitForOutput. Preserve
  Alibaba OpenAPI operation names, exact wire fields and native pagination semantics;
  never invent NextToken for a page-only API. Document intentional defaults and v0
  migration differences rather than claiming upstream source compatibility.

- Reviewed anonymous STS RPC uses explicit AnonymousProvider; it never retrieves source
  credentials. Signed operations still require credentials and nil/typed-nil remain invalid.
- Every blocking operation accepts `context.Context` as its first argument and
  preserves cancellation/deadline errors for `errors.Is`.
- Use standard `net/http`, injectable HTTP clients, and `time.Duration` for timeouts.
- Require Go 1.27 or newer.
- Import `encoding/json/v2` directly for JSON; do not add
  legacy `encoding/json`, third-party JSON libraries, or an experiment flag.
- Keep client configuration private after construction; do not mutate shared state
  or caller requests. Document concurrency contracts of extension interfaces.
- Use concrete request/response types and `errors.As` for service errors.
- Pointers
  represent meaningful absence; do not require helpers for every ordinary value.
- Retry only when the operation's idempotency and error policy allow it.
- Bound total
  attempts and elapsed time, respect context during backoff, and never retry arbitrary
  writes by default.
- Add dependencies only with issue justification.
- Keep the runtime core standard
  library only unless a concrete requirement prevents it.
- Keep signing, encoding, and retry machinery internal until real use requires an
  extension contract. Do not implement cloud operations without protocol evidence.
- Never log credentials, authorization headers, or raw request/response bodies by
  default. Do not commit real credentials or make live cloud calls in unit tests.
- Make renewable STS role providers with credentials.Cache the primary application
  guidance. Config/service Options accept providers only, never bare AK/SK/token
  fields. The user-approved #68 route adds config.LoadDefaultConfig and native CLI
  Profile/OAuth discovery and bounded renewal; follow docs/default-configuration.md.
  Long-lived keys require explicit StaticProvider/EnvProvider or an explicitly
  enabled profile provider; default discovery accepts temporary/OAuth sources.
  Preserve custom providers, reject nil/typed-nil direct service configurations and
  perform no credential HTTP retrieval during construction. This supersedes older
  blanket no-discovery exclusions, without enforcing an STS-only runtime.

## pkg.go.dev definition of done

- Every public package has a `doc.go` overview explaining usage and limitations.
- Every exported declaration, field, and interface method has a Go doc comment
  beginning with its name. Explain zero values, defaults, errors, cancellation,
  ownership, and concurrency when relevant. Use native Go comments, not `@param`.
- Every public package has at least one runnable external `Example` with deterministic
  output and no account/network requirement. Add operation examples as operations land.
- Keep `LICENSE`, canonical module imports, and runnable README snippets current.
- Run `go run ./internal/cmd/doccheck`, `go vet ./...`, and `go test ./...` once for
  a meaningful change. CI runs the race detector on Linux. Check formatting once.
- `docs/releasing.md` describes indexing; local docs passing does not mean the package
  has already been published or indexed.

## Validation and communication

- Avoid redundant validation and repeated retries.
- Repeat only after a meaningful
  state change or a failure whose reason is newly understood.
- If a result can be verified by opening it in a browser, tell the user the exact URL
  and what to inspect instead of checking it using curl or similar local requests.
- For DNS, server-side CloudFormation status and outputs are the source of truth.
  Do not validate DNS locally; provide user verification steps only if needed.
- Do not spawn sub-agents unless the user explicitly requests delegation.

## Language and development sequence

- The user-approved 2026-10-07 route in docs/development-path.md and
  docs/product-generator-roadmap.md is authoritative and overrides conflicting older
  metadata-first, per-operation snapshot/overlay and handwritten-doc prerequisites.
  Historical acceptance remains recorded; superseded plans must not constrain new work.
- The initial five-stage route and #44 review fix are integrated into main as recorded
  in docs/generator-integration.md (2026-10-08). New product/protocol/capability expansion
  starts with a separate issue and acceptance scope; preserve this accepted baseline.
- Execute source normalization -> complete DSL operation/model discovery and IR -> batch
  Go emission -> sparse capability policy -> documentation automation/profile expansion.
  Complete official DSL and the official semantic parser are primary; canonical metadata
  is optional enrichment/cross-check input after representation normalization.
- Discover operations/reachable models automatically without per-API metadata, field/model
  selection or handwritten-doc overlays. Overlays supply reviewed policy and compatibility
  exceptions. #31 is a five-operation bridge, not product-wide acceptance or a permanent
  per-API authoring workflow. Preserve runtime/AWS conventions, Go 1.27 and JSON v2.
- Publish deterministic coverage/reasons; discovered, lowered, emitted, compiled and live
  coverage are distinct. Selected unsupported behavior fails before writes. Normalize
  indexed inputs/itemName wrappers before source conflict decisions.
- Products use only `service/<product>`; #81 removes the legacy bridge and emitter.
  Follow docs/service-consolidation.md for explicit migration.
  Product changes run `sdkgen product-check` (also available as `sdkgen check`).
  Product capabilities use optional source-bound policies under
  policies/ (#37); resolve exact wire paths against complete IR, fail invalid policy
  before writes, and report unlisted actions as unreviewed. Never infer retry safety or
  paginator/waiter behavior from operation names or token-shaped fields. Follow
  docs/capability-policy.md; caller inputs stay owned copies and retry remains opt-in.
- Establish bilingual route docs, then actual issue dependencies, then stage implementation
  on separate issue branches. Each stage includes docs/tests/Examples/pkg.go.dev acceptance;
  the documentation automation stage does not postpone earlier documentation.
- Follow docs/product-documentation.md for #38 documentation: reuse official parser
  prose with source coordinates and Apache notices; generate English Go comments and
  paired usage/contracts/source indexes. Report missing language/prose coverage, never
  infer runtime policy or promote upstream account/resource examples into Go tests.

- Follow [documentation-style.md](docs/documentation-style.md).
- #93 optional canonical field prose follows docs/canonical-prose-enrichment.md. Normalize indexed/itemName representations before exact DSL field/type matching; preserve source bytes, hashes, JSON pointers and exclusion reasons. It must not change runtime models/policy or become a per-operation prerequisite. Regenerate prose after IR changes and run its offline check/tests before Go gates.
- Keep English in name.md and Chinese in name.zh-CN.md, with reciprocal links.
- Update both together.
- Write natural Chinese.
- Use short English sentences and bullet points.
- Keep commands, API names, source references and acceptance evidence accurate.
- GitHub issues, issue drafts/forms and PR templates use English only.
- PR text, diagnostics, identifiers and code comments use English.
- Source literals and original upstream notices stay unchanged.
- Go package and symbol documentation remains English-primary for pkg.go.dev.
- Provide
  equivalent Chinese usage/behavior guidance in the paired package guides.
- Complete the eleven shared foundation capabilities and their handwritten ECS/STS
  contract tests before starting generator implementation. Metadata research is allowed
  earlier; interfaces alone do not satisfy the foundation acceptance gate.
- Read docs/development-path.md before selecting an issue.
- #92 binary responses follow [response-streaming.md](docs/response-streaming.md): implement bounded ownership, cancellation and codec publication in the shared runtime before DSL/Go emission. No late-read retry or unbounded request stream.
- Binary ROA generation follows [fc-binary-generation.md](docs/fc-binary-generation.md): schema v6/openapi-http-v1, complete semantic program review and renamed field/model reuse fixtures. Preserve request bytes and stream ownership; no action-name special cases or inferred retry.
- Document the path, create or
  update the issue, then implement. Keep the issue dependency graph acyclic.
- Every capability includes implementation, behavior tests, executable examples, paired
  documentation, and recorded verification. Keep generator and benchmark issues separate.
- Generator changes also run `go run ./internal/cmd/sdkgen check`.
- Edit pinned metadata,
  overlays, templates or handwritten validators under an issue; do not edit generated
  files directly. Review schema drift and rerun generation before public API checks.
- Signed protocol lowering also reviews explicit product signature initialization;
  unsupported/dynamic algorithms fail selected generation before writes, never silently become ACS3.
- Production generation uses pinned official Darabonba sources/imports and the official
  semantic parser projection. Run the frontend check/tests with Node 22 before Go gates.
  Never silently resolve DSL/metadata conflicts; maintain the paired decision document
  and machine-readable approvals. Explorer browser evidence and CLI evidence are distinct.
- Preserve third-party source, README and license notice bytes under sources/darabonba
  and sources/openapi-meta;
  their upstream documents are source artifacts. All project-authored source/tool guides
  and decisions remain bilingual; upstream notices are not relabeled under our MIT license.
