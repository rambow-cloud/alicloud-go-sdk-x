# Structured errors

[中文](errors.zh-CN.md)

- ErrIncompleteOperation identifies middleware that reports success without completing an output.
- Each attempt owns its decode state; earlier failed attempts cannot publish data.
- Output is assigned only after the whole successful pipeline completes.

- All runtime failures return `*alicloud.OperationError` with service, action and `Metadata` (request ID, HTTP status, attempts). `errors.Is` preserves cancellation and deadlines; `errors.As` retrieves `*alicloud.APIError`.
- The service message may contain sensitive input and is accessible explicitly through `Message`; default error formatting omits it.
- Operation error formatting also omits underlying transport text, which may contain URLs.
- Do not log the raw cause by default. `ErrResponseTooLarge` is inspectable with `errors.Is`.
- Metadata is also returned alongside runtime errors; typed successful outputs carry it separately from wire fields.
- Errors are immutable after return.
