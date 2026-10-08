# STS acceptance workload / STS 验收任务

## English

From this directory, with Go 1.27+: `go test -v ./...`, then `go run .`.
This separate consumer module pins official STS v2.1.0 and uses `replace ../..` for
the SDK checkout. Record the exact root Git commit. Scripted HTTP responses use
fixed fictional identities/keys; no operation reaches the network or creates resources.
Initial dependency downloads may require network access. Keep comparisons outside
the SDK core; see [tasks and acceptance rules](../../docs/sts-consumer-acceptance.md).

Both SDKs execute four native actions. The generated SDK additionally composes its
role provider/cache with two signed consumer reads; no application response
translation or refresh loop is needed. This workload does not evaluate the official
credentials library's role provider/refresh capabilities, performance or independent
UX. Those cannot be inferred from successful fixture execution. Official v2.1.0
methods here lack context arguments and return Body envelopes; this is a pinned
version observation, not a claim about every official SDK version.

`consumer_review_test.go` is a separate external package using only public SDK
imports and fresh fixtures. Run just this supplemental review with
`go test -count=1 -v -run '^TestConsumerReview' ./...`. It checks caller option/input
isolation, in-flight cancellation, 24 concurrent role-signed reads with one shared
issuance, narrow mocks, anonymous provider/token isolation and non-retrying errors.
See [the supplemental report](../../docs/sts-consumer-review.md). This agent-authored
review is separate from the independent developer task result.

## 中文

Go 1.27+ 下从此目录执行 `go test -v ./...` 和 `go run .`。独立模块固定官方 STS
v2.1.0，通过 `replace ../..` 使用检出 SDK，记录根目录准确 Git 提交。脚本 HTTP
使用虚构身份/密钥，操作不联网、不建资源；初次下载依赖可能需要网络。对比依赖与 SDK
核心隔离，[任务和规则](../../docs/sts-consumer-acceptance.md)详述验收。

两个 SDK 均执行四个原生操作；生成 SDK 另组合 role provider/cache 完成两次签名消费者
读取，不需应用 response translator 或刷新循环。本任务未评估官方 credentials 库的
role provider/刷新能力、性能或独立 UX，不能从 fixture 成功推断。此固定官方 v2.1.0
方法不接 context、输出带 Body envelope，这不是对全部官方版本的断言。

`consumer_review_test.go` 为独立外部包，只导入公开 SDK API，使用新 fixture。
通过上述 `go test -count=1 -v -run '^TestConsumerReview' ./...` 单独执行补充验收，
检查 options/输入隔离、调用中取消、24 个角色签名并发读取复用一次共享签发、小 mock、
匿名 provider/token 隔离与签发错误不重试。见[补充报告](../../docs/sts-consumer-review.md)；
这是代理编写的补充评审，与独立开发者任务结果分开记录。
