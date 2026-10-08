# Credentials

[中文](credentials.zh-CN.md)

### Default configuration and local sign-in

- Use config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile("oss-sftp")) to discover native CLI OAuth/temporary credentials and region, then pass the returned Config to a generated NewFromConfig.
- Native OAuth refresh/exchange uses cached, bounded providers; interactive login stays with the CLI.
- See the authoritative [default configuration contract](default-configuration.md) for precedence, supported modes, session ownership and explicit long-lived opt-in.
- This user-approved #68 scope supersedes the previous blanket no-default-chain/no-native-Profile/OAuth limitation.

### STS-first application configuration

- Prefer a renewable `stscreds.AssumeRoleProvider` wrapped in `credentials.Cache` for application clients.
- The full generated-client composition is in [the STS guide](sts-credentials.md) and the runnable README/ExampleAssumeRoleProvider.
- Configure a separate source provider for STS to avoid recursive retrieval.
- Cache coalesces refresh; a copied STS token in StaticProvider never renews itself.

- `alicloud.Config` and service Options expose only `CredentialsProvider`, never bare AccessKey/KeySecret/SecurityToken fields.
- Explicit provider injection is required even when environment credentials are populated and Config is constructed directly.
- Nil/typed-nil providers, including a nil ProviderFunc, fail during runtime/chain/cache construction before HTTP; construction does not retrieve credentials.
- Per-operation configuration replacement applies the same validation without changing the client.

| Source                    | Explicit declaration                                                                               | Intended use                                                                   |
| ------------------------- | -------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| Renewable role            | `stscreds.NewAssumeRoleProviderFromClient(...)`, then `credentials.NewCache(...)`                  | Primary application path; generated STS -> role provider -> generated consumer |
| Long-lived AK/SK          | `credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: id, AccessKeySecret: secret})` | Secondary deliberate opt-in; for example an authorized source for STS          |
| Static temporary snapshot | `NewStaticProvider` with SecurityToken and ExpiresAt                                               | Already-issued temporary credentials; expiration checked, never renewed        |
| Environment               | `credentials.EnvProvider{}`                                                                        | Explicit environment lookup, never auto-registered                             |
| Custom provider           | A `credentials.Provider` implementation or `ProviderFunc`                                          | Application-managed source; preserve context/concurrency/error contracts       |

- All deliberate custom providers remain supported, including ones returning long-lived keys.
- STS-first is guidance and a provider-only configuration rule, not an STS-only runtime.
- AWS Go SDK v2 likewise accepts providers and documents an explicit [StaticCredentialsProvider](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html).
- Our LoadDefaultConfig provides documented Alibaba-native discovery, including CLI Profile/OAuth, while retaining deliberate long-lived opt-in.
- Its bounded supported chain is not a claim of source compatibility with AWS or every Alibaba CLI mode.

### Provider contracts

- `Provider` retrieves copied credential snapshots with a context and must be concurrency safe.
- NewStaticProvider copies explicit keys.
- A zero/nil StaticProvider returns ErrMissingCredentials; an already-canceled context takes precedence.
- Formatting redacts values, but exported fields remain sensitive and must not be logged directly.

- EnvProvider reads ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET and optional ALIBABA_CLOUD_SECURITY_TOKEN each time.
- All three absent returns ErrNotFound; any incomplete configuration returns ErrMissingCredentials.
- NewChain copies an explicit ordered provider list, skips only ErrNotFound and stops on other failures or invalid snapshots.
- Chain itself discovers no sources; config.LoadDefaultConfig registers its documented temporary/profile providers.
- Automatic process/metadata discovery is outside this scope.
- Use ProviderFunc for a custom source and preserve cancellation errors.
- See [cache contracts](credential-cache.md).

- Issue [#53](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/53) verifies the provider-only rule, no environment fallback, constructor/per-call rejection and offline role signing/cache reuse.
- Synthetic Examples and CI do not establish live role renewal or overall Beta acceptance.
- Separate [live STS evidence #55](live-sts-renewal.md) now records real issuance/reuse, forced refresh, natural expiry renewal and cleanup with a dedicated minimal-permission role; it does not add native Profile/OAuth support.

- Explicit `credentials.AnonymousProvider{}` is a marker for reviewed anonymous STS OIDC/SAML operations.
- Those operations never retrieve even a configured source provider.
- Signed operations reject the marker; nil/typed-nil remain invalid everywhere.
- No implicit key discovery is added.
- See [protocol contracts](sts-anonymous-rpc.md).
