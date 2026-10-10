# 内部 OSS V4 签名

[English](oss4-signing.md)

## 临时凭据修正（#126）

- 原生 OAuth 核验发现临时 AccessKey ID 使用 `STS.` 前缀。原实现复用了地域校验规则，因此在发送请求前拒绝了句点。
- 单独校验凭据标识符：允许 ASCII 字母、数字、连字符、下划线和句点；拒绝空白、控制字符、非 ASCII 文本及签名范围或请求头分隔符。地域校验规则保持原样。
- 使用带句点的合成标识符验证既有独立 STS 签名向量，增加运行时临时 provider 集成和非法标识符拒绝测试。已有离线 OSS Example 同步使用带句点的合成标识符。
- 已授权的真实核验只使用现有原生临时 Profile 和 GET 请求，不公开凭据、资源标识符或正文。SDK 线路证据不代表已经生成并验收 `service/oss`。
- 实现后运行文档检查、vet、完整 Go 测试和格式检查；合并前要求最终提交 CI 通过。

- 2026-10-10 本地核验：根包运行时测试、独立 STS 签名向量、非法输入拒绝测试、确定性 Example、文档检查（17 个包）、vet 和格式检查通过。首次完整 Go 测试只有两个 codegen 测试失败，原因是新 worktree 未安装固定版本的 Node parser。执行 npm ci 安装锁文件依赖后，这两项重跑通过（37.426 秒）；其他包已在首次测试中通过。
- 修复后的真实只读核验：通过 oss-sftp 原生临时 Profile 读取 ListBuckets、GetBucketInfo 和 GetBucketReplicationLocation，均只尝试一次，返回 HTTP 200、application/xml 和 OSS 请求 ID。修复前三项均在发送 HTTP 前报 OSS V4 输入无效；修复后的原始响应结构与官方 CLI 一致。公开 OSS 生成和浏览器界面核验仍分别验收。

- 对应 #92，在有界响应流和 FC 二进制生成之后建设内部签名能力。
- 依据[官方 V4 规范](https://www.alibabacloud.com/help/en/oss/developer-reference/recommend-to-use-signature-version-4)和已审核的 GatewayOSS 0.0.42 来源，使用标准库独立实现，不复制原生 SDK 或增加运行时依赖。
- 显式指定 bucket、region 和时间。规范资源路径包含 bucket，并保留对象键中的斜线；query 按编码后的名称排序，空值使用无等号的子资源名，不复用 ACS3 的请求结构。
- 签入原生 content-type/content-md5 和 x-oss 请求头，包括统一设置的日期、token、UNSIGNED-PAYLOAD。额外签名头须显式登记。请求正文保持原样，不读取或关闭。
- 规范请求头名称，拒绝有歧义的重复请求头/query 键、错误路径/query 字节、注入、不完整 scope 和混入的 ACS3 字段。先在独占副本上验证，失败不修改请求。
- 验收包括官方签名密钥、规范请求摘要和签名向量，独立合成的路径/query/请求头/STS/正文所有权用例，取消与失败不修改请求，确定输出的外部 Example、配对文档和 Go/CI 门禁。
- 本阶段不增加公共操作、认证模式、OSS 端点构造或生成器。XML 特征、IR/Gateway 接入、请求重放和校验和、真实云验收仍未完成；不调用云 API 或创建资源。

## 验证

- Node 22 前端/IR/说明检查通过。Go 1.27.1/Windows 下，文档检查、完整 vet 和根模块测试通过（codegen 耗时 118.271 秒）。
- 三组参数完整的原生向量，以及文档中的规范请求摘要/给定密钥 MAC 通过。正文所有权、STS 替换、中文与斜线、空值子资源、额外请求头及安全拒绝均有离线夹具。
- 最后评审核对了原生路径规范化，并收紧逐跳请求头校验，相应签名测试覆盖这些修正。Go 创建的中文 URL 路径提示会规范化；解码后不匹配的原始路径报错。最终提交的 Linux race/Windows CI 尚未完成，没有增加运行时依赖或调用云 API。

## 来源决策

- [固定向量记录](../metadata/oss4-signature-evidence.json)注明 GatewayOSS 压缩包/源码哈希，以及原生 SDK 的提交和文件。仅复用合成向量参数与预期签名。
- 文档中的规范请求摘要，以及使用其给定派生密钥计算的 MAC 均吻合；但从页面显示的 `yourAccessKeySecret` 派生出的密钥不同。该显示值不能证明端到端向量通过；保留差异，使用原生 SDK 中参数完整的基础、STS 和额外请求头向量验证密钥派生。
- 规范化签名时，请求头名称统一小写。Authorization 的逗号空格采用 Gateway 格式；签名值与原生测试单独对齐，不依赖格式空格。
- 拒绝重复的标量请求头/query 键、预签名认证 query 和尚未展开的 `x-oss-meta-*`；这些转换由 Gateway/正文准备阶段在签名前处理。额外请求头必须实际存在、唯一且不属于必选集合，不自动推断登记。
- `UNSIGNED-PAYLOAD` 认证的是审核过的 OSS V4 请求结构，不提供正文完整性验证。XML MD5、流式 CRC 和重放另行实现。

## 内部契约

- `SignOSS4(ctx, request, credentials, options)` 接收调用者独占的 HTTPS 虚拟主机/服务级请求。准确指定 bucket、region 和非零时间，不自动发现；URL 路径不含虚拟主机 bucket。
- 成功时发布 URL 和请求头副本，其他请求字段、Body 和 GetBody 不变。失败或取消不修改请求；过期错误仍可用 `errors.Is` 识别。
- 审核过的协议范围只接受标量请求头/query。原生 content-type/content-md5/x-oss 请求头自动签入；额外名称须显式登记，逐跳头不支持。Host 使用实际的 Request.Host/URL.Host；签入 content-length 需正文长度已知且为正，并且没有 transfer encoding。
- 不支持 V1/V2、预签名 URL、CloudBox、端点构造或校验和计算。内部 helper 没有共享可变状态，调用者签名期间须独占请求。
- 使用 `go test ./internal/signing` 运行合成向量和确定输出的外部 Example。测试不发送 HTTP 请求，也不输出授权头或凭据值。
