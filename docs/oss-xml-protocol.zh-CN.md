# OSS XML 与流式处理后续路线

[English](oss-xml-protocol.md)

- 对应 #92；当前仅调研，尚未修改生产来源、IR 或运行时。
- 官方 OSS 产品为 ec489e5c3deae95496daae2b41503ac58b221adb 中的 oss-20190517，原始 main.tea SHA256 为 0713683b286a49e4e80fff023f61db426dc035fb8755a7b4616a82f65c22e6ff。90 个 WithOptions 函数调用 execute，初始化 SPI/GatewayOSS，并通过 bucket hostMap 传入主机参数。
- 来源响应模式：XML 79、JSON 3、none 4、string 1、binary 3；请求模式：XML 84、binary 3、JSON 2、multiFormData 1。这是源码数量，不代表解析、降低或生成验收。
- 官方 registry 调研解析到 GatewayOSS 0.0.42；归档 SHA1 为 0b31bfdd4c86a28a6c93d74c0622e3167558435f，SHA256 为 5bf16e8d5283b8174c34d17fd4cb085422c7ef64a8d129beb93dd404216b7cab。该调研结果不是生产锁定；生产生成前还须固定 OSSUtil、GatewayOSS_Util、Time 导入及许可和来源证据。
- GatewayOSS 包含专用签名、bucket 主机和子资源行为、XML 正文与错误解析，以及流和校验和规则。不能只增加 encoding/xml 就将 OSS 交给 ACS3 处理。
- 开发前先提交配对路线和验收标准，固定完整原始产品、辅助模块及导入，将准确的辅助函数、根节点、主机和签名证据提取到 IR，再实现公共协议组件和通用 Go 输出。不支持的模式或显式选中的未知操作仍须在写入前失败。
- 明确流的转移与关闭、context 生命周期、请求哈希与重放、响应校验和规则。独立签名、XML、流所有权测试及虚构值 Example 随实现交付，并执行前端、Go 和 CI 门禁；不创建 OSS 资源或执行真实写操作。
- FC InvokeFunction 还需要独立的 readable 正文、请求头转换和原生错误验收；不能将 JSON/none 的有界缓冲读取记作二进制流完成。

## 首项验证

- 在 tools/ossxml 建立独立模块，固定官方 helper 和模型版本。只使用合成 ACL、CORS、地域 XML，不调用 HTTP，也不读取凭据。
- 分别验证结构化根节点和标量根节点。对比原始 helper 转换与明确的样本规范化，不引入逐操作的生产 overlay。
- 记录非法 XML 的处理方式。新运行时必须返回解析错误，不能沿用 helper 静默返回空结果的行为。
- 这项对比仅提供来源和表示差异证据。生产来源固定、语义投影、签名、XML 生成及流处理仍待完成。

## 实际结果（2026-10-09）

- [固定版本对比](../tools/ossxml/README.zh-CN.md)：在 Windows/amd64、Go 1.27.1 环境下，三个测试、两个结构化根节点子测试和一个外部 Example 均通过；独立模块的 vet 也已通过。
- 原始 helper/model 转换丢失 ACL/CORS 内部字段，样本中明确去除包装后能够保留。地域响应须保留标量根节点对应的字段。helper 吞掉非法 XML 错误，标准解码器会报告错误。
- 这是文档中明确版本组合的离线表示差异结果，不是 OSS 客户端实现，不代表所有官方版本都有同样缺陷，也不是云上验收。CI 须等对应 PR 的检查通过后才能记为通过。
