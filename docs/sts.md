# STS usage

[中文](sts.zh-CN.md)

- Import `service/sts`; the old `services/sts` package is removed in #81.
- Official Darabonba/parser -> complete IR -> shared Go backend generates AssumeRole, GetCallerIdentity, AssumeRoleWithOIDC and AssumeRoleWithSAML.
- Use `sts.NewFromConfig` and `stscreds.NewAssumeRoleProvider` with generated AssumeRoleInput. NewAssumeRoleProviderFromClient forwards to the same implementation.
- The reusable helper copies pointers/options, validates reviewed role/session/policy rules, parses RFC3339 expiration and rejects incomplete or expired credentials.
- Keep the source provider separate. Wrap role credentials in credentials.Cache for synchronized bounded renewal.
- Signed operations require an explicit or loader-selected provider; reviewed anonymous operations do not retrieve signing credentials. Construction performs no credential HTTP calls.
- Cancellation and structured service errors remain inspectable. AssumeRole issuance does not retry with Standard.
- See the [generated guide](products/sts.md), [role composition](sts-credentials.md), [consumer acceptance](sts-consumer-acceptance.md) and [migration](service-consolidation.md). Successful live federation remains outside the current accepted scope.
