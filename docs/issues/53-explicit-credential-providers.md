# Explicit credential providers

## Problem and evidence

- The user requires STS-first usage and long-lived access keys to be secondary, available only through explicitly declared providers.
- Config already exposes only CredentialsProvider, and sources are explicit; however the README first-call path uses long-lived keys directly, and root NewClient, NewChain and NewCache accept typed-nil provider pointers.
- A nil ProviderFunc also passes root construction.
- These values can panic during signing/refresh rather than fail before requests.

- AWS Go SDK v2 documents provider injection and an explicit StaticCredentialsProvider for hard-coded keys: https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html .
- AWS permits custom providers and automatic discovery in its default loader; this SDK intentionally has no implicit discovery.
- This scope does not impose an STS-only runtime or reject deliberate custom providers.

## Scope and dependencies

- Depends on merged #49 and #51.
- Follow the updated docs/development-path.md and AC-01/05/11, UX-01/04.
- Make STS provider/cache primary in the README and a complete deterministic external Example.
- Document static/env providers as explicit secondary paths.
- Reject nil/typed-nil sources at runtime, chain and cache construction, including operation configuration replacement; do not retrieve credentials during construction.
- Preserve StaticProvider, ProviderFunc and exact Alibaba wire names.

- No new public credentials wrapper, generator changes, dependencies, implicit Profile/environment discovery, live calls or STS-only enforcement.

## Acceptance criteria

- [ ] Config/service Options contain no bare access-key/secret/token fields; a populated environment never substitutes for an omitted provider.
- [ ] Runtime and generated NewFromConfig reject nil/typed-nil/nil-function sources before HTTP or retrieval; operation replacement preserves output and OperationError wrapping on rejection.
- [ ] Chain/cache reject typed-nil sources; explicit static/env/custom and STS providers remain usable, context/error/redaction contracts remain intact.
- [ ] A runnable external Example and equivalent README snippets compose explicit source -> generated STS -> role provider/cache -> generated ECS and verify role signing with scripted HTTP.
- [ ] Paired docs, AGENTS.md and acceptance scope state STS-first guidance, explicit long-lived opt-in and deliberate differences from AWS, without claiming live/Beta acceptance.
- [ ] Formatting, bilingual/doccheck, vet, full Go tests/Examples and Linux race/Windows CI pass; no generated outputs change.

## Verification plan

- Use synthetic environment keys, rejecting providers that panic if called during construction, scripted transport call counts/signing assertions, reflection checks of credential configuration shape, existing rotation/concurrency contracts and runnable Example output.
- Execute local Go gates once; CI covers Linux race, Windows and unchanged generation/frontend checks.
