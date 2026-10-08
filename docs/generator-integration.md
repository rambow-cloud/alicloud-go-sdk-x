# Generator review and main integration

[中文](generator-integration.zh-CN.md)

- The user-authorized review and merge completed on 2026-10-08 (Asia/Shanghai).
- The compatibility bridge, five product-generator stages and review fix are integrated into `main` at implementation merge `19e9107e876b2309808511ad0dbced633cb54d7d`.
- Issue [#33](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/33) tracks this acceptance and its documentation.
- Earlier branch/dependency descriptions in issue drafts and specifications describe implementation history, not outstanding merges.

| Issue | Reviewed PR                                                      | Reviewed head | Main merge | Passing CI                                                                                |
| ----- | ---------------------------------------------------------------- | ------------- | ---------- | ----------------------------------------------------------------------------------------- |
| #31   | [#32](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/32) | `92095fe`     | `33bfa4a`  | [37614002853](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37614002853) |
| #34   | [#39](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/39) | `292bb5d`     | `1ca9cf8`  | [37615648634](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37615648634) |
| #35   | [#40](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/40) | `7427e10`     | `59508cc`  | [37623726048](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37623726048) |
| #36   | [#41](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/41) | `4afa9ff`     | `1f1ed96`  | [37629756588](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37629756588) |
| #37   | [#42](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/42) | `629e810`     | `fd27fe9`  | [37635577018](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37635577018) |
| #38   | [#43](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/43) | `7098116`     | `5f507a5`  | [37645170995](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37645170995) |
| #44   | [#45](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/45) | `859f8f6`     | `19e9107`  | [37699342435](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37699342435) |

- Review covered source/import pins, strict semantic lowering, canonical representation normalization, complete discovery/reachable IR, batch typed RPC output, owned input snapshots and reconciliation, native capabilities, licensed prose and notices.
- Each PR has an English-primary review record with Chinese explanation.
- The role-alias finding in #44 is fixed by #45: conflicting roles within a model fail before rendering, while legal cross-model token names and fields shared by separate adapters remain valid.
- No additional merge-blocking findings remained in the reviewed pinned profile.

- Merge commits preserve issue commits.
- PRs were merged in dependency order, subsequent bases retargeted to `main`, and each resulting main file tree compared to its reviewed head.
- Every comparison matched; the implementation merge tree is `11e36c09e3493c40c5f07f8f998fe1ac946e9173`, identical to #45's verified head.
- Passing PR evidence includes Linux race, Windows, Node/frontend checks, Go tests/Examples, doccheck, vet and regeneration.
- Superseded main-push CI runs in this merge batch were canceled where still active; the final snapshot check was retained.
- Documentation-only integration updates receive their own bilingual/format checks and PR CI, without repeating unchanged local SDK tests.

- The accepted source route remains complete pinned official Darabonba DSL -> official semantic parser -> normalized IR -> our Go backend -> shared runtime.
- There is no per-operation metadata whitelist or Tea runtime dependency.
- Optional metadata enriches representations; sparse policies supply reviewed behavior, not every model/field.

| Product | Discovered | Emitted | Unsupported | Emitted models | English operation prose | English field prose |
| ------- | ---------- | ------- | ----------- | -------------- | ----------------------- | ------------------- |
| ECS     | 380        | 283     | 97          | 1,453          | 274/283                 | 2,401/6,476         |
| VPC     | 403        | 295     | 108         | 1,240          | 295/295                 | 3,482/6,489         |
| STS     | 4          | 1       | 3           | 5              | 1/1                     | 14/19               |
| Total   | 787        | 579     | 208         | 2,698          | 570/579                 | 5,897/12,984        |

- `service/` contains complete supported models; `services/` remains the five-operation reference bridge.
- Four native paginators, one ECS running waiter and seven reviewed operation policies are accepted; 572 emitted operations remain unreviewed for these capabilities.
- API wire names/native cursors remain Alibaba's, with AWS-style calling conventions.
- Retry stays opt-in and conservative.
- English pkg.go.dev comments, offline Examples, Apache terms/notices and paired usage/contracts/source guides are present.
- Missing prose and unavailable Chinese semantic translations are explicitly reported.

- This acceptance covers the pinned RPC profile, compilation and offline contracts.
- Broader protocols/products, further reviewed capabilities, benchmark #20, live Explorer evidence, version releases and pkg.go.dev indexing have separate acceptance.
- Source license provenance and release limits remain recorded in the source/release guides.
