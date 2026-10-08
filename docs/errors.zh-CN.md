# 结构化错误

[English](errors.md)

- ErrIncompleteOperation 表示中间件返回成功却未完成输出；不会发布之前失败尝试的解码值。
- 每次尝试独立拥有解码状态，只有整个成功调用流程 完成后才更新输出。

- 所有运行时失败返回 `*alicloud.OperationError`，包含服务、操作和 `Metadata`（请求 ID、HTTP 状态、尝试次数）。
- `errors.Is` 保留取消和超时；`errors.As` 提取 `*alicloud.APIError`。
- 服务消息可能包含敏感输入，通过 `Message` 显式读取；默认错误文本省略该消息。
- 操作错误文本也省略可能含 URL 的底层传输文本。
- 默认不要记录原始 底层错误。
- `ErrResponseTooLarge` 可由 `errors.Is` 识别。
- 运行时错误同时返回元数据；成功的强类型输出将其与API 响应字段分开保存。
- 返回后错误不可修改。
