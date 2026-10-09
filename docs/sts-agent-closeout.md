# STS agent consumer closeout

[中文](sts-agent-closeout.zh-CN.md)

## Scope and reviewer

- Issue #60. The user selected implementation-agent consumer acceptance on 2026-10-09.
- This is agent execution, not independent human usability or an unaided docs-only task.
- Optional human follow-up #76 retains its original NOT RUN record. Do not substitute test duration for human task time.
- Follow [the revised route](sts-ecs-vpc-path.md): complete STS, ECS #74, VPC #75, then release #61.
- No new cloud calls, resource creation, source updates, tag or release are required here.

## Pinned consumer

- The SDK/workload commit and actual environment are recorded in [machine evidence](acceptance/sts-agent-result.json).
- The separate `examples/stsacceptance` module imports public SDK packages and pins official STS v2.1.0, OpenApi v2.1.13 and Tea v1.3.13.
- Its explicit local replace selects the checked-out SDK. Official comparison dependencies remain outside the runtime module.
- All fixtures use synthetic profiles/credentials and scripted HTTP. No transport opens a connection.
- Production operations/models still come from official Darabonba and complete IR. Credential adapters/cache/Profile compose over generated interfaces.

## Tasks

| Task                   | Required consumer evidence                                                                                                                  |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Configuration/identity | StsToken loading, OAuth refresh/rotation/persistence/reconstruction, explicit long-lived opt-in, options/Metadata/cancellation              |
| Provider/cache         | Native generated AssumeRole to reusable provider/cache, ownership, role-signed consumers, 24 concurrent reads and canceled-waiter isolation |
| Mocks/errors           | Small AssumeRoleAPI, errors.Is/As, pre-cancellation, non-retrying issuance and safe error formatting                                        |
| Anonymous operations   | OIDC/SAML native fields and token encoding, no source retrieval/signature, explicit anonymous marker and sensitive formatting               |
| Official comparison    | The pinned official-v2 four-action fixture and documented calling differences; no performance/superiority measurement                       |

- The external OAuth test expires only synthetic access-token state. It checks rotated refresh-token reuse, persisted settings and reconstructed cached credentials.
- These tests establish offline behavior. They do not turn live OAuth rotation, natural OAuth expiry or live federation NOT RUN into PASS.
- `.github/scripts/sts-consumer-record.cjs` records only successful required test events. Missing/duplicate/skipped/failed cases cannot become PASS; raw output is not included in the public record.

## Verification

```powershell
go -C examples/stsacceptance test -json -count=1 ./...
go -C examples/stsacceptance run .
node --test .github/scripts/*.test.cjs
node .github/scripts/check-doc-language.cjs
go run ./internal/cmd/doccheck
go vet ./...
go test ./...
```

- Node 22 frontend checks/tests precede the Go gates. CI also runs both generator checks, the source-update rehearsal, Linux race and Windows tests.
- Record final-head CI links on #60 and its PR. Use existing accepted source/live evidence where behavior is unchanged.
- The read-only release guard requires agent STS evidence and ECS/VPC product reports. It remains blocked while either product is NOT RUN.

## Results

- Initial #60 closeout: PASS, 11 external consumer tests across all five tasks, pinned to [3d54426](https://github.com/rambow-cloud/alicloud-go-sdk-x/commit/3d54426d15b6b79b238138a0d8a52bf9ebc0d62c).
- Environment: Go 1.27.1, windows/amd64; Node 22.21.1. Execution: 2026-10-09 00:14:07.135–00:14:09.855 UTC.
- The command took 2.720 seconds including compilation. Summed Go test elapsed time: identity 0.080 seconds, comparison 0.010 seconds, other groups rounded to zero by Go. These are automated execution times, not human task times or a benchmark.
- PASS: runnable external consumer, official STS v2.1.0 comparison, frontend check and 57 frontend tests, 25 automation tests, both generator checks, doccheck (18 public packages), vet, root Go tests and formatting.
- PASS: paired documentation and local links. No runtime, generated SDK, source, IR or capability-policy changes were needed.
- Expected release block: the read-only guard returns `BLOCKED: ecs product acceptance is not PASS`. ECS and VPC records remain NOT RUN.
- Final-head Linux race, Windows and automation results are recorded on the closing PR and #60 before merge. The machine record remains bound to the tested workload commit; later evidence-only changes do not alter it.
- Independent human usability remains NOT RUN under optional #76. Live federation and natural OAuth expiry remain outside this offline result.

- Shared-workload refresh for #75: all 11 cases PASS at 85795cf3cf1604afe59b0e8af03d5df0c9d3ba38; the machine record contains latest timestamps/timings. Initial #60 execution details above remain historical. Runtime/STS source behavior is unchanged; this refresh covers the new consumer/CI guard baseline.

- #81 service consolidation: all 11 cases PASS at e33e5f93d856027569969056e9d4c25e92e51212, using generated service/sts and the shared adapter. The updated machine record contains actual timings; earlier closeout/live evidence stays historical.

- #81 guide correction: discovery now precedes Go emission in the tool commands. All 29 consumer cases and 24 product subtests PASS at 29d468ad5f8999e92b4e60e4140cb8432efb8a13; machine records use this pin. Go/runtime/source/policy behavior is unchanged from the locally checked e33e5f93d856027569969056e9d4c25e92e51212 workload, so its unaffected gates are reused. Final-head CI remains required on PR #82.

- #83 shared RPC/runtime regression: 11 existing STS cases plus the pinned official JSON-helper comparison PASS at 01f25a9a571c3b59024dfbef7e2e96f7605f4a55. This is automated agent evidence; earlier live pins and independent UX limitations remain unchanged.

- #85 refresh: 11 existing STS consumer cases plus pinned official JSON/simple/form helper comparisons PASS at 0fa3989f5df2e809082126eb9f177a8cc5ec218c. Both consumer modules pass vet and account-free programs. Historical live and human evidence stays unchanged.
