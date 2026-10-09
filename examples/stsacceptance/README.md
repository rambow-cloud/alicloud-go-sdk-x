# STS acceptance workload

[中文](README.zh-CN.md)

- From this directory, with Go 1.27+: `go test -v ./...`, then `go run .`.
- This separate consumer module pins official STS v2.1.0 and uses `replace ../..` for the SDK checkout.
- Record the exact root Git commit.
- Scripted HTTP responses use fixed fictional identities/keys; no operation reaches the network or creates resources.
- Initial dependency downloads may require network access.
- Keep comparisons outside the SDK core; see [tasks and acceptance rules](../../docs/sts-consumer-acceptance.md).

- Both SDKs execute four native actions.
- The generated SDK additionally composes its role provider/cache with two signed consumer reads; no application response translation or refresh loop is needed.
- This workload does not evaluate the official credentials library's role provider/refresh capabilities, performance or independent UX.
- Those cannot be inferred from successful fixture execution.
- Official v2.1.0 methods here lack context arguments and return Body envelopes; this is a pinned version observation, not a claim about every official SDK version.

- `consumer_review_test.go` is a separate external package using only public SDK imports and fresh fixtures.
- Run just this supplemental review with `go test -count=1 -v -run '^TestConsumerReview' ./...`.
- It checks caller option/input isolation, in-flight cancellation, 24 concurrent role-signed reads with one shared issuance, narrow mocks, anonymous provider/token isolation and non-retrying errors.
- See [the supplemental report](../../docs/sts-consumer-review.md).
- This agent-authored review is separate from the independent developer task result.

- `TestExternalDefaultProfileWorkload` exercises the new LoadDefaultConfig -> native temporary CLI Profile -> generated STS consumer path.
- The independent identity task now includes [default configuration](../../docs/default-configuration.md) and uses a temporary fictional profile; no interactive login is required for this kit.

- #83 adds `TestOfficialRPCShrinkJSONHelperParity`: generated ECS JSON query values are compared with the existing pinned OpenAPI v2.1.13 helper. It covers nil/empty values, nesting, non-ASCII text and int64 precision. This is account-free helper parity, not all-action ECS compatibility; no new dependency is added.

- #85 additionally compares pinned official simple-array and form-flattening helpers. This is offline helper parity, not full VPC compatibility or live acceptance.
