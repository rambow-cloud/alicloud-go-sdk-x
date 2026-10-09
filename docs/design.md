# Shared runtime design

[中文](design.zh-CN.md)

- [RPC expansion #83](dsl-rpc-expansion.md) permits native DSL map fields inside concrete operation inputs. Dynamic JSON fields use explicit IR encoding; request/response roots remain typed structs.

- Runtime-first sequence is defined in development-path.md.
- Before v1, APIs may change.
- The module is github.com/rambow-cloud/alicloud-go-sdk-x and requires Go 1.27.

- Public packages provide middleware stages, endpoint.Resolver, retry.Retryer, credentials.Provider/Cache/Chain, pagination.Paginator[T], waiter.Waiter[T], typed ECS/STS reference clients, testing helpers and optional OTel.
- Signing stays internal.
- The root package owns Config, the common HTTP invocation boundary, operation errors and response metadata.
- Product packages own encoding, models, idempotency and waiter policies.

- Operations accept context first and return typed outputs.
- Configuration is private after construction; input query/header/body data is copied before use.
- Custom extension objects must document concurrency.
- Inject net/http clients; never modify the caller's client.
- Default total request timeout is 30s; shorter caller deadlines win.
- Limit response reads to 8 MiB by default and disable automatic authenticated redirects.

- Use encoding/json/v2 directly and retain strict duplicate-name/UTF-8 handling; unknown response fields are tolerated for forward compatibility.
- Optional scalars use pointers only where absence matters.
- Do not expose generic maps as product request/response APIs.

- Middleware separates once-per-operation Initialize/Serialize/Build from per-attempt Finalize/Deserialize.
- Every attempt obtains credentials and signs its actual endpoint, query and payload bytes.
- Default retries are off.
- An opt-in standard policy has at most three attempts, bounded full-jitter backoff and a per-policy retry budget; non-idempotent writes are not retried.

- Preserve APIError and add Unwrap-capable OperationError.
- Default error text excludes raw messages/bodies/query values; callers can inspect fields with errors.As and cancellation with errors.Is.
- Never log credentials or configure global OTel providers/exporters.

- Provider precedence is explicit.
- EnvProvider reports absent vs incomplete keys; a chain skips only absent providers.
- Cache expiry/early refresh and concurrent refresh ownership are documented.
- STS helpers live outside credentials to avoid import cycles.

- Application guidance is STS-first: generated STS -> renewable role provider/cache -> service client.
- The user-corrected [default configuration #68](default-configuration.md) adds an explicit LoadDefaultConfig bootstrap that discovers temporary environment and native CLI Profile/OAuth sources with bounded refresh.
- Config/service Options still accept providers only; long-lived sources require deliberate provider registration.
- Custom providers remain supported.
- Direct constructors reject nil/typed-nil before requests and do not discover sources or perform credential HTTP calls during construction.

- Human-facing Markdown is paired English/Chinese; Go comments and GitHub issues use English.
- Every public package has doc.go and a runnable external Example.
- Internal-only packages do not pretend to have user APIs.
- Metadata/schema licensing and version provenance are required before generation.
- Release and browser indexing steps are in releasing.md.

- The default operation authentication is ACS3.
- Explicit AnonymousRPC mode is limited to the pinned empty-body HTTPS POST RPC contract: native common query framing, no signature/source credentials, provider retrieval skipped.
- Generated OIDC/SAML select this mode from reviewed official handoff evidence; nil providers remain invalid and explicit AnonymousProvider configures anonymous-only use.
- Signing internals stay private.
- See [evidence](sts-anonymous-rpc.md).
