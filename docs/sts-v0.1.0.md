# v0.1.0 STS delivery

[中文](sts-v0.1.0.zh-CN.md)

## Current delivery route (2026-10-09)

- Follow [STS/ECS/VPC delivery](sts-ecs-vpc-path.md): finish #60, then ECS #74 and VPC #75, then #61 release/indexing.
- #60 now requires truthful implementation-agent consumer acceptance. Independent human usability moves to optional follow-up #76; its record remains NOT RUN and does not block this release.
- Existing STS-only first-release scheduling and required-human #60 gates below are historical and superseded by this decision.
- Preserve actual technical/live/source evidence. Agent test duration is not human task time.
- Publication remains pending until both product acceptance issues pass. No tag is created by this change.

### Authority and scope

- The user-approved 2026-10-08 first release target is **v0.1.0: STS through the complete Darabonba pipeline**.
- This scoped experimental release takes priority over the earlier ECS/VPC/STS candidate Beta schedule, while preserving its broader [acceptance criteria](product-acceptance.md) and historical evidence.
- Completing this milestone does not prove that broader Beta.
- No release tag or new cloud resources are created by this planning change.

- Use pinned official STS 2015-04-01 DSL/imports, the official semantic parser, normalized complete IR, our Go backend and the shared runtime.
- Ship all four actions in that pinned source under `service/sts`: AssumeRole, GetCallerIdentity, AssumeRoleWithOIDC and AssumeRoleWithSAML.
- Reconcile the source inventory during an update rather than claiming perpetual or whole-cloud coverage.
- Keep Go 1.27, direct encoding/json/v2, AWS-style calling conventions, explicit credential providers, source ownership, safe errors and licensed bilingual documentation.

### Baseline and remaining work

- The planning-time baseline discovered four actions and emitted only AssumeRole.
- OIDC/SAML report DSL_PROTOCOL_PROFILE (Anonymous RPC and doRPCRequest); GetCallerIdentity reports DSL_OPERATION_SIGNATURE (no request-model argument).
- Fix the underlying frontend/IR/backend/runtime contracts under separate issues.
- Do not handwrite those operations, silently sign anonymous requests, introduce implicit key discovery or edit generated outputs.
- Anonymous authentication is an explicit per-operation protocol contract, not permission to omit credentials from signed operations.

- Accepted baseline: #51 full-DSL provider/cache composition, #53 STS-first guidance and explicit source registration, and #55 real 900-second expiry renewal with three STS issuances/four generated ECS reads and complete temporary-IAM cleanup.
- Those closed issues remain delivered evidence; do not recreate their cloud setup or repeat the same live run merely to populate a new milestone.

### Delivery order and evidence

1. Support requestless RPC signatures and generate GetCallerIdentity, preserving
   operation names, context, Options, mock seam, exact complete responses and metadata.
2. Support reviewed Anonymous RPC/doRPCRequest lowering and generate OIDC/SAML;
   preserve exact fields and omission, avoid signing/provider retrieval and redact
   request tokens/assertions as well as response credentials in default formatting.
   Items 1 and 2 can proceed independently on separate issue branches.
3. After both, accept the four-operation generated product: reproducible offline
   wire/response/negative contracts, provider/cache regression, a pinned external
   consumer and official-v2 comparison, authorized GetCallerIdentity live evidence,
   and a real source-update rehearsal with reviewed drift. Existing #55 remains
   the scoped live AssumeRole/renewal evidence unless a relevant change invalidates it.
4. Prepare compatibility/migration notes, licenses, English/Chinese release notes
   and the versioned pkg.go.dev inspection. Publish only after required prepublication
   gates; close the milestone only after publication/indexing evidence is recorded.

| Gate                  | Required evidence                                                                                                                                                          |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Generation            | Four discovered/lowered/emitted actions; complete reachable models; deterministic frontend/IR/Go checks; unsupported selected behavior fails before writes                 |
| Behavior              | All four compile and have independent offline protocol/field/response/error/cancellation tests and external Examples; signed/anonymous separation and sensitive formatting |
| Credentials           | Explicit providers; generated AssumeRole -> provider/cache -> consumer; existing ownership/concurrency/cancellation contracts and scoped #55 live renewal                  |
| Consumer/maintenance  | Version-pinned external STS workloads and official-v2 examples; docs-only independent Go-user task evidence; reviewed real upstream revision rehearsal                     |
| Live scope            | AssumeRole/renewal #55 plus authorized GetCallerIdentity; SDK/CLI and Explorer browser evidence recorded separately                                                        |
| Documentation/release | English pkg.go.dev comments, paired guides, licenses, Linux race/Windows CI, immutable v0.1.0 tag/release and same-version pkg.go.dev browser evidence                     |

- OIDC/SAML successful live federation requires separately authorized IdP/provider/ assertion fixtures.
- Those live cases are **outside this first release's required live scope**, declared before execution; mark them NOT RUN and disclose the limit in the release notes.
- All four actions' offline correctness remains mandatory.
- The user's subsequent #68 correction makes native default configuration/Profile/OAuth renewal required before publication; follow [the updated route](default-configuration.md).
- Federation credential-provider helpers remain separate future scope.
- STS has no pagination/waiter workload here; preserve shared engines without inventing STS pagination.
- ECS is only the existing role-credential consumer; ECS/VPC product-wide acceptance and benchmark #20 remain independent.

### GitHub coordination

- Use a GitHub milestone named `v0.1.0` and a repository-linked Projects v2 project named `alicloud-go-sdk-x: v0.1.0 STS`.
- Milestone membership identifies release scope; issues contain acceptance/dependencies/evidence; Project Status tracks execution.
- Keep one status label per issue: ready prerequisites -> status:ready/Todo; active work -> status:in-progress/In Progress; unmet issue dependencies -> status:blocked/Todo with named prerequisites; delivered acceptance -> status:done/Done.
- The Project is maintained when issue state changes; this change does not claim automatic synchronization.
- Never mark release Done because only code is merged.

- The parent tracks child completion and remains open through the release gate.
- Children refer to the parent as membership, not as a blocking prerequisite, keeping the dependency graph acyclic.
- Prior closed STS evidence joins the milestone/Project; historical generation/foundation issues and unrelated projects retain their scope.
- No arbitrary deadline is imposed.
- Read this document before choosing the next v0.1.0 issue.

- [Milestone v0.1.0](https://github.com/rambow-cloud/alicloud-go-sdk-x/milestone/4) | [Project](https://github.com/orgs/rambow-cloud/projects/3) | [Parent #57](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/57)

| Issue | Work                                                                       | Blocking prerequisites         |
| ----- | -------------------------------------------------------------------------- | ------------------------------ |
| #58   | Requestless GetCallerIdentity generation                                   | Accepted generator baseline    |
| #59   | Anonymous OIDC/SAML RPC generation                                         | Accepted generator baseline    |
| #60   | Four-operation acceptance, consumer comparison and source-update rehearsal | #58, #59; accepted #51/#53/#55 |
| #61   | Scoped release and versioned pkg.go.dev evidence                           | #60                            |

- Next implementation: #58, then #59; no API is handcrafted to bypass either gate.
- The parent aggregates delivery and is not a prerequisite that blocks its children.

- #58/PR #63 and #59/PR #64 are merged: all four pinned STS actions emit with native signed/anonymous separation. #60 delivers consumer/real-source/live-identity evidence and the independent developer handoff; keep its UX gate open until the user-arranged Go developer records actual results. #61 publication/indexing follows that required acceptance.
- Historical planning counts below describe the earlier baseline, not current coverage.
- See [acceptance evidence](sts-v010-acceptance-report.md).
