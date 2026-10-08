# Live STS renewal

[中文](live-sts-renewal.zh-CN.md)

### Scope and prerequisites

- The user authorizes local, unsandboxed live STS renewal validation and creation of a dedicated validation role on 2026-10-08.
- Use baseline f4cebddd3660d0d09d2b56e6319f5895c2f4d4a5, the explicit local Aliyun CLI profile oss-sftp and cn-hangzhou.
- The bounded CLI preflight confirms IdentityType=Account.
- An Alibaba Cloud account cannot call AssumeRole directly.
- Create a uniquely named RAM user with no console login, one temporary AccessKey and a policy permitting only sts:AssumeRole on one new role.
- The role trusts only that user and has a separate policy permitting only ecs:DescribeRegions.
- Keep authentication secrets/tokens in process memory, not command arguments or files; key identifiers are used only as deletion parameters when cleaning up.
- Record ownership in a private cleanup journal and delete the temporary key, attachments, user, role and two policies after the run.
- Do not change existing identities, permissions or resources.
- This user instruction overrides the earlier no-creation constraint for this dedicated setup only.

- First establish an issue and separate issue branch.
- Run one bounded CLI GetCallerIdentity preflight, processing output locally.
- Account identity stays in ignored private setup data; expose only safe identity type/status.
- Allow the CLI's existing OAuth refresh and check the management profile's cached expiration locally.
- Require its lifetime to cover setup/cleanup or stop for renewed login.
- Use the account profile only to provision/clean up the dedicated setup; the generated STS client uses the explicitly constructed RAM-user source.
- Record any authorization failure without broadening permissions or retrying unchanged credentials.

### Verification sequence

1. Configure generated service/sts with separate source credentials. Use
   NewAssumeRoleProviderFromClient with the authorized role/session/optional identity,
   DurationSeconds=900 and real time. Count live issuance calls without logging keys,
   tokens, headers, bodies, role/account/resource identifiers or raw errors.
2. Wrap with credentials.Cache. Retrieve a valid role snapshot and use it in generated
   ECS DescribeRegions. Verify the returned AssumeRole identity/metadata and
   repeated retrieval/read reuse without another issuance. An ECS permission failure
   is distinct from STS failure; never broaden permissions to make a check pass.
3. Invalidate the cache once and retrieve again. Verify a second live issuance, changed
   credential values by an in-memory boolean comparison, valid expiry and successful
   role-authenticated read. Report this as forced refresh, not natural renewal.
4. Leave the cache idle until the second snapshot's real server expiration plus a small
   margin. Wait in interruptible intervals of at most 60 seconds, reporting progress.
   Then retrieve/use credentials without manual invalidation, time substitution or a
   handwritten issuance loop. Verify another live issuance, a new future expiration,
   no expired snapshot served and successful generated-client role-authenticated reads.
5. Preserve separate source identity, input ownership and safe structured failures.
   Publish sanitized timing/count/status evidence, SDK/CLI/Go versions and commit.
   Keep the disposable harness and local-only reports in ignored .git/ storage.

- The pinned full-DSL STS package currently emits AssumeRole only, not GetCallerIdentity.
- Use the returned AssumedRoleUser ARN as the role-identity assertion, and generated ECS as the authenticated consumer.
- The CLI preflight is separate evidence; do not claim a generated GetCallerIdentity operation. [Role assumption prerequisites](https://www.alibabacloud.com/help/en/ram/user-guide/assume-a-ram-role) justify the dedicated RAM-user source and narrow sts:AssumeRole policy.

- The [official STS FAQ](https://www.alibabacloud.com/help/en/ram/support/faq-about-ram-roles-and-sts-tokens) specifies a 900-second minimum.
- Issuing another token does not revoke the earlier one.
- This experiment establishes only scoped live issuance/reuse/forced refresh/natural expiry renewal, not native Profile/OAuth renewal, every role/trust scenario, independent UX acceptance or overall Beta.
- Explorer browser verification remains separate: open [AssumeRole](https://api.aliyun.com/api/Sts/2015-04-01/AssumeRole) to inspect the exact role/session/duration fields; do not share returned credentials.

### Recorded evidence

- Issue [#55](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/55), 2026-10-08.
- SDK commit f4cebddd3660d0d09d2b56e6319f5895c2f4d4a5; Go 1.27.1 / Windows amd64; Aliyun CLI 3.4.11; cn-hangzhou.
- The run started at 2026-10-08T09:40:30.8039616Z.
- Source and consumer are separately configured explicit providers.
- No production code, generated output or dependency changed.

| Case                       | Result  | Evidence                                                                                                                                                                   |
| -------------------------- | ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CLI management preflight   | PASS    | Account identity; valid OAuth management snapshot; source lifetime covered setup/cleanup                                                                                   |
| Dedicated setup            | PASS    | One RAM user without console login, one temporary source key, one role, two narrow custom policies; exact-user trust readback                                              |
| Live issuance and role use | PASS    | Generated AssumeRole, matching AssumedRoleUser identity, keys/token/future expiry, HTTP 200/request-ID/one-attempt metadata, successful generated ECS read                 |
| Cache reuse                | PASS    | Repeated retrieval/ECS read retains one issuance                                                                                                                           |
| Forced refresh             | PASS    | One Invalidate triggers a second live issuance; changed keys/token; successful role-authenticated ECS read                                                                 |
| Natural expiry renewal     | PASS    | Idle until 2026-10-08T09:55:30Z + 5 seconds; next retrieval automatically obtains third live issuance; new expiry 2026-10-08T10:10:34Z; signed generated ECS read succeeds |
| Source/input ownership     | PASS    | Separate source and original input unchanged; total three issuances and four ECS reads                                                                                     |
| Cleanup                    | PASS    | Temporary key deleted, both policy attachments detached, user/role and both custom policies deleted; no test identities retained                                           |
| Explorer browser           | NOT RUN | SDK/CLI execution is separate; manual link above                                                                                                                           |

- Natural renewal completed at 2026-10-08T09:55:35.4615676Z; no virtual clock, further Invalidate or manual issuance loop was used.
- Forced refresh and natural renewal are separately observed.
- Live results do not establish early/background refresh, concurrent live refresh, every role/trust/condition, SDK-native Profile/OAuth renewal, independent UX or overall Beta.
- Offline contracts retain their own evidence.

- Disposable local harness SHA256 (not published as production SDK code): Go 40267af9bf8aa2afa07f708799c637690201bc557129477623522e2682efd0c6; PowerShell bac2ff7b24a8ee264da3f9c1106854766d34a31146361f15a617b2b80b949b12.
- Keys/secrets/tokens, account/role/user identifiers and raw bodies are excluded from this document and GitHub evidence.
