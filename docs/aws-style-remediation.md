# AWS-style review remediation

[中文](aws-style-remediation.zh-CN.md)

- This path precedes implementation of the review on commit abcf961.
- Use AWS Go SDK v2 calling conventions with Alibaba OpenAPI operation names, wire fields, token values and page semantics.
- Preserve Go 1.27, JSON v2, explicit credentials, no-retry defaults, thirty-second operation timeout, HTTPS rules and standard-library core dependencies.
- This is an independent SDK; neither upstream SDK's source compatibility is promised.

| Order | Issue scope                                                | Acceptance                                                                                                                                                                                                     |
| ----- | ---------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1     | runtime attempt isolation and interrupted response retries | no stale output/successful incomplete short circuit; bounded idempotent EOF/read retries; cancellation/non-idempotent/JSON failures unchanged                                                                  |
| 2     | typed middleware and service Options                       | Initialize -> Serialize -> Build once; Finalize/Deserialize per attempt; owned typed input/output; NewFromConfig and concrete service Options; call retry/transport overrides stay isolated                    |
| 3     | pagination and generator policy collections                | current page delivered on repeated continuation, configurable stop; overflow-safe pages; dedicated options and NextPage call options; multiple token-only/page-only/dual adapters without invented wire fields |
| 4     | reusable waiter API                                        | inputs on Wait/WaitForOutput, independent concurrent waits, dedicated options/default acceptor override, all-ID safety and cancellation preserved                                                              |

- Actual GitHub issues were created before code, ordered #26 -> #27 -> #28 -> #29.
- Each issue gets its own branch/commit, tests and equivalent English/Chinese documentation.
- Regenerate through the emitter/templates; do not patch generated code.
- Close only after full local gates and exact-commit Linux race/Windows CI.
- Benchmarks, ROA/body, default discovery and changes to default retry/timeout policies are separate work.

- The review reproduced stale decoded output across retries, loss of a page when a token repeats, signed integer overflow in ECS legacy page traversal, and an idempotent response read ending in io.ErrUnexpectedEOF without retry.
- Regression fixtures must address those triggers directly, alongside generated clients and existing ECS/STS/VPC contracts.

- Migration notes must distinguish additive changes from early-v0 paginator/waiter changes.
- Keep the existing New(Config) convenience path; NewFromConfig accepts service functional options with validation.
- Token strings remain opaque; empty items alone do not end token pagination.
- Page-only services retain their native page fields.
- Dedicated paginator ClientOptions apply to each fetch and NextPage options override them for that fetch.
- Duplicate continuation defaults to safely stopping after returning the current page; an opt-out is explicit and may permit cycles.
- Failed/canceled fetches do not advance.
- Wait returns error; WaitForOutput returns the successful typed response.
- Waiter input is copied per invocation and per poll; overrides never mutate shared waiter configuration.

- References: [AWS paginator](https://github.com/aws/aws-sdk-go-v2/blob/main/service/ec2/api_op_DescribeInstances.go), [AWS waiter](https://github.com/aws/aws-sdk-go-v2/blob/main/service/ec2/api_op_DescribeInstanceStatus.go), [AWS middleware](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/middleware.html).
- Alibaba protocol evidence remains pinned under metadata/ with paired operation guides.
