# v0.1.0 STS acceptance evidence

[中文](sts-v010-acceptance-report.zh-CN.md)

- Issue #60: technical acceptance and the independent developer gate are separate. #58/PR #63 and #59/PR #64 are merged.
- Implementation version for the live/consumer run is `38cf05ac458d2e6ed3af165350fb77a10e7e8817`, now on main through `f31ad13e4ed3ba9974dfd5d5a9efff8c01aaf3a6`.
- Go 1.27.1, Windows/amd64, Node 22.21.1, parser 2.2.1; Linux race and Windows CI preserve separate results.
- The acceptance branch adds maintenance guards/fixtures and docs without changing accepted STS source/model/policy/generated artifact bytes or signed runtime behavior.

### Scoped coverage

| Native action      | Discovered/lowered/emitted/compiled | Independent offline contracts/Examples                                             | Live                                                                                     |
| ------------------ | ----------------------------------- | ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| AssumeRole         | PASS                                | PASS; provider/cache consumption and narrow mock                                   | PASS #55: three issuances, four ECS reads, real 900-second expiry renewal and cleanup    |
| GetCallerIdentity  | PASS                                | PASS; complete output, wire/presence/errors/cancellation                           | PASS #60: six identity fields/presence match CLI, coherent metadata, single ACS3 attempt |
| AssumeRoleWithOIDC | PASS                                | PASS; unsigned RPC/token encoding/provider isolation/complete output/redaction     | NOT RUN; successful federation outside required v0.1.0 live scope                        |
| AssumeRoleWithSAML | PASS                                | PASS; unsigned RPC/assertion encoding/provider isolation/complete output/redaction | NOT RUN; successful federation outside required v0.1.0 live scope                        |

- All four pinned actions emit with 19 reachable named/inline models; the empty identity Input is not an official DSL model.
- Emission reports retain compilation/live as not-assessed; this report supplies actual acceptance evidence separately.
- STS has no native pagination/waiter workload in this scope.
- Other products retain their own evidence and are not promoted to first-release product acceptance.

### Consumer and maintenance

- `examples/stsacceptance` pins the official [STS v2.1.0](https://github.com/alibabacloud-go/sts-20150401/tree/v2.1.0), darabonba-openapi/v2 v2.1.13 and tea v1.3.13; go.sum locks its isolated dependency graph.
- Our explicit local replace uses the checked-out SDK revision, recorded by the independent developer.
- Both fixtures execute four actions; our workload also performs native AssumeRole -> provider/cache -> two role-signed identity reads.
- Three external tests and runnable consumer PASS.
- No translator or application refresh loop is needed; performance and official credential-library refresh are not measured.
- Fixed versions have concrete context/options/envelope/mock differences, not a source-compatibility or superiority claim.

- Real source rehearsal compares `c321394a58d9b6e513fabb898ee9857a2c6df852` with `d2c0338636a58a6cafc5316d2ed5d158f1f5b162`, retaining original licensed bytes, blob hashes and pinned imports.
- Native fields/models/types/requiredness and semantic prose are unchanged.
- Real initializer signing/endpoint mappings and anonymous handoffs change; prose coordinates move.
- Historical v2 signed initialization and Anonymous/ callApi are unsupported, not accepted as ACS3.
- Candidate four-action discovery, emission and standalone contracts/Examples PASS; stale policy and historical signed selection fail before writes.
- Current candidate STS bytes match production; original source/IR/policies/generated outputs remain unchanged. [Commands/review](sts-source-rehearsal.md) and [machine evidence](acceptance/sts-source-rehearsal.json) distinguish actual historical drift, the no-change production update and synthetic negative cases.

- The authorized read-only live identity run at 2026-10-08 11:33:27 UTC uses `oss-sftp` OAuth/STS only as an explicitly injected in-memory snapshot after CLI authentication. [Sanitized evidence](acceptance/sts-identity-live.json) contains no keys/account IDs/ ARNs/tokens/raw responses; no cloud objects were created.
- This is not native Profile/ OAuth renewal.
- Explorer browser evidence is NOT RUN, distinct from CLI/SDK comparison. #55 natural-renewal evidence is reused: signed credential retrieval/signing/cache and AssumeRole source behavior are preserved; anonymous additions do not alter that path.

### Gate status and handoff

- The later #68 native configuration supplement adds config.LoadDefaultConfig and renewable CLI Profile/OAuth, with nine external consumer tests including the native loader.
- The authorized 2026-10-08 15:35:31 UTC native run at implementation `656ce39dda0b89ae743645ba1328974a937dc780` passed three generated identity reads, one forced OAuth exchange, cache reuse, authentication-only persistence, reconstructed session reuse and lock release. [Sanitized evidence](acceptance/profile-oauth-live.json) contains no identities or credentials.
- No CLI subprocess or new cloud resources were used.
- Actual live refresh-token rotation and natural OAuth expiry waiting remain NOT RUN because the new login access token was valid.
- The paired [configuration contract](default-configuration.md) records the reviewed PascalCase/ camelCase protocol decision.
- Independent human tasks must use this new behavior; their NOT RUN status and the publication/indexing gate remain unchanged.

- AC-01/02: typed context/options, all modeled fields/presence, independent protocol and strict-JSON fixtures.
- AC-05: explicit provider/cache plus scoped #55 live renewal.
- AC-06: reviewed opt-in retry and non-retrying issuance.
- AC-07: endpoint resolution and isolated explicit overrides.
- AC-08: middleware/OTel lifecycle and secret isolation.
- AC-09: structured safe errors, cancellation and metadata.
- AC-10: narrow mocks and offline test helpers.
- AC-11: paired guides/English Go docs/Examples and licensed provenance; actual publication/indexing remains pending.
- AC-12: deterministic generation, reviewed source policy and real-source rehearsal.
- These are scoped STS mappings, not broader Beta.
- The separate [supplemental consumer review](sts-consumer-review.md) records fresh external-package evidence without replacing independent human acceptance.

- Local verification is recorded in the linked PR: frontend/IR checks, 56 Node cases, both generators, doccheck, formatting/bilingual checks, vet, Go tests/Examples, isolated consumer and standalone regenerated STS.
- Exact-head Linux race/Windows CI must pass before merging the technical delivery.

- UX-04/05 independent docs-only tasks: **NOT RUN — user arranging another Go developer**.
- The author's fixture execution does not prove independent task time/success.
- Provide [instructions](sts-consumer-acceptance.md) and the [paired result template](sts-independent-result-template.md) at the exact PR commit.
- Keep #60 open until actual independent results pass. #61 release/indexing is blocked by this required evidence; no tag or pkg.go.dev indexing is claimed.
