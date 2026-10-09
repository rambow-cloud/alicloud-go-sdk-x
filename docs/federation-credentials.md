# Renewable federation credentials

[中文](federation-credentials.zh-CN.md)

- Issue: #89. Uses the generated `service/sts` OIDC and SAML APIs from the pinned official Darabonba source.
- `feature/stscreds.NewAssumeRoleWithOIDCProvider` and `NewAssumeRoleWithSAMLProvider` adapt those APIs to `credentials.Provider`.
- Configure the exchange client with `credentials.AnonymousProvider{}`. Application clients use the returned temporary credentials.
- Wrap the provider in `credentials.Cache` for bounded shared renewal. The helper itself neither caches nor retries issuance.

## Token ownership and renewal

- Supply a concurrency-safe `TokenProvider`. It returns current material on every refresh.
- `TokenProviderFunc` accepts a context-aware function. Custom source errors retain their identity; the custom source owns their safe formatting.
- `NewFileTokenProvider` stores a filename without reading it. Retrieval reads a regular file up to 1 MiB and trims surrounding whitespace.
- Replace the file atomically when the external IdP rotates its token. The SDK never writes it or logs its contents.
- Supply the original OIDC token or Base64-encoded complete SAML response. No decoding is performed.
- Leave `OIDCToken`/`SAMLAssertion` nil in the helper input; combining embedded tokens and a source is rejected.
- Constructors copy inputs and option registrations. Shared APIs/sources/callbacks must support concurrency; callbacks must not retain request or option objects.
- A cached provider reads fresh material after expiration, early refresh or `Invalidate`. Cache invalidation does not revoke issued cloud credentials.

## Validation

- Role and identity-provider ARNs are required. OIDC also requires a role session name.
- An explicit duration below 900 seconds fails; nil preserves the service default. Maximum lifetime remains a service decision.
- Reviewed token length: OIDC 4..20,000 bytes; SAML 4..100,000 bytes. This helper does not validate JWT signatures or XML assertions locally.
- Responses require nonempty access key, secret, security token and future RFC3339 expiration. Invalid responses never enter the cache.
- Cancellation and API errors remain inspectable with `errors.Is`/`errors.As`.
- File/provider formatting omits paths and secrets; file errors omit raw operating-system messages.

## Default OIDC discovery

- Set `ALIBABA_CLOUD_ROLE_ARN`, `ALIBABA_CLOUD_OIDC_PROVIDER_ARN` and `ALIBABA_CLOUD_OIDC_TOKEN_FILE`.
- Optional `ALIBABA_CLOUD_ROLE_SESSION_NAME` defaults to `alicloud-go-sdk-x`.
- Order: explicit provider, explicit profile, temporary environment keys/token, OIDC environment, automatic native profile.
- A partial selected OIDC source stops loading; it never silently falls back.
- Loading performs no HTTP or token-file read. Exchange uses the public `https://sts.aliyuncs.com` origin; the application region follows the existing option/environment/profile order.
- Long-lived access keys still require explicit opt-in. SAML sources are registered explicitly because the SDK does not invent an assertion source.

## Examples and evidence

- Run `go test ./feature/stscreds ./config` for deterministic OIDC/SAML examples, token rotation, cache coalescing, anonymous wire requests, input ownership, invalid results, cancellation and discovery precedence.
- No external identity provider or cloud resource is needed for these tests.
- Real federation renewal belongs to #94. Offline success does not establish live federation acceptance.
- Protocol evidence: pinned `sources/darabonba/products/sts/main.tea`, [official OIDC operation](https://www.alibabacloud.com/help/en/ram/developer-reference/api-sts-2015-04-01-assumerolewithoidc) and [official CLI environment variables](https://www.alibabacloud.com/help/en/cli/environment-variables).
