# OSS XML 表示差异验证

[English](README.md)

- 对应 #92。本模块用合成 XML 验证官方来源中的表示差异。
- 在仓库根目录执行 `go -C tools/ossxml test ./...` 和 `go -C tools/ossxml vet ./...`。
- 测试不发起网络请求；首次下载依赖需要联网。不使用账号、Profile 或凭据。
- 依赖仅属于这个独立对比模块，SDK 运行时不依赖 Tea、官方 SDK 或第三方 XML 库。

## 固定输入

- [完整 OSS DSL](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/oss-20190517/main.tea)仍是计划中的生产模型来源。
- [OSS Teafile](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/oss-20190517/Teafile)声明 Go 产品版本 v2.0.1；本对比使用该版本的[生成模型](https://github.com/alibabacloud-go/oss-20190517/blob/935681aa139c270f2b6743e57a3c0dab548f39b4/client/client.go)。
- 调研解析到的 GatewayOSS 0.0.42 导入 GatewayOSS_Util 0.0.8，后者声明 Go helper v0.0.6。本对比使用该 helper 的[XML 根节点定义](https://github.com/alibabacloud-go/alibabacloud-gateway-oss-util/blob/cd82cbd16bcb3f0988e125ee63679e887867f225/client/structs.go)和[解析入口](https://github.com/alibabacloud-go/alibabacloud-gateway-oss-util/blob/cd82cbd16bcb3f0988e125ee63679e887867f225/client/client.go)。
- helper 使用 [tea-xml v1.1.3](https://github.com/alibabacloud-go/tea-xml/blob/v1.1.3/service/service.go)。具体模块版本和校验数据库哈希记录在 go.mod/go.sum。
- 本测试明确对比产品声明的模型版本与调研中 gateway 声明的 helper 版本，不代表未经修改的 OSS v2.0.1 安装默认使用 helper v0.0.6，也不代表所有官方版本行为相同。

## 已验证的差异

- ACL 和 CORS：helper 保留结构化根节点包装。直接转换到声明的响应模型时，body 存在，但内部字段丢失；在样本中明确去除这层包装后，字段能够保留。
- 地域：标量根节点本身就对应 DSL 字段。保留包装能够读取值；直接使用 encoding/xml 解码到 body 模型，会丢失根节点文本。
- 非法 XML：固定版本的 helper 吞掉解析错误并返回空 map，标准 XML 解码器则返回错误。
- 外部 Example 输出确定，无需账号；CI 在 Windows 和 Linux race 环境运行这个独立模块。

## 生成器决策

- 先规范化 helper 与模型的表示，再判断来源是否冲突。
- 根节点及标量、结构化差异必须来自固定的原生声明和完整语义模型，不能按操作名猜测，也不能建立永久逐操作手写表。
- 这里包含明确的测试样本，尚未实现生产规范化器。完整来源、导入与许可固定，来源漂移时的写前拒绝，XML 协议输出、OSS 签名和流式处理仍待完成。
- 生产 XML 解析必须保留字段，遇到非法正文时返回错误。与上游 helper 有意不同的行为，须记录来源和行为证据。
- 后续按[OSS 路线](../../docs/oss-xml-protocol.zh-CN.md)推进。
