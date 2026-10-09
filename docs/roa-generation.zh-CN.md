# ROA 生成能力的实施阶段

[English](roa-generation.md)

- 对应 #92，本次是部分交付；保留现有 RPC 生成能力。
- 先实现并验证公共运行时中的准确路径编码。
- 再将官方 ROA 的参数、请求头和 JSON 正文降低为完整 IR，生成类型化调用。
- [显式响应模式](roa-response-modes.zh-CN.md)及 [FC JSON/none 生成](fc-roa-product.zh-CN.md) 在路径能力上扩展。XML、流的所有权及其他签名协议仍是 #92 的独立后续阶段。

## 路径约定

- `Request.Path` 表示解码后的路径，可选的 `Request.RawPath` 保留参数段内部的转义。
- RawPath 必须是绝对路径、转义有效，且解码结果与 Path 完全一致；无效组合在读取凭据或调用传输器前失败。
- 签名与 HTTP 传输使用同一份规范化编码路径。经过编码的斜杠仍属于参数段，不变成路径分隔符。
- 内部路径构造器接受命名占位符和准确匹配的参数集合，将每个值按 RFC3986 编码为单个路径段，不清理路径，不将参数解释为 URL。
- 不修改输入 map。空值、缺失参数、格式错误的模板、无效 UTF-8 或已取消的 context 都会失败。
- 离线示例和线路径测试覆盖空格、斜杠、百分号、中文，以及无效输入不触发传输的行为。

## 官方产品调研

- FC 来源：[固定版本的完整 DSL](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/fc-20230330/main.tea)。
- 它包含多个路径参数、独立请求头 map、请求模型的 JSON 正文、共享响应模型和 `none` 响应正文。
- 固定端点表的 `ap-southeast-5` 主机名带有尾随空白。保留上游原始字节，接受该条目前需记录绑定来源的规范化决策。
- FC 发现 73 个操作，离线生成 72 个；二进制 InvokeFunction 暂不支持。尚无真实 FC 验收。
