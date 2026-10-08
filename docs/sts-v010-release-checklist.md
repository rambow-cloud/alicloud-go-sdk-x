# v0.1.0 release checklist / v0.1.0 发布清单

## English

#61 remains open: required independent #60 acceptance is NOT RUN. Publication and
same-version browser indexing are NOT RUN. Notes/commands are prepared; no tag is
created by this preparation. User authorization to complete the release already
exists; the pending item is actual acceptance evidence, not another permission request.

The user's later #68 correction also requires default configuration/native CLI
Profile/OAuth and its scoped evidence before the new #60 human handoff. Track
docs/acceptance/profile-oauth-live.json separately from historical manual snapshots.

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

These are post-acceptance commands, not executed preparation steps. Never overwrite
an existing tag. If publishing fails after tag creation, inspect actual remote state
and finish the release for that same tag; do not recreate or move it.

4. Record tag target/release URL and open these exact same-version pages in a browser:

| Package        | Exact URL                                                                                | Inspect                                                                                     |
| -------------- | ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| STS            | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/service/sts          | Version v0.1.0; Apache license; four client actions, model/field docs and runnable Examples |
| Credentials    | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/credentials          | Version; MIT; explicit providers including AnonymousProvider, Cache and Examples            |
| STS helper     | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/feature/stscreds     | Version; MIT; full-DSL constructor/ownership/renewal docs and Examples                      |
| Config         | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/config               | Version; MIT; default loader/options/precedence and offline Example                         |
| Native Profile | https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x@v0.1.0/feature/profilecreds | Version; MIT; native modes/OAuth persistence/ownership/errors and offline Example           |

If absent, use the browser Request action, then record the actual result/time;
do not claim indexing from local tests, curl or DNS. In particular check that the
indexer can build Go 1.27/direct JSON v2; incompatibility keeps indexing NOT RUN/FAIL
until resolved. Preserve our chosen Go baseline rather than quietly downgrading it.

5. Add actual release/browser evidence to this checklist or an explicitly linked
   paired report. Close #61/#57 and milestone 4 only when these required cases pass;
   align issue status labels and Project 3. A merged preparation PR does not satisfy
   publication/indexing or close either issue.

## 中文

用户后续 #68 修正还要求在新的 #60 人员交接前完成默认配置/原生 CLI Profile/OAuth 及
限定证据；docs/acceptance/profile-oauth-live.json 与历史手动快照分别记录，发布前必需。

#61 保持开放：#60 必需独立验收 NOT RUN，发布/同版本浏览器索引也 NOT RUN。说明/命令
已准备，此变更不创建标签。用户已授权完成发布，等待的是实际验收证据，不再次请求许可。

1. 接收用户安排的独立 Go 开发者完成的上述任务报告，仅将真实结果转写到
   `docs/acceptance/sts-independent-result.json`，保留脱敏源码/耗时/障碍/版本。全部必需
   项通过才完成 #60。
2. 审核合并最终说明/证据，准确 main 候选提交需要 Linux Go 1.27 race、Windows、
   automation/issue 及隔离消费者/真实来源演练通过；核 module/JSON v2、MIT/Apache 和
   固定来源。上述发布检查拒绝未完成的人验收或之后的 SDK 变化，不创建标签/发布。
3. 确认远端 v0.1.0 标签不存在、#60 已关闭，记录准确 SHA/CI URLs，验收后才执行共享
   命令创建/推送不可变注释标签并发布。禁止覆盖；标签创建后发布失败须检查远端状态，
   完成同标签发布，不重建或移动。
4. 记录 tag 目标/release URL，在浏览器打开英文表内准确同版本 STS/credentials/stscreds
   及 config/profilecreds 页面，核 v0.1.0、对应 Apache/MIT、导出 API/字段说明及 Example。缺失则使用 Request
   并记录真实时间/结果；不从本地检查/curl/DNS 宣称索引。尤其确认索引器支持 Go 1.27/
   直接 JSON v2，不支持时保持未跑/失败并解决，不静默降低版本。
5. 将实际发布/浏览器证据写入本清单或明确链接的双语报告；全部必需通过后关闭 #61/#57/
   milestone 4 并同步标签/Project 3。合并准备 PR 不等于发布或索引，不关闭 issue。
