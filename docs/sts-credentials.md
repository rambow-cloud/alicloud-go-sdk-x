# Full-DSL STS credentials

[中文](sts-credentials.zh-CN.md)

- Issue [#51](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/51) implements the offline composition part of [AC-05 / UX-04](product-acceptance.md). `stscreds.NewAssumeRoleProviderFromClient` accepts `service/sts.AssumeRoleAPI` and complete native input/options.
- The returned provider composes with credentials.Cache and generated clients without a response translator.
- #81 removes the bridge. NewAssumeRoleProvider accepts generated service/sts types; FromClient forwards to that implementation.
- [Reuse review #72](sts-reuse-review.md) separates shared helper rules from wire models and applies the same credential checks to both constructors.

### Generation boundary

- `service/sts` operations, models and protocol bindings come from official Darabonba DSL, the official semantic parser and complete IR.
- This handwritten adapter turns generated responses into `credentials.Credentials`. It implements no signing or HTTP protocol.
- Shared internal role validation does not depend on the compatibility request type. Native inputs go directly to the generated API.
- Cache and native Profile role composition reuse this adapter. Use [LoadDefaultConfig](default-configuration.md) for source discovery; the adapter itself performs none.
- Both constructor spellings use service/sts types. Old imports and selective models are removed; see [migration](service-consolidation.md).

### Composition

- This is the primary application credential path.
- The README and external ExampleAssumeRoleProvider provide a complete offline STS-to-ECS program, including role signing assertions and cache reuse.
- A long-lived source is optional and must be explicitly constructed with credentials.NewStaticProvider; an environment source must be explicitly declared as credentials.EnvProvider{}.
- Custom sources remain supported.
- Config/Options never accept bare keys, discover a fallback or retrieve credentials during construction; nil/typed-nil providers fail before requests.

```go
// Function body: the application supplies ctx and a separate source provider.
api, err := sts.NewFromConfig(alicloud.Config{
    Region: "cn-hangzhou", CredentialsProvider: source,
})
if err != nil { return err }
roleARN, session := "acs:ram::123456789012:role/example", "application"
provider, err := stscreds.NewAssumeRoleProviderFromClient(api, sts.AssumeRoleInput{
    RoleARN: &roleARN, RoleSessionName: &session,
})
if err != nil { return err }
cache, err := credentials.NewCache(provider, credentials.CacheOptions{})
if err != nil { return err }
client, err := ecs.NewFromConfig(alicloud.Config{
    Region: "cn-hangzhou", CredentialsProvider: cache,
})
if err != nil { return err }
_, err = client.DescribeRegions(ctx, nil)
return err
```

- Import the root module as alicloud, credentials, feature/stscreds, service/sts and service/ecs.
- Replace the placeholder ARN only for authorized live use.
- Runnable external ExampleNewAssumeRoleProviderFromClient uses scripted HTTP and synthetic credentials with no account/network; the generated-consumer integration test covers both generated clients and expiry-driven rotation.

### Contracts

- Construction and each retrieval copy input pointers and option registrations.
- Caller/API changes and option-slice mutations do not persist.
- Both constructors copy each call's option slice, including when custom APIs mutate it.
- API/callback objects remain shared and must be concurrency safe; do not mutate during construction or retain callback objects.
- The provider is concurrency safe, has no automatic cache, redacts formatting and returns errors for nil/zero provider or invalid constructor arguments (including typed-nil API and nil callbacks).

- Reuse existing reviewed helper role/session, identity, policy JSON and duration validation only.
- Send the full-DSL snapshot without a legacy request/response conversion.
- Nil duration stays absent; explicit duration below 900 including zero fails.
- Other optional presence and int64 width are retained.
- Authorization, role maximum lifetime and additional rules remain server decisions; generated STS operation validation/policy is unchanged.

- Nil output/credential containers fail.
- Keys/token must be nonblank, otherwise ErrMissingCredentials.
- Missing/empty/expired expiration returns ErrExpired; malformed expiration has a bounded diagnostic without its value.
- Parse RFC3339 to UTC, require future expiry and return Source=sts.AssumeRole.
- API errors retain errors.Is/As identity; check cancellation/deadlines before and after the call.
- Never log credentials/raw bodies.

- Use separate source credentials for STS to avoid recursive retrieval.
- The source is not altered.
- Wrap with credentials.NewCache for bounded/shared early refresh per [cache contracts](credential-cache.md); canceling a caller does not cancel the shared refresh.
- AssumeRole remains non-retrying even with Standard.
- No Profile/OAuth discovery, new dependency, generated-file edit, live role execution or cloud write is included. #51 itself did not exercise live role/refresh.
- Independent UX-04 and overall Beta/release acceptance remain open.

- The separate [live renewal record #55](live-sts-renewal.md) now establishes scoped real issuance/reuse, forced Invalidate and automatic renewal after genuine 900-second expiration with generated ECS reads and full temporary-IAM cleanup.
- It does not replace independent UX-04 or overall Beta/release acceptance, and does not prove live background/concurrent refresh or native Profile/OAuth renewal.

### Evidence

- Local gates: doccheck, vet, full Go tests/Examples, product-check and formatting; Node frontend contracts precede Go checks.
- Tests cover input/response boundaries, omission/ int64, concurrent snapshots, structured/cancellation errors, cache cancellation isolation, signed source STS -> role cache -> signed generated ECS requests, virtual clock expiry rotation and 503 non-retry.
- Linux race/Windows evidence is recorded in CI separately; pending CI is not PASS.
- Future synthetic times do not prove live renewal.
