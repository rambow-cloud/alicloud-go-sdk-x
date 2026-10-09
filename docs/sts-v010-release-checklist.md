# v0.1.0 release checklist

[中文](sts-v010-release-checklist.zh-CN.md)

- The user revised the route on 2026-10-09: #60 agent STS acceptance, ECS #74, VPC #75, then #61 publication/indexing.
- Follow [the current route](sts-ecs-vpc-path.md). Earlier STS-only publication and required-human #60 rules are superseded.
- Independent human UX is optional #76. Its original record remains NOT RUN; agent tests must never be labeled as human evidence.
- [v0.1.0 publication](releases/v0.1.0-publication.md) passed. Same-version browser indexing remains NOT RUN. STS closeout created no tag or release; publication followed ECS/VPC acceptance.

## Required gates

- [Fresh consumer evidence](release-consumer-refresh.md) records the integrated SDK pin and preserves historical records.

1. Record actual agent consumer results in acceptance/sts-agent-result.json with the tested SDK/workload revision. Preserve scoped live/source/Profile evidence and declare limits.
2. Complete #74 and #75 and their product-result records. Generation counts, empty terminal pages and skipped waiter cases do not satisfy unexecuted required cases. Review #47/PR #48 independently.
3. Require final-main Linux Go 1.27 race, Windows Go 1.27, automation and linked-issue checks. Review module/JSON v2, source/policy locks, MIT runtime and Apache generated LICENSE/NOTICE. Update paired release notes to the actual accepted scope.
4. Run the read-only guard. It checks STS agent tasks, both product reports, native OAuth/live/source evidence, clean main and changes since recorded revisions. It creates no tag or release. Missing/NOT RUN product evidence remains blocking.
5. Confirm #60/#74/#75 are closed and remote v0.1.0 is absent. Record the exact candidate and CI URLs, then use the commands below as part of #61.

## Publication commands (after gates)

```powershell
node .github/scripts/sts-release-check.cjs
git tag -a v0.1.0 <verified-main-commit> -m "v0.1.0: generated STS, ECS and VPC"
git -c credential.helper= -c 'credential.helper=!gh auth git-credential' -c credential.interactive=false push origin refs/tags/v0.1.0
gh release create v0.1.0 --verify-tag --title "v0.1.0: generated STS, ECS and VPC" --notes-file docs/releases/v0.1.0.md --prerelease
```

- Authorization for release completion already exists. The user now defers execution until ECS/VPC completion.
- Never overwrite a tag. If publishing fails after tag creation, inspect remote state and finish the same version.

## Same-version browser evidence

| Package              | Exact URL                                                                                |
| -------------------- | ---------------------------------------------------------------------------------------- |
| service/sts          | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/service/sts          |
| service/ecs          | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/service/ecs          |
| service/vpc          | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/service/vpc          |
| credentials          | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/credentials          |
| feature/stscreds     | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/feature/stscreds     |
| config               | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/config               |
| feature/profilecreds | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/feature/profilecreds |

- Open these pages in a browser. Check v0.1.0, licenses, complete documented actions/fields, reviewed adapters and runnable Examples.
- If absent, use Request. Record actual indexing/time; local tests, curl and DNS do not establish indexing.
- Verify the indexer accepts Go 1.27/direct JSON v2; do not silently downgrade the baseline.
- Finish #61/#57 and milestone only after actual publication/indexing. Keep labels and Project status aligned.
