# [Release]: Deliver v0.1.0 STS through the complete Darabonba pipeline

### Problem and scope

- The user selects STS end-to-end Darabonba delivery as v0.1.0.
- Current pinned STS discovers four actions but emits only AssumeRole.
- Preserve the official DSL/parser -> normalized IR -> our Go backend -> shared runtime route, Go 1.27/JSON v2, AWS conventions and bilingual pkg.go.dev support.
- Authoritative scoped plan: docs/sts-v0.1.0.md.
- This experimental scope supersedes the broader candidate Beta schedule without claiming overall Beta or full-cloud coverage.

### Acceptance and dependencies

- All four pinned actions must be generated, compiled and independently offline-tested/documented.
- Preserve explicit source providers and generated STS provider/cache composition; retain #51/#53/#55 evidence.
- Require scoped live AssumeRole renewal and GetCallerIdentity, external consumer/official-v2 comparison and independent-user task evidence, real source-update rehearsal, immutable release/licenses/migration notes and same-version pkg.go.dev browser evidence.
- Successful OIDC/SAML live federation and federation provider helpers are explicitly future scope; record NOT RUN, never claim live success.
- No STS pagination is invented.

- Child tasks will be linked after creation; parent remains open until every required release case passes.
- Child membership is not a dependency on this parent.
- This planning task creates no tag/cloud resources.

### Verification and documentation

- Node 22 frontend checks/tests before Go generation/product checks, doccheck, vet, tests/Examples, Linux race and Windows CI.
- Record discovered/lowered/emitted/compiled/offline/live/published separately; required SKIP/NOT RUN keeps the gate open.
- Maintain docs/sts-v0.1.0.md, AGENTS.md, development/acceptance/release guides and paired issue specs.
- Maintain one status label and Project Status per issue.
