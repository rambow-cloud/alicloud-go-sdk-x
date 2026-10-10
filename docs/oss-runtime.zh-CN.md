# OSS 共享运行时

[English](oss-runtime.md)

- 后续来源固定和前端验收遵循 [OSS 语义 IR](oss-semantic-ir.zh-CN.md)。暂存操作完成降低，不代表已经生成公开的 service/oss。

- 对应 issue #92。先把已完成的 XML/OSS4 组件接入共享 HTTP 调用链，再推进产品代码生成。
- 协议依据：GatewayOSS 0.0.42 main.tea，SHA256 为 `71caf417a396688b8a4a38b0e0f431d5adfb7ae043d98fa504f07ac1b319f44c`。XML 请求和 MD5 见第 92–99 行，路径与查询拆分见第 125–146 行，XML 错误见第 153–181 行，桶域名见第 305–334 行。[签名](oss4-signing.zh-CN.md)和 [XML 编解码](xml-model-codec.zh-CN.md)沿用各自证据。

## 实现范围

- 显式 OSS4 鉴权和缓冲 XML 响应模式。XML 成功响应必须使用具体类型的操作 codec，不回退到 JSON，也不猜根节点。
- 桶名与解码后的对象路径分别传入。解析地域级服务地址后，校验桶名并只添加一次桶前缀。空桶名表示服务级请求。
- 必须明确地域和使用 DNS 主机名的 HTTPS 地址。拒绝 IP 地址、已包含当前桶前缀的地址及中间件造成的签名主机不一致。暂不支持 CNAME、路径式寻址、接入点、CloudBox 和依赖用户 ID 的网关寻址。
- 生成器需将 DSL 子资源放入 Request.Query。OSS Path 中出现 `?` 或 `#` 时必须提供相应的 RawPath 编码，不把对象名字节解释成查询语法。
- 对已审核的非空 XML 请求字节设置 application/xml；每次尝试在 Finalize 中间件之后重新计算 Content-MD5。请求保持独占副本，先缓冲后发送，上限八 MiB。
- 复用 provider、签名、中间件、显式重试、超时、大小限制及响应流所有权。暂不支持无界流式上传和自动响应完整性校验。
- 保留调用者的 OSS Accept 请求头，不为 JSON/二进制操作强制设置 XML，也不根据响应解码模式猜测内容协商规则。
- 在大小限制内解析 Error 根节点 XML 错误，提取 RequestId/EC，并支持 OSS 响应头回退。空错误体以 HTTP 状态码字符串作为 code；XML 无效时使用 InvalidErrorResponse。默认错误字符串不包含 message、EC 或原始响应体。
- XML 命名空间仍精确匹配。原生按节点名匹配的行为及 ListBuckets 大小写、包装层差异，需要后续生成器策略明确处理。

## 编码前确定的验收

- 独立离线 HTTP 测试覆盖桶和服务地址、转义对象路径及子资源、STS 请求头、无 ACS 请求头、中间件修改后的 XML MD5、类型化输出、无效 XML/错误、响应大小限制及所有权。
- 重试测试证明不会发布旧输出，每次尝试重新获取凭据、重新计算摘要；任意写操作仍不具备默认重试资格。
- 无效模式、桶名、地域、地址和签名主机在发送前失败；初始桶名、地域、地址无效时不获取凭据。
- errors.Is 能识别取消；响应流不提前读取，底层关闭一次。保留 ACS3 和匿名调用测试。
- 提供输出确定的外部 Example、英文公共 API 注释，并同步更新两个语言版本。
- 实现后各运行一次 Node 22 前端检查和测试、sdkgen check/product-check、格式检查、doccheck、vet、Go 测试；评审最终差异和最终提交的 CI。
- 本阶段不注册生产 OSS 来源、不改 IR、不生成 service/oss、不批准 ListBuckets 转换，也不访问真实 OSS。#92 保持打开。

## 下一阶段

- 固定完整 OSS 产品源码及导入闭包，将可支持的 Gateway 初始化、hostMap 和准确 XML 信息投影到 IR。
- 通过通用后端生成 service/oss，明确不支持的原因，并用更名合成产品证明编译流程可以复用。ListBuckets 仅在审核协议证据后解决。

## 使用方式与默认行为

- 将 Operation.Authentication 设为 AuthenticationOSS4；填写 Config.Region，并提供不含桶前缀的 BaseEndpoint 或自定义端点解析器。目前尚未注册 OSS 默认端点目录。
- Request.Bucket 选择桶域名寻址；对象名字节放在 Path/RawPath，子资源参数放在 Query。
- RequestBodyXML 接收已编码的字节，根节点和类型校验由操作 codec 负责。ResponseBodyXML 必须使用 InvokeModel 与 Codec.Decode；内部公共 XML codec 精确校验根节点名称及命名空间。
- 参见无需云账号的[运行时 Example](../oss_example_test.go)。其中的小型显式 codec 演示扩展入口；生产生成将复用已有的严格 XML codec。
- 默认行为保持不变：不自动重试、总超时三十秒、请求和响应上限八 MiB。即使自定义响应上限更大，内部 XML codec 的文档上限仍是八 MiB。
- 本阶段未实现 CNAME/路径式地址、预签名、V1/V2、无界上传 reader、响应 CRC/MD5 策略和真实云验收。

## 本地验证（2026-10-10）

- Node 22.21.1：前端检查通过，110 项前端测试通过。生产发现仍包含四个产品，尚未加入 OSS。
- Go 1.27.1：格式、sdkgen check/product-check、doccheck（17 个公共包）、vet、语言及链接检查（263 个项目文档文件）通过。
- 首次完整 Go 测试中，除 internal/codegen 外均通过。该包的隔离编译样本漏复制新接入的 checksum/XML 依赖。已补齐样本和 STS 来源更新演练的依赖；四项受影响的前端演练测试及 internal/codegen 全套测试通过，后者耗时 125.278 秒。
- 评审补充了最终桶域名的 DNS 长度校验及大小写混用请求头的测试；受影响的根包 vet/测试通过，测试耗时 3.189 秒。未访问云端，也未重新生成来源或 IR。
- 最终评审移除了对 OSS Accept 的推断；JSON/none 调用会保留显式请求头，也不会擅自补充缺失的 Accept。修正后的根包 vet/测试通过，测试耗时 3.040 秒。
- 最终提交的 CI 和合并证据记录在对应 PR/issue 中，本地结果不代替这些证据。
