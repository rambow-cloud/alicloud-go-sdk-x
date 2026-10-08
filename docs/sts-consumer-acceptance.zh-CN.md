# STS 消费者验收

[English](sts-consumer-acceptance.md)

- #60 验收四个生成操作与消费者维护，与 #61 发布分开。
- 隔离的 `examples/stsacceptance` 模块固定官方 STS v2.1.0；本 SDK 通过显式本地 replace 使用当前检出源码。
- 报告记录 `git rev-parse HEAD`、`go version`、系统和模块版本；对比依赖不进入 SDK 运行时模块。

- 实现者离线运行只是技术证据；用户安排的独立 Go 开发者仍需仅按文档完成任务。
- 真实成功联邦调用预先范围外且记 NOT RUN；离线包不要传入真实凭据或创建云资源。

### 独立开发者交接

- 检出交接时提供的准确提交，使用 Go 1.27+，阅读上述 STS 组合、凭据、生成产品和匿名协议指南，在 `examples/stsacceptance` 执行 `go test -v ./...` 与 `go run .`。
- 固定占位值与脚本 HTTP 测试数据无需账号，第一次构建可能下载依赖，操作运行不访问网络。

- 在独立临时 Go 模块内、不接受实现者协助，完成：

1. 临时文件写入虚构原生 StsToken/OAuth CLI JSON Profile，通过 config.LoadDefaultConfig
   显式选取文件/Profile 并注入脚本 HTTP 传输实现，以 context/options 调用 GetCallerIdentity，
   检查输出/Metadata；阅读[默认配置](default-configuration.zh-CN.md)，展示长期密钥显式启用
   与模式范围。离线任务无需真实登录。
2. 调用原生 AssumeRole，再用生成 client、AssumeRoleProvider 与 Cache 消费临时凭据，不写
   响应格式转换器 或应用刷新循环；使用测试数据，不使用云角色。
3. 替换小 AssumeRoleAPI 测试替身，用 errors.As 识别 APIError、errors.Is 识别取消。
4. 配置 AnonymousProvider，显式传入测试数据调用 OIDC/SAML，解释签名操作为何不能使用
   此标记、nil 被拒绝及 STS 为什么不提供分页或状态等待器。
5. 运行固定官方 v2 对比任务，记录 context/options/envelope/凭据提供者/测试替身的具体差异，
   不推断速度或优劣。

- 按[结果模板](sts-independent-result-template.zh-CN.md)记录各项 PASS/FAIL、起止/分钟、所用文档、 编译错误、障碍、协助、改进和源码/环境版本。
- 必需项 FAIL/NOT RUN 使 #60/#61 保持开放。
- 证据不得发布账号身份、凭据、身份断言 或原始请求/响应。
