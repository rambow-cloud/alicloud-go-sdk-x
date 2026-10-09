# STS consumer acceptance

[中文](sts-consumer-acceptance.zh-CN.md)

## Current #60 consumer acceptance

- The user now selects implementation-agent acceptance for #60. See [the closeout report](sts-agent-closeout.md) and machine evidence in acceptance/sts-agent-result.json.
- Use the pinned current SDK/workload revision; run all standalone tests and the runnable consumer.
- Native StsToken, OAuth refresh/rotation/persistence/reconstruction and explicit long-lived opt-in use synthetic profiles and scripted HTTP.
- Record agent execution, actual case results and automated test duration. Do not claim an independent human/docs-only task or superiority.
- The independent developer instructions below now apply to optional #76. Its original record stays NOT RUN.
- Release #61 follows ECS #74 and VPC #75 under [the revised route](sts-ecs-vpc-path.md).

- Issue #60 validates the four generated actions and consumer maintenance, separately from publication #61.
- The isolated `examples/stsacceptance` module pins the official STS v2.1.0 comparison; the SDK uses an explicit local replacement to this checked-out repository.
- Record `git rev-parse HEAD`, `go version`, OS and module versions in the report.
- Official comparison dependencies never enter the SDK runtime module.

- The implementation author's offline execution is technical evidence.
- The user has arranged an independent Go developer; their docs-only result remains required.
- Successful live federation is NOT RUN/outside release scope; never supply real credentials to this offline kit or create cloud resources.

### Independent developer handoff

- Check out the exact commit supplied with the handoff, use Go 1.27+, and read [STS composition](sts-credentials.md), [credentials](credentials.md), [the generated guide](products/sts.md) and [anonymous RPC](sts-anonymous-rpc.md).
- From `examples/stsacceptance`, run `go test -v ./...` and `go run .`.
- These run with scripted local HTTP fixtures and fixed placeholders.
- Dependencies may download on the first build; operation execution has no account/network access.

- In a separate scratch Go module, without implementation-author assistance:

1. Create a fictional native StsToken/OAuth CLI JSON profile in a scratch temporary
   file. Use config.LoadDefaultConfig with explicit profile/file selection and an
   injected scripted transport; call GetCallerIdentity with context/operation options
   and inspect output/Metadata. Read [default configuration](default-configuration.md)
   and demonstrate explicit long-lived opt-in and supported-mode limits; no real login
   is needed for this offline task.
2. Call native AssumeRole, then compose the generated client with
   AssumeRoleProvider and Cache to consume temporary credentials, with no response
   translator or application refresh loop. Use fixtures, not a cloud role.
3. Substitute a small AssumeRoleAPI mock; classify an APIError with errors.As and a
   canceled operation with errors.Is.
4. Configure AnonymousProvider and call OIDC/SAML with explicit fixture fields;
   explain why signed operations cannot use this marker, nil is rejected, and STS
   supplies no paginator or waiter.
5. Run the pinned official-v2 comparison workload; record the concrete differences
   in context/options/envelope/provider/mock use. Do not infer timing or superiority.

- Record each task's PASS/FAIL, start/end/elapsed minutes, docs used, compilation errors, obstacles, assistance and proposed fixes in the paired [result template](sts-independent-result-template.md).
- Report source revision and versions.
- Any required FAIL/NOT RUN keeps #60 and release #61 open.
- Do not publish account identity, credentials, assertions or raw requests/responses in evidence.
