# STS 验收任务

[English](README.md)

- Go 1.27+ 下从此目录执行 `go test -v ./...` 和 `go run .`。
- 独立模块固定官方 STS v2.1.0，通过 `replace ../..` 使用检出 SDK，记录根目录准确 Git 提交。
- 脚本 HTTP 使用虚构身份/密钥，操作不联网、不建资源；初次下载依赖可能需要网络。
- 对比依赖与 SDK 核心隔离，[任务和规则](../../docs/sts-consumer-acceptance.zh-CN.md)详述验收。

- 两个 SDK 均执行四个原生操作；生成 SDK 另组合角色凭据提供者/cache 完成两次签名消费者读取，不需应用 响应格式转换器 或刷新循环。
- 本任务未评估官方 credentials 库的角色凭据提供者/刷新能力、性能或独立 UX，不能从测试数据成功推断。
- 此固定官方 v2.1.0 方法不接 context、输出带 Body envelope，这不是对全部官方版本的断言。

- `consumer_review_test.go` 为独立外部包，只导入公开 SDK API，使用新测试数据。
- 通过上述 `go test -count=1 -v -run '^TestConsumerReview' ./...` 单独执行补充验收， 检查 options/输入隔离、调用中取消、24 个角色签名并发读取复用一次共享签发、小测试替身、 匿名凭据提供者/token 隔离与签发错误不重试。
- 见[补充报告](../../docs/sts-consumer-review.zh-CN.md)； 这是代理编写的补充评审，与独立开发者任务结果分开记录。

- 新增 TestExternalDefaultProfileWorkload 覆盖 LoadDefaultConfig→原生临时 CLI Profile→ 生成 STS 消费者。
- 独立身份任务现包含[默认配置](../../docs/default-configuration.zh-CN.md)与临时虚构 Profile，本离线验收包不要求交互登录。

- #83 新增 `TestOfficialRPCShrinkJSONHelperParity`，将生成的 ECS JSON 参数与已有固定版本 OpenAPI v2.1.13 helper 对比，覆盖 nil 与空值、嵌套、非 ASCII 文本及 int64 精度。此为无需账号的 helper 对比，不代表所有 ECS 操作兼容；未新增依赖。
