# Native OAuth refresh validation

[中文](native-oauth-renewal.zh-CN.md)

- Issue: #94. Use the authorized local `oss-sftp` Profile; cloud calls are read-only.
- Establish this scope before executing the ignored local harness.
- Keep tokens in memory/native configuration; publish no credentials, identities or raw bodies.

## Forced refresh scope

- Use the real clock. Do not simulate passage of time or claim natural expiration.
- Invalidate only the selected profile's cached STS/access-token expiration timestamps under the SDK session lock. Keep token values and other configuration unchanged.
- Use the native SDK loader/provider to perform a real OAuth refresh grant, then a real STS exchange and generated GetCallerIdentity reads.
- Observe whether the server actually replaces the refresh token. No replacement means rotation is not demonstrated.
- Verify cache reuse, atomic persistence, unchanged non-authentication configuration, reconstructed-provider reuse and released session lock.
- If the refresh grant fails before any token update, restore the previous expiration metadata under the lock. Never restore old tokens after a successful server rotation.
- Raw transport data is not logged. The report contains only counts, booleans, source pin and sanitized statuses.

## Natural expiry follow-up plan

- Use the session issued by the successful refresh above. Its original STS expiry is 2026-10-09 19:19:56 UTC; its access-token expiry is 19:19:57 UTC.
- Wait for both timestamps to pass on the real clock. Do not change cached timestamps, token values or the provider clock.
- Require the original expiration values and configuration modification time before execution. If another process changes the session first, record that this session was not tested.
- Use a fresh native provider, one real refresh grant and STS exchange, then two generated identity reads. Check cache reuse, persisted session reuse, token replacement, unchanged other settings and lock release.
- Record the actual SDK commit and start/end times. Keep live OIDC/SAML separate; publish no credentials or identity values.

## Separate unfinished cases

- Natural OAuth expiration passed in the recorded follow-up below.
- OIDC/SAML live federation needs the role/IdP configuration the user will provide.
- A forced refresh does not complete the full #94 acceptance criteria.

## Recorded result

- PASS on 2026-10-09 16:19:58 UTC (2026-10-10 in Asia/Shanghai), Windows/amd64, Go 1.27.1.
- Tested SDK: `0365dd8733ffa6aa86061b226d5e8f79d83e70d0`.
- [Sanitized machine record](acceptance/native-oauth-refresh-live.json): one real refresh grant, one STS exchange and two generated identity reads.
- The server replaced the refresh token. Persistence and reconstructed provider reuse passed.
- Cache reuse, complete future-expiring temporary credentials, HTTP 200, one attempt per identity read, matching request IDs, unchanged non-authentication configuration and released session lock passed.
- No CLI subprocess or new cloud resources; the published record contains no credentials, identity values or raw bodies.
- At this forced-refresh run, natural expiration and live federation were NOT RUN.

## Natural expiry result

- PASS: 2026-10-09T19:20:10.1469286Z to 2026-10-09T19:20:10.9488128Z, SDK `fc9bef40aa2dfe2d51bebba67442f43ae08e2dc5`, go1.27.1, windows/amd64.
- The original session file was last modified at 2026-10-09T16:19:58.6162088Z; STS expired at 2026-10-09T19:19:56Z, access token at 2026-10-09T19:19:57Z. Both expired before retrieval on the real clock. No cache timestamps or clock were changed.
- One real refresh grant, one STS exchange and two generated identity reads passed. Server refresh-token replacement: PASS. Cache, atomic persistence, reconstructed provider reuse, unchanged other settings and lock release passed.
- [Sanitized machine record](acceptance/native-oauth-natural-expiry-live.json). No CLI subprocess, cloud provisioning, published credentials or identity values.
- Live OIDC/SAML still needs user configuration; #94 stays open.
