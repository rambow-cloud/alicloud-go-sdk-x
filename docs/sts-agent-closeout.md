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

- Pending execution at the final pinned workload revision. No task success or time is inferred.
