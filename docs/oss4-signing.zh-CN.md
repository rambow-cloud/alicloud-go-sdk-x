# 内部 OSS V4 签名

[English](oss4-signing.md)

- 对应 #92，在有界响应流和 FC 二进制生成之后建设内部签名能力。
- 依据[官方 V4 规范](https://www.alibabacloud.com/help/en/oss/developer-reference/recommend-to-use-signature-version-4)和已审核的 GatewayOSS 0.0.42 来源，使用标准库独立实现，不复制原生 SDK 或增加运行时依赖。
- 显式指定 bucket、region 和时间。规范资源路径包含 bucket，并保留对象键中的斜线；query 按编码后的名称排序，空值使用无等号的子资源名，不复用 ACS3 的请求结构。
- 签入原生 content-type/content-md5 和 x-oss 请求头，包括统一设置的日期、token、UNSIGNED-PAYLOAD。额外签名头须显式登记。请求正文保持原样，不读取或关闭。
- 规范请求头名称，拒绝有歧义的重复请求头/query 键、错误路径/query 字节、注入、不完整 scope 和混入的 ACS3 字段。先在独占副本上验证，失败不修改请求。
- 验收包括官方签名密钥、规范请求摘要和签名向量，独立合成的路径/query/请求头/STS/正文所有权用例，取消与失败不修改请求，确定输出的外部 Example、配对文档和 Go/CI 门禁。
- 本阶段不增加公共操作、认证模式、OSS 端点构造或生成器。XML 特征、IR/Gateway 接入、请求重放和校验和、真实云验收仍未完成；不调用云 API 或创建资源。

## 验证

- 实现和最终提交 CI 尚未完成。
