# [Enhancement]: Add default configuration and native CLI Profile/OAuth credentials

### Problem and evidence

- The user rejects the existing no-default-discovery/no-native-Profile/OAuth limitation.
- Manual CLI snapshots are insufficient for a modern local developer experience.
- The earlier #53 deliberate AK/SK opt-in constraint was overextended into a blanket discovery exclusion.
- Authoritative revised route: docs/default-configuration.md.
- Pinned official CLI fa14dd7b0359b5be229f9d770a662a86e69c13c1 documents the native profile fields, CN/INTL client IDs, OAuth refresh and STS exchange protocols.

### Scope and dependencies

- Member of #57 and milestone v0.1.0; prerequisite for updated #60 consumer acceptance and #61 publication.
- Preserve accepted #51/#53/#55/#58/#59 behavior and historical evidence.
- Introduce stdlib-only config.LoadDefaultConfig and feature/profilecreds, native CLI JSON profiles, OAuth token refresh/STS exchange, StsToken, explicit-opt-in AK/RamRoleArn, ChainableRamRoleArn, copied configuration and shared bounded caches.
- Explicit provider -> explicit profile -> complete temporary environment -> selected default profile.
- Region precedence and sanitized errors are documented before code.
- No generator/source edits, new dependency, implicit long-lived keys, subprocesses, browser launch, unrelated CLI configuration writes, arbitrary credential URIs or IAM/IdP creation.
- Other CLI modes return explicit unsupported-mode errors and are not claimed.

### Acceptance

- [ ] LoadDefaultConfig(ctx, options...) returns existing Config consumable by generated STS/ECS; service constructors retain strict provider-only configuration.
- [ ] Provider/profile/environment/region precedence, copied state, missing vs invalid sources, typed-nil, strict bounded JSON and duplicate-profile/cycle failures are deterministic and tested.
- [ ] Default loading supports OAuth/StsToken and rejects implicit long-lived keys; explicitly constructed profile/EnvProvider/StaticProvider overrides remain available. Role profiles use generated AssumeRole/helper/cache without translation or refresh loops.
- [ ] OAuth reuses valid STS snapshots, refreshes expired access tokens and exchanges native bearer credentials; rotated sessions persist atomically before exchange and survive reconstruction, unknown/sibling configuration is retained, bounded file locks serialize SDK processes, detected CLI edits fail safely; concurrent callers share bounded renewal, cancellation remains recognizable and default errors/formatting reveal no secrets. Initial login/revocation and shared-file ownership are documented.
- [ ] New public packages have complete English Go comments/doc.go and deterministic external Examples; paired guides and AGENTS/development/acceptance/release routes supersede stale exclusions together.
- [ ] doccheck/vet/tests/format and exact-head Linux race/Windows CI pass; an authorized read-only native local-profile identity check records sanitized results. Record actual live refresh separately from offline tests and prior role-renewal evidence; missing live credentials remain an explicit gate rather than invented PASS.

### Verification

- Offline temporary profile files and injected HTTP, no accounts/network in tests.
- Source-bound independent token/exchange/role fixtures, concurrency/cancellation, negative status/body/expiry/security tests, SDK documentation gates and existing CI.
- Read-only local identity and OAuth exchange are within prior local-profile authorization; never print/cache real credentials in committed artifacts.
- Keep #60/#61 open until new behavior and independent human acceptance/publication/indexing gates are met.
