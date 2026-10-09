# Release consumer evidence refresh

[中文](release-consumer-refresh.zh-CN.md)

- Issue: #61. Required product issues #60, #74 and #75 are closed.
- Refresh isolated STS, ECS and VPC consumers against the integrated SDK before publication. Preserve previous records and live/source pins.
- This is implementation-agent acceptance. Automated test time does not measure human usability; optional independent review #76 remains NOT RUN.

## Candidate and results

- SDK: `0365dd8733ffa6aa86061b226d5e8f79d83e70d0`; Go 1.27.1, Windows/amd64.
- Both isolated modules: `go vet ./...`, `go test -json -count=1 ./...` and account-free `go run .` passed on 2026-10-09 16:04 UTC.
- STS: all five required tasks, eleven required tests and two additional protocol parity tests passed.
- ECS: ten required tests and fourteen subtests passed. VPC: eight tests and ten subtests passed.
- Current records: [STS](acceptance/sts-agent-result.json), [ECS](acceptance/ecs-product-result.json), [VPC](acceptance/vpc-product-result.json).
- Historical records: [STS](acceptance/sts-agent-result.historical.json), [ECS](acceptance/ecs-product-result.historical.json), [VPC](acceptance/vpc-product-result.historical.json). Earlier closeout guides describe those runs.
- Live/source evidence retains its actual historical revisions. [Current existing-resource evidence](live-resource-followup.md) is separate; no cloud calls occurred during this consumer refresh.
- Generation coverage: 4 STS, 380 ECS and 403 VPC actions. This does not claim all-action live acceptance or complete capability review.

## Remaining release gates

- Require unchanged accepted source/workload behavior and clean main in the read-only release guard.
- Require exact-final-main Linux race, Windows and automation CI; implementation PR checks do not replace this gate.
- Review source/policy locks, Go 1.27/direct JSON v2, MIT runtime and Apache generated notices, paired release notes and examples.
- Publish immutable experimental v0.1.0 after the gates. Publication and same-version browser indexing remain NOT RUN in this record.
- Follow [the release checklist](sts-v010-release-checklist.md); close #61/#57 and the milestone only after actual indexing evidence.
