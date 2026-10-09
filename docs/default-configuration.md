# Default configuration and native profiles

[中文](default-configuration.zh-CN.md)

## Quick start

- Log in once with `aliyun configure --mode OAuth --profile oss-sftp`.

- Load the profile with `config.LoadDefaultConfig`.
- Pass the result to `sts.NewFromConfig`.

- This example makes a live, read-only identity request.
- It prints only the HTTP status.
- Package Examples use offline fixtures.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/rambow-cloud/alicloud-go-sdk-x/config"
    "github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

func main() {
    ctx := context.Background()
    cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile("oss-sftp"))
    if err != nil { log.Fatal(err) }
    client, err := sts.NewFromConfig(cfg)
    if err != nil { log.Fatal(err) }
    output, err := client.GetCallerIdentity(ctx, nil)
    if err != nil { log.Fatal(err) }
    fmt.Println(output.Metadata.HTTPStatusCode)
}
```

## Configuration

- `config` and `feature/profilecreds` use only the standard library.

- Loader options: `WithRegion`, `WithCredentialsProvider`, `WithSharedConfigProfile`, `WithSharedConfigFile`, `WithHTTPClient`, `WithCredentialsCacheOptions`.

- Default file: `~/.aliyun/config.json`, resolved with `os.UserHomeDir`.

- Supported modes: OAuth, StsToken, AK, RamRoleArn, ChainableRamRoleArn.

- AK and AK-based role profiles require an explicit `profilecreds.NewProvider` with `AllowLongLived: true`.
- Explicit StaticProvider/EnvProvider and custom providers also work.

- CloudSSO, process, URI and metadata modes return `ErrUnsupportedMode`.
- Complete OIDC environment configuration discovers a reloadable token file under #89. SAML assertion sources are registered explicitly. See [federation credentials](federation-credentials.md).

- Service constructors still require a provider.
- Loading reads local JSON and creates a cache; it makes no credential HTTP calls.

## Source order

1. Explicit credentials provider.

2. Explicitly selected profile.

3. Complete temporary environment credentials: ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET and ALIBABA_CLOUD_SECURITY_TOKEN.

4. Complete OIDC environment configuration: role ARN, OIDC provider ARN and token filename.

5. Profile selected by ALIBABA_CLOUD_PROFILE, then CLI `current`, then `default`.

- Present but incomplete environment credentials stop loading.
- Long-lived environment keys require explicit opt-in.
- Invalid sources are not silently skipped.

- Region order: WithRegion → ALIBABA_CLOUD_REGION_ID → ALIBABA_CLOUD_REGION → selected profile.
- The SDK does not invent a region.

- Missing files/profiles return `credentials.ErrNotFound`.
- Invalid JSON, duplicate profiles and role-source cycles return safe errors.
- File and OAuth response reads are limited to 1 MiB.

## Renewal and file writes

- Reuse valid STS credentials.
- Refresh an expired OAuth access token with POST /v1/token, then exchange it through POST /v1/exchange.

- Cache calls share one refresh with a deadline.
- Canceling one waiter does not cancel another waiter.

- Save rotated OAuth tokens before exchange; save renewed STS fields after exchange.
- This lets new processes and the CLI reuse the session.

- Change only the selected profile authentication fields.
- Preserve root, unknown and sibling-profile settings.
- Configuration settings remain constructor-time snapshots.

- SDK writers use a context-bounded file lock and same-directory file replacement.
- Detected external edits stop saving.

- Do not run CLI reconfiguration during renewal: the CLI does not use the SDK lock.
- A crash-left lock fails after the refresh deadline; the SDK does not steal locks.

- Initial login is interactive.
- Revoked, expired or invalid_grant sessions return `ErrLoginRequired`; inspect it with errors.Is.
- Default errors omit tokens, bodies and URLs.
- No browser or process is launched.

- Role profiles use the generated STS helper/cache with separate source credentials.
- Inputs are copied.
- Default session name: alicloud-go-sdk-x; absent duration is omitted.

## Protocol and evidence

- Protocol source: official CLI commit `fa14dd7b0359b5be229f9d770a662a86e69c13c1`, [profile fields](https://github.com/aliyun/aliyun-cli/blob/fa14dd7b0359b5be229f9d770a662a86e69c13c1/config/profile.go) and [refresh/exchange](https://github.com/aliyun/aliyun-cli/blob/fa14dd7b0359b5be229f9d770a662a86e69c13c1/config/configure.go).
- Use the CN/INTL endpoints and public client IDs.

- Live CN exchange uses AccessKeyId/AccessKeySecret/SecurityToken/Expiration.
- The pinned CLI uses camelCase tags with legacy case folding.
- JSON v2 accepts only these two complete schemas; mixed, duplicate and ambiguous fields fail.

- [Native live report](acceptance/profile-oauth-live.json): implementation `656ce39dda0b89ae743645ba1328974a937dc780`, 2026-10-08, 3 identity reads, 1 forced exchange, cache reuse, authentication-only persistence, session reconstruction and lock release: PASS. No CLI subprocess, new resources or published credentials/identities.

- Actual live refresh-token rotation and natural OAuth expiry wait: NOT RUN.
- The fresh access token was still valid.
- Offline rotation fixtures and [#55 role renewal](live-sts-renewal.md) are separate evidence.

- Offline tests cover source order, nil/typed-nil, explicit AK, cycles, ownership, role composition, concurrency, cancellation, token rotation, persistence and invalid HTTP/JSON/expiry responses.
- Doccheck, vet, tests and Linux race/Windows CI cover both packages.

- #68 replaced the old no-discovery rule in #53. #60 independent developer acceptance and #61 publication/indexing remain required.
- See [consumer tasks](sts-consumer-acceptance.md).
