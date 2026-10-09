# External temporary credentials

[中文](external-credentials.zh-CN.md)

- Issue: #90. Runtime dependencies remain standard library only.
- Constructors validate configuration without HTTP, subprocesses or token retrieval.
- Register explicit providers with `config.WithCredentialsProvider`, or use the native modes below.
- Wrap direct providers in `credentials.Cache`; default loading and profile providers already cache them.
- Cache refreshes share one deadline. Canceling a waiter does not cancel other waiters.

## Sources

| Provider | Native profile | Retrieval and renewal |
| --- | --- | --- |
| `externalcreds.URIProvider` | CredentialsURI: `credentials_uri` | GET a configured HTTP(S) broker; require complete STS and future expiration. |
| `externalcreds.ProcessProvider` | External: `process_command` | Execute copied argv without a shell; accept the native CLI StsToken output schema. |
| `externalcreds.ECSMetadataProvider` | EcsRamRole: optional `ram_role_name` | Fixed IMDS origin; obtain an IMDSv2 token, discover one role if needed, then retrieve STS. |
| `stscreds.AssumeRoleWithOIDCProvider` | OIDC: `ram_role_arn`, `oidc_provider_arn`, `oidc_token_file` | Reload token file and use the generated anonymous STS API. |
| `externalcreds.CloudSSOProvider` | CloudSSO: native login/account/access configuration fields | Reuse valid native STS or exchange a current portal access token. |

## Discovery

- Order: explicit provider → explicit profile → temporary environment keys → OIDC environment → URI environment → automatic profile → ECS IMDSv2 when the default file is absent.
- URI environment variable: `ALIBABA_CLOUD_CREDENTIALS_URI`.
- Optional metadata role: `ALIBABA_CLOUD_ECS_METADATA`.
- `ALIBABA_CLOUD_DISABLE_EXTERNAL_PROCESS=true` or `1` rejects discovered URI/External sources.
- `ALIBABA_CLOUD_ECS_METADATA_DISABLED=true` or `1` rejects discovered metadata sources.
- Explicit provider registration is deliberate opt-in and does not inspect these discovery flags.
- Missing explicit files/profiles, malformed selected sources and incomplete environment sources fail. They do not fall through to metadata.
- Metadata lookup is lazy; loading configuration does not probe a local machine or invent a region.

## Bounds and ownership

- External retrieval defaults: five seconds total, one MiB response/stdout. Explicit provider options may set smaller or larger bounds, up to 16 MiB.
- All metadata requests share that deadline. Session token TTL request: 21,600 seconds.
- Metadata never falls back to IMDSv1. Native HTTP transports are copied with proxy use disabled for metadata.
- Concrete `http.Client` values are copied with redirects disabled. Custom clients must honor context, prohibit redirects, and avoid proxying metadata tokens.
- HTTP and process errors hide URLs, arguments, stdout, stderr and response bodies. `errors.Is` preserves cancellation/deadline identity.
- Providers copy scalar configuration and argv; token providers/transports remain shared and must support concurrency.
- Use a trusted URI or executable. URI permits HTTP for local/private broker compatibility; configure HTTPS for remote brokers.
- Processes receive no stdin. Stderr is discarded, stdout is bounded, and the direct process is killed at its deadline. Inherited pipes have a bounded wait; process trees are not managed.
- `ParseCommand` supports quoted arguments and literal Windows paths. It does not expand shell syntax or environment variables.
- Process output must be a native `StsToken` profile with Unix-second `sts_expiration`. Recursive profile/process discovery is rejected. AK requires explicit `ProcessOptions.AllowLongLived` or profile `Options.AllowLongLived`; its zero value rejects AK.

## CloudSSO sessions

- Login once using the CLI's CloudSSO configuration flow. The SDK does not start a browser.
- Native fields: `cloud_sso_sign_in_url`, `cloud_sso_account_id`, `cloud_sso_access_config`, `access_token`, `cloud_sso_access_token_expire`.
- Portal exchange: POST `/cloud-credentials` with bearer access token and `AccountId`/`AccessConfigurationId` JSON.
- Complete `CloudCredential` or legacy top-level credential responses are accepted; ambiguous envelopes, rejected codes and inconsistent expiry fields fail.
- Profile renewal reloads login state. Changed account/portal/access configuration returns `ErrConfigurationChanged`.
- Valid native STS state may outlive the login token and is reused. Once renewal needs an expired login token, `ErrLoginRequired` requires CLI re-login.
- Renewed STS is cached in memory. CloudSSO configuration/login tokens are not written or refreshed by the SDK. OAuth persistence remains separate.

## Verification and sources

- Offline fixtures cover exact HTTP sequence/headers, no IMDSv1 downgrade, proxy isolation, lazy construction, cache reuse, token reload, changed/expired CloudSSO sessions, output bounds, redirects, cancellation, command ownership and explicit AK opt-in.
- A deterministic package Example uses an injected broker; it needs no network or account.
- Live IMDS, URI, External and CloudSSO accounts were not used for this change. Live federation remains #94.
- Native fields and CloudSSO portal protocol: CLI commit `fa14dd7b0359b5be229f9d770a662a86e69c13c1`, [profile](https://github.com/aliyun/aliyun-cli/blob/fa14dd7b0359b5be229f9d770a662a86e69c13c1/config/profile.go), [CloudSSO refresh](https://github.com/aliyun/aliyun-cli/blob/fa14dd7b0359b5be229f9d770a662a86e69c13c1/cloudsso/refresh.go).
- URI and IMDS protocol cross-check: credentials-go v1.4.5, [URI](https://github.com/aliyun/credentials-go/blob/v1.4.5/credentials/providers/uri.go), [ECS](https://github.com/aliyun/credentials-go/blob/v1.4.5/credentials/providers/ecs_ram_role.go).
- [Default configuration](default-configuration.md), [federation credentials](federation-credentials.md).
