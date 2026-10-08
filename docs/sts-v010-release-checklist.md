# v0.1.0 release checklist

[中文](sts-v010-release-checklist.zh-CN.md)

- #61 remains open: required independent #60 acceptance is NOT RUN.
- Publication and same-version browser indexing are NOT RUN.
- Notes/commands are prepared; no tag is created by this preparation.
- User authorization to complete the release already exists; the pending item is actual acceptance evidence, not another permission request.

- The user's later #68 correction also requires default configuration/native CLI Profile/OAuth and its scoped evidence before the new #60 human handoff.
- Track docs/acceptance/profile-oauth-live.json separately from historical manual snapshots.

1. Receive the user-arranged independent Go developer's completed
   [task report](sts-independent-result-template.md). Transcribe only actual results
   into `docs/acceptance/sts-independent-result.json`, retaining sanitized source/
   timings/obstacles/version evidence. Complete #60 only when all required cases pass.
2. Review and merge the final notes/evidence. At the exact main candidate commit,
   require Linux Go 1.27 race, Windows Go 1.27, automation/linked-issue checks plus
   isolated consumer and real-source rehearsal. Review module/JSON v2 and MIT/Apache
   licenses/source lock. `node .github/scripts/sts-release-check.cjs` refuses incomplete
   human evidence or SDK changes since that evidence; it creates no tag or release.
3. Confirm remote tag v0.1.0 is absent and #60 is closed. Record the exact candidate
   SHA and passing CI URLs. Create/push an immutable annotated tag and publish:

```powershell
git tag -a v0.1.0 <verified-main-commit> -m "v0.1.0: generated STS"
git -c credential.helper= -c 'credential.helper=!gh auth git-credential' -c credential.interactive=false push origin refs/tags/v0.1.0
gh release create v0.1.0 --verify-tag --title "v0.1.0: generated STS" --notes-file docs/releases/v0.1.0.md --prerelease
```

- These are post-acceptance commands, not executed preparation steps.
- Never overwrite an existing tag.
- If publishing fails after tag creation, inspect actual remote state and finish the release for that same tag; do not recreate or move it.

4. Record tag target/release URL and open these exact same-version pages in a browser:

| Package        | Exact URL                                                                                | Inspect                                                                                     |
| -------------- | ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| STS            | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/service/sts          | Version v0.1.0; Apache license; four client actions, model/field docs and runnable Examples |
| Credentials    | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/credentials          | Version; MIT; explicit providers including AnonymousProvider, Cache and Examples            |
| STS helper     | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/feature/stscreds     | Version; MIT; full-DSL constructor/ownership/renewal docs and Examples                      |
| Config         | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/config               | Version; MIT; default loader/options/precedence and offline Example                         |
| Native Profile | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/feature/profilecreds | Version; MIT; native modes/OAuth persistence/ownership/errors and offline Example           |

- If absent, use the browser Request action, then record the actual result/time; do not claim indexing from local tests, curl or DNS.
- In particular check that the indexer can build Go 1.27/direct JSON v2; incompatibility keeps indexing NOT RUN/FAIL until resolved.
- Preserve our chosen Go baseline rather than quietly downgrading it.

5. Add actual release/browser evidence to this checklist or an explicitly linked
   paired report. Close #61/#57 and milestone 4 only when these required cases pass;
   align issue status labels and Project 3. A merged preparation PR does not satisfy
   publication/indexing or close either issue.
