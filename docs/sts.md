# STS AssumeRole

[中文](sts.zh-CN.md)

- For the complete service/sts client, use NewAssumeRoleProviderFromClient; see the [full-DSL credentials guide](sts-credentials.md).
- The constructor below retains the older services/sts reference bridge contract.

- `sts.New` creates a typed AssumeRole client using separate source credentials. `sts.AssumeRoleAPI` permits a small fake.
- Input validates session syntax, minimum duration, external/source identity and policy JSON; authorization and role-specific maximum lifetime remain service decisions.
- Optional ExternalID supports confused-deputy protection.
- Responses parse Expiration as RFC3339 and redact credential formatting. `stscreds.NewAssumeRoleProvider` copies input/options and validates nonempty keys/token plus future expiration.
- Wrap it in `credentials.NewCache` for synchronized early refresh.
- Keep the STS source provider separate from this assumed-role provider to avoid recursive retrieval; the helper never changes the source.
- Context failures remain inspectable.
- AssumeRole is conservatively non-idempotent, so Standard does not retry token issuance.
- Only this STS operation is covered; no real-account validation is claimed.
- Protocol: [official AssumeRole reference](https://help.aliyun.com/zh/ram/developer-reference/api-sts-2015-04-01-assumerole).

- The operation/models/interfaces are generated; prose-only validation stays handwritten. [Generated field guide](generated/sts.md).
