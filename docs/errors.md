# Structured errors / 结构化错误

## English

ErrIncompleteOperation identifies middleware that reports success without completing
an output. Each attempt owns its decode state; earlier failed attempts cannot publish
data. Output is assigned only after the whole successful pipeline completes.

All runtime failures return `*alicloud.OperationError` with service, action and `Metadata` (request ID, HTTP status, attempts). `errors.Is` preserves cancellation and deadlines; `errors.As` retrieves `*alicloud.APIError`. The service message may contain sensitive input and is accessible explicitly through `Message`; default error formatting omits it. Operation error formatting also omits underlying transport text, which may contain URLs. Do not log the raw cause by default. `ErrResponseTooLarge` is inspectable with `errors.Is`. Metadata is also returned alongside runtime errors; typed successful outputs carry it separately from wire fields. Errors are immutable after return.

## 中文

ErrIncompleteOperation 表示 middleware 返回成功却未完成输出；不会发布之前失败尝试的解码值。
每次尝试独立拥有解码状态，只有整个成功 pipeline 完成后才更新输出。

所有运行时失败返回 `*alicloud.OperationError`，包含服务、操作和 `Metadata`（请求 ID、HTTP 状态、尝试次数）。`errors.Is` 保留取消和超时；`errors.As` 提取 `*alicloud.APIError`。服务消息可能包含敏感输入，通过 `Message` 显式读取；默认错误文本省略该消息。操作错误文本也省略可能含 URL 的底层传输文本。默认不要记录原始 cause。`ErrResponseTooLarge` 可由 `errors.Is` 识别。运行时错误同时返回元数据；成功的类型化输出将其与线协议字段分开保存。返回后错误不可修改。
