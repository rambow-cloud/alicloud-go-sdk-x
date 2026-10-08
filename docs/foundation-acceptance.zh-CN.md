# 基础验收映射

[English](foundation-acceptance.md)

- Issue #19 以实际行为而非接口存在作为生成器门槛。
- #3–#7、#9–#18 交付基础与语言约束。
- GitHub 记录验收提交证据和 CI；本指南把标准对应到可评审产物。

| 标准              | 证据                                                                                                                                         |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| 协议              | 官方 ACS3 固定向量；ECS/STS 核实的线协议测试数据和来源                                                                                       |
| 凭据              | 合并、取消隔离、提前刷新、拒绝过期来源、失效、来源期限                                                                                       |
| retry             | 显式/默认行为、非安全写拒绝、预算/jitter/Retry-After、新签名、取消退避                                                                       |
| HTTP              | 输入/配置复制、仅成功发布输出、响应关闭、重定向、大小/UTF-8/重复 JSON 失败                                                                   |
| 分页器/状态等待器 | cursor 循环、空页、全部 ID Running、终止状态、总期限/调用者期限                                                                              |
| 测试替身/testing  | 单操作 API、函数 fake、sdktest 脚本/时钟                                                                                                     |
| tracing           | 层级、传播、尝试状态，不含 query/body/凭据/错误文本                                                                                          |
| 集成              | TestFoundationSTSCacheRetryPaginationWaiterAndTelemetry：chain → source cache → STS → role cache → ECS → retry → 分页器 → 状态等待器 → spans |
| pkg.go.dev        | 十二个公共包均有 doc.go、导出文档、外部 Examples、规范 module 与 MIT 许可证                                                                  |
| 语言/依赖         | 配对指南/矩阵；Markdown 门槛；doccheck 的 JSON v2 与核心标准库依赖门槛                                                                       |
| 可移植性/并发     | 本地检查与验收提交的 Linux race/Windows CI                                                                                                   |

- 测试完全离线，使用占位凭据；验证选定参考操作，不证明所有协议或真实云行为。
- 生成器 #8 只有此门槛通过才能 ready。
- 基准 #20 保持独立。
- 本地检查不代表发布标签或 pkg.go.dev 已索引。
