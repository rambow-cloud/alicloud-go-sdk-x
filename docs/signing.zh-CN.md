# ACS3 签名

[English](signing.md)

- 内部签名器对实际传输的方法、转义路径、规范查询、选定 header 和 SHA256 请求体摘要实现 ACS3-HMAC-SHA256。
- RPC 使用 `/`；ROA 路径保留分段转义，包括编码的斜杠。
- 查询空格使用 `%20` 而非 `+`；重复值确定排序。
- Host、content-type 和所有 x-acs header 参与签名。
- STS token 被签名，轮换时移除旧 token。
- 每次运行时尝试使用新的密码学随机 nonce、时间戳和凭据快照。
- 固定向量测试匹配[官方签名样例](https://help.aliyun.com/zh/sdk/product-overview/v3-request-structure-and-signature)，无需账号。
- OSS/SLS 等自定义签名协议不在范围内。
- 签名器保持内部；中间件修改签名数据应在 Finalize 调用 next 前进行。
