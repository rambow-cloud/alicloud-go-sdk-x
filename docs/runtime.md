# HTTP runtime

[中文](runtime.zh-CN.md)

- Credential configuration follows [STS-first provider contracts](credentials.md).
- Config and service Options never accept bare AK/SK/token fields.
- Explicit static, environment and custom providers remain valid; nil/typed-nil providers fail before requests, without implicit fallback or retrieval during construction.

- Generated services expose a concrete Options type with the alicloud.Config fields, NewFromConfig(config, optFns...) and the existing New(config, optFns...) convenience path.
- Both validate and return (\*Client, error).
- Client.Options returns a snapshot with copied middleware registrations.
- Operation options apply to an isolated copy of configuration, including Retryer/HTTPClient/EndpointResolver, before execution; nil functions and invalid limits fail.
- Explicit NoRetry disables a retry override.
- Hooks, providers and transports remain shared objects with concurrency contracts.
- Root InvokeModel provides the typed Serialize boundary; CallOptions.Config selects a full validated per-call configuration snapshot for advanced wire callers.

- Construct `alicloud.Client` with an explicit credential provider.
- Defaults are a 30-second total deadline, eight MiB per response, no retry and reviewed HTTPS endpoints.
- Standard `http.Client` values are copied and redirects disabled; custom Do implementations must honor context and never follow signed redirects.
- Core packages depend only on the standard library. `Invoke` copies query/header/body data, signs each attempt with fresh credentials and nonce, closes response bodies and decodes JSON v2.
- Unknown fields are tolerated; duplicate names and invalid UTF-8 fail.
- Successful decoding assigns output atomically; callers must not share output pointers or mutate inputs during a call.
- All invocation failures wrap OperationError and preserve causes and metadata.
- Request bodies, including middleware changes, are limited to eight MiB and buffered for replay.
- Middleware runs Initialize/Serialize/Build once and Finalize/Deserialize per attempt.
- Modify signed fields before calling next in Finalize.
- Region overrides also replace an existing RegionId query parameter.
- Hooks, HTTP clients, resolvers, providers, sleep functions and retry policies remain shared and must be concurrency safe.
- Runtime supports ACS3 JSON OpenAPI and explicitly reviewed anonymous RPC. OSS/SLS, streaming and complete product coverage remain outside this scope.
