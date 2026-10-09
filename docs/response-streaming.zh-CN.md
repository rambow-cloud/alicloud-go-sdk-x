# 有界响应流

[English](response-streaming.md)

- 对应 #92。先实现共享运行时契约，再生成二进制 DSL 操作。
- `ResponseBodyStream` 成功后返回由调用者持有的 `io.ReadCloser`，不预先读取响应。底层调用使用 `StreamingOutput`；模型编解码器将运行时提供的流赋给输出中导出的顶层 `io.ReadCloser` 字段。
- 调用者负责关闭流。读取结束、读取失败、超限、取消或超时也会关闭底层响应，且只关闭一次。
- `Config.Timeout` 覆盖请求及返回后的读取；`MaxResponseBytes` 限制可交付字节数。默认值仍为 30 秒和 8 MiB；调用者更短的 context 截止时间优先。
- 同一个流只允许一个读取者；Close 可以与 Read 并发。`errors.Is`/`errors.As` 保留取消、超时和传输错误，默认错误文本不输出底层错误内容。
- `Codec.DecodeStream` 只挂接收到的流，并复制需要保留的响应头；不得读取、关闭或替换流，也不得保留响应对象。
- 成功的中间件必须让最终输出保留原流。缺失或替换流时，在发布结果前报错；失败尝试中的临时流由运行时关闭。
- 发布前可按既有审核策略重试；发布后的读取错误不能触发新的 HTTP 请求。
- 请求仍使用独立复制、可重放且不超过 8 MiB 的字节切片。本阶段不增加无限制的请求流。
- 使用无需账号的测试验证生命周期、准确字节上限、取消、错误脱敏、重试以及编解码器和中间件的流归属；同步提供确定性外部 Example、公共 Go 文档和常规 Go 检查。
- FC 二进制操作和 OSS XML、签名、校验和仍需各自完成解析器、IR 和输出器验收。基础运行时通过不代表已支持生成的二进制操作或真实云调用。

## 使用方式

- 为已审核操作设置 `ResponseBody: alicloud.ResponseBodyStream`。
- `Invoke` 接收 `*alicloud.StreamingOutput`。先检查调用错误，成功后再读取或关闭 `Body`。
- 流只读取一次；读取错误和调用错误分别处理。成功后使用 `defer output.Body.Close()` 确保关闭。
- `Headers` 是独立复制的 map。`StatusCode` 和返回的 `Metadata` 描述初始成功响应，不代表正文已经读完。
- `InvokeModel` 需提供 `Encode` 和 `DecodeStream`；流模式不要求 `Decode`。输出中的流字段可以使用任意名称。
- 复用输出变量不会关闭调用者之前持有的流；必须先关闭旧流。
- 使用 `go test . -run ExampleStreamingOutput` 运行 [ExampleStreamingOutput](../stream_test.go)，其传输为本地注入，不需要云账号。

## 验证结果

- 本地 Windows、Go 1.27.1：全量包测试、17 个公共包的文档检查、vet、格式和语言文件及本地链接检查均通过。
- 测试覆盖：不预读二进制正文、准确上限和超限、EOF、未读取时取消、阻塞读取、超时、并发关闭、读取及关闭错误脱敏、结构化服务错误、发布前重试，以及编解码器和中间件违反流归属规则时的处理。
- 完整生成及隔离编译测试耗时 112.468 秒并通过。随后补充关闭错误测试、增加截止时间测试余量，受影响测试也通过。
- 最终提交的 Linux race 和 Windows CI 尚未完成；没有执行真实二进制云 API 调用。
