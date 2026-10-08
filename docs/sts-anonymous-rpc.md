# Anonymous STS RPC

[中文](sts-anonymous-rpc.zh-CN.md)

- Issue #59 supports AssumeRoleWithOIDC/SAML from pinned official DSL.
- These functions use authType=Anonymous and the eight-argument doRPCRequest handoff; this differs from the signed callApi path.
- The pinned imported OpenApi 0.3.23 main.tea:176-290 supplies Action, Version, Format=json, UTC Timestamp and SignatureNonce in the query, plus action/version headers.
- With no request.body, the body is empty.
- The Anonymous branch does not obtain credentials or add a signature.
- Preserve this reviewed wire contract; do not reuse ACS3 signing or silently accept dynamic/unknown handoffs.

- The explicit credentials.AnonymousProvider marker lets an application construct a client without inventing source keys; nil/typed-nil providers remain invalid.
- It never supplies signing credentials.
- Generated operations choose the reviewed AuthenticationAnonymousRPC mode; signed operations remain signed and reject an anonymous marker.
- A supplied custom/static/cache provider is not retrieved by an anonymous operation.
- Configuration is not mutated.
- This follows the explicit-marker idea in [AWS AnonymousCredentials](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws#AnonymousCredentials) while retaining our per-operation protocol and strict missing-provider conventions.

- OIDCToken/SAMLAssertion are explicit native request fields.
- Runtime preparation strips source authentication headers/query parameters, preserves exact service field values, and rejects unsupported method/path/body before transport.
- Policies redact complete sensitive request/response models in String/GoString; explicit JSON remains raw.
- Error/tracing defaults never record raw URLs/bodies/tokens.
- Retry remains opt-in and these token-issuing operations remain non-idempotent under the reviewed policy.

- Acceptance uses independent offline common/service query, encoded token/assertion, empty-body/header/provider-isolation, complete response/presence, JSON-v2/error/context, middleware/tracing/formatting and signed-operation regression contracts.
- Generate English comments, paired guides/source indexes and deterministic external Examples.
- Successful live OIDC/SAML federation is NOT RUN, outside v0.1.0 required live scope; IdP/provider/assertion provisioning and federation credential helpers are future work.
