# OSS 校验和基础层

[English](oss-checksums.md)

- 对应 #92，接续内部 XML 编解码层与 OSS4 签名器。后续按照官方 DSL/Gateway 接入操作时，明确选择哪些操作需要完整性校验及其规则。
- 使用标准库实现可取消的 CRC64 和 Content-MD5 字节计算。保留已完成 CRC 作为初始值的追加/续算语义；取消时不返回部分校验和。不保留或修改输入，不增加原生 SDK 依赖。
- CRC 算法以官方 Go SDK 固定提交 `2614e2cfdc3618f204cee8bb247f9359e8e509aa` 的完整源码和测试为依据。原生 `TestCRC64` 使用 Go `crc64.ECMA`，向量为 `123456789` → `995dc9bbdf1939fa`，`This is a test of the emergency broadcast system.` → `27db187fc15bbc72`。复用事实向量，不复制原生实现。
- 服务端 CRC 按无符号十进制 uint64 验证。缺失、非十进制或溢出均明确失败；不一致返回脱敏的哨兵错误。不把 ETag 当校验和，也不把缺失头当成验证成功。
- 可接受能够解析为 uint64 的前导零；拒绝空白、正负号和十六进制形式。错误不包含原始服务端值或正文。空正文保留 CRC 初始值，Content-MD5 为标准空正文摘要。
- 测试独立的官方向量、追加/分块续算、空输入、MD5/base64 已知向量、输入所有权、畸形/溢出/不一致脱敏，以及计算前和计算中的取消；提供无需账号的外部 Example、配对文档和 Go、文档、vet、格式检查，合并前要求最终提交 CI 通过。
- 本阶段不自动启用上传/下载校验，不验证部分或范围下载，不合并独立计算的分片 CRC，不读取无界流，也不修改重试和重放规则。流 EOF、提前关闭、取消、所有权及操作专用来源策略另行完成。
- 来源：[官方 CRC64 协议](https://www.alibabacloud.com/help/en/oss/user-guide/check-data-transmission-integrity-by-using-crc-64)、[固定原生测试](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/2614e2cfdc3618f204cee8bb247f9359e8e509aa/oss/utils_crc_test.go)、完整固定 GatewayOSS 0.0.42（`main.tea` SHA256 为 `71caf417a396688b8a4a38b0e0f431d5adfb7ae043d98fa504f07ac1b319f44c`）。校验算法与 CRC 头缺失时的操作策略分别跟踪。

## 验证

- 实现和检查尚未完成。
