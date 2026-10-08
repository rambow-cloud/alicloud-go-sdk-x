# Supplemental STS consumer review

[中文](sts-consumer-review.zh-CN.md)

- Issue [#60](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/60).
- Result: **PASS for the five supplemental technical cases**, subject to exact-head Linux race/Windows CI before integration.
- This is an implementation-agent review, not a Go developer's independent docs-only acceptance.
- The human result remains NOT RUN; #60/#61 and publication/indexing gates remain open.

- Reviewed main baseline: `3e6a5a720e9ba6afd5c995102b7b7b56a7f955b3`.
- Environment: Go 1.27.1, Windows/amd64, Node 22.21.1.
- The isolated consumer module uses a local SDK replace and pins official STS v2.1.0, OpenApi v2.1.13, Tea v1.3.13.
- This change adds tests and documentation only; SDK runtime, generated outputs, source pins and policies stay at the reviewed baseline.

### Method and findings

- The separate `main_test` package in [consumer_review_test.go](../examples/stsacceptance/consumer_review_test.go) imports only public SDK packages, builds its own transport/mocks and does not reuse the original consumer fixture or SDK internals.
- Its transports never open a connection.
- Credential/identity/token/ARN values are fictional.
- Consumer construction follows the [STS composition](sts-credentials.md), [credentials](credentials.md), [cache contracts](credential-cache.md), [product guide](products/sts.md) and [anonymous protocol](sts-anonymous-rpc.md).
- Public declarations and signing evidence were inspected during review, so this is not labeled an unaided docs-only exercise.

| Case                          | Observation                                                                                                                                                                                                                                                                                            | Result |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ |
| Identity/options/cancellation | Native fields and RequestID/status/attempt metadata; per-call endpoint override leaves the next call unchanged; transport cancellation retains errors.Is and OperationError                                                                                                                            | PASS   |
| Native role/provider/cache    | Caller mutates its session after provider construction; original session and explicit 900 seconds survive. Canceling the initiating waiter leaves the shared issuance alive; 24 concurrent consumer reads use the issued role token with exactly one issuance and no application refresh loop          | PASS   |
| Small mock/errors             | A single AssumeRoleAPI mock suffices; errors.As and errors.Is preserve the exact APIError; already-canceled retrieval never reaches the mock                                                                                                                                                           | PASS   |
| Anonymous federation          | Both OIDC/SAML preserve `+ / = &` and spaces, common query fields and empty POST body; four calls across AnonymousProvider and an unusable custom provider perform zero source retrievals. Signed identity rejects the marker; nil/typed-nil construction fails; default model formatting hides tokens | PASS   |
| Unsafe retry/error formatting | AssumeRole with Standard and HTTP 503 executes once; APIError and OperationError retain code/status/request ID/attempts; the explicit service Message is available while default Error strings omit it                                                                                                 | PASS   |

- Initial execution at 2026-10-08 13:15:03 UTC passed four cases; the cache case failed because the reviewer incorrectly expected the AWS-style slash after `Credential=<key>`.
- Alibaba ACS3 uses a comma before SignedHeaders.
- Correcting that fixture assertion and adding early-error reporting made that case pass on its targeted rerun.
- No SDK fix was needed.
- This is evidence for the documented boundary: AWS-like Go calling conventions with Alibaba-native wire semantics.

- The pinned official-v2 four-action fixture also passed in this session.
- Its methods use RuntimeOptions and response Body envelopes, while this SDK exposes context-first operations, service functional options, direct native outputs plus Metadata and small operation interfaces.
- This review does not measure performance, evaluate all official versions or evaluate the official credentials library's refresh support.

- A documentation defect was found and corrected: the previous acceptance report grouped AC-06 through AC-12 under mismatched capability descriptions.
- Its references now match [the authoritative acceptance table](product-acceptance.md): retry, endpoint, middleware/OTel, errors, testing, documentation and generation respectively.

### Verification and limits

- Commands from the repository root after configuring local Go caches:

```text
go -C examples/stsacceptance test -count=1 -v -run '^TestConsumerReview' ./...
go -C examples/stsacceptance test -count=1 -v -run '^TestConsumerReviewNativeRoleCacheAndCanceledWaiter$' ./...
go -C examples/stsacceptance test -count=1 -v -run '^TestPinnedOfficialWorkload$' ./...
```

- The first command's initial FAIL and the targeted correction above are both part of the record; combined local results cover all five cases.
- Final full-module Linux race/Windows runs are supplied by the linked PR's exact-head CI.
- Documentation, vet and root tests are recorded there too.
- Test elapsed time is not independent developer task time.
- No fresh cloud calls, IAM/IdP provisioning or source updates were performed; already accepted [live renewal](live-sts-renewal.md), [identity evidence](acceptance/sts-identity-live.json) and [source rehearsal](sts-source-rehearsal.md) remain distinct.

- No blocking runtime defect was found in this bounded review.
- Successful live OIDC/SAML remains NOT RUN outside the predeclared required live scope; STS has no native paginator/waiter in this workload.
- Native Profile/OAuth discovery, broader product acceptance, human task usability/timing, release publication and same-version pkg.go.dev browser indexing are not established by this report.
