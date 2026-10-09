# Release consumer evidence refresh

[中文](release-consumer-refresh.zh-CN.md)

- Issue: #61. Required product issues #60, #74 and #75 are closed.
- Refresh isolated STS, ECS and VPC consumers against the integrated SDK before publication. Preserve previous records and live/source pins.
- This is implementation-agent acceptance. Automated test time does not measure human usability; optional independent review #76 remains NOT RUN.

## Candidate and results

- SDK: `426108073437898ab93bc24fc3bc61311530d03e`; Go 1.27.1, Windows/amd64. This clean candidate integrates FC generation, the pinned OSS comparison and the JSON assertion correction. Final-main CI remains a separate gate.
- Both isolated modules: `go vet ./...`, `go test -json -count=1 ./...` and account-free `go run .` passed on 2026-10-09 21:19 UTC.
- STS: all five required tasks, eleven required tests and two additional protocol parity tests passed.
- ECS: ten required tests and fourteen subtests passed. VPC: eight tests and ten subtests passed.
- Current records: [STS](acceptance/sts-agent-result.json), [ECS](acceptance/ecs-product-result.json), [VPC](acceptance/vpc-product-result.json).
- Historical records: [STS](acceptance/sts-agent-result.historical.json), [ECS](acceptance/ecs-product-result.historical.json), [VPC](acceptance/vpc-product-result.historical.json). Earlier closeout guides describe those runs.
- The 16:04 UTC records are preserved unchanged: [STS](acceptance/sts-agent-result.2026-10-09T1604.json), [ECS](acceptance/ecs-product-result.2026-10-09T1604.json), [VPC](acceptance/vpc-product-result.2026-10-09T1604.json).
- Current generation uses IR schema 5 and source manifest SHA256 `fda7bc19fbbeb45b230c5dc9c4e69452a1dfe57331452f62dc653261b6d13254`. Product records identify this current corpus; historical source manifests remain in the preserved records.
- Live/source evidence retains its actual historical revisions. [Current existing-resource evidence](live-resource-followup.md) is separate; no cloud calls occurred during this consumer refresh.
- Generation coverage: 4 STS, 380 ECS and 403 VPC actions. This does not claim all-action live acceptance or complete capability review.
- FC has 72 generated offline actions and one binary exclusion. This refresh does not add FC consumer or live acceptance to the STS/ECS/VPC release scope.

## Remaining release gates

- Require unchanged accepted source/workload behavior and clean main in the read-only release guard.
- Require exact-final-main Linux race, Windows and automation CI; implementation PR checks do not replace this gate.
- Review source/policy locks, Go 1.27/direct JSON v2, MIT runtime and Apache generated notices, paired release notes and examples.
- The immutable experimental [v0.1.0 publication and browser inspection](releases/v0.1.0-publication.md) passed after the gates. All seven same-version pages passed the user's browser checks.
- Follow [the release checklist](sts-v010-release-checklist.md); close #61/#57 and the milestone only after actual indexing evidence.

- The 20:39 UTC records are preserved unchanged: [STS](acceptance/sts-agent-result.2026-10-09T2039.json), [ECS](acceptance/ecs-product-result.2026-10-09T2039.json), [VPC](acceptance/vpc-product-result.2026-10-09T2039.json). The new refresh follows PR #114 and changes no live/source result.
