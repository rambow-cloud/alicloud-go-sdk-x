# Sparse capability policies

[中文](README.zh-CN.md)

- Follow [#37 specification](../docs/capability-policy.md).
- Policies bind product/version/ source-lock hash and record evidence; the generator resolves exact wire paths against complete IR.
- They supply reviewed paginator/waiter/retry/client-token/validator/sensitive/ naming behavior, not per-field model declarations.
- Missing policy is allowed and means unreviewed/no Standard retry.
- Generated coverage records policy file hashes and individual capabilities.
- Do not edit generated adapters directly or infer support from names.

- Initial policies cover seven actions: four paginators, one reusable waiter, five idempotent reads, one conservatively non-retrying ClientToken write and STS sensitivity.
- This is not full capability or live acceptance.
- Evidence links are review records, not mutable network generation inputs.
- All generation and Examples run offline.
