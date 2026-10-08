# 离线测试

[English](testing.md)

- 注入使用 `sdktest.NewTransport` 的 `http.Client`，脚本化服务响应或传输错误。
- `Step.Check` 验证签名后的线协议参数，可拒绝请求。
- 脚本和 header 被复制；检查函数共享，必须并发安全。
- 脚本耗尽报错，绝不连接网络。
- 传输统计尝试次数，但不保留敏感请求。
- `RoundTripperFunc` 支持自定义行为。
- `Clock.Sleep` 立即推进虚拟时间并尊重取消，但不推进真实 context 定时器。
- 负数推进会 panic。
- ECS/STS 提供单操作接口，用小型手写 fake 替代整个服务的测试替身。
- 运行时、重试、分页和状态等待器测试使用这些辅助组件，无需真实凭据或云账号。
