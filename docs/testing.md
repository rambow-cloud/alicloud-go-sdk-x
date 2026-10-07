# Offline testing / 离线测试

## English

Inject an `http.Client` with `sdktest.NewTransport` to script service responses or transport errors. `Step.Check` asserts signed wire parameters and can reject requests. Scripts and headers are copied; checks remain shared and must be concurrency safe. Exhaustion fails and never opens a network connection. The transport counts attempts but does not retain sensitive requests. `RoundTripperFunc` supports custom behavior. `Clock.Sleep` advances virtual time immediately and honors cancellation; it does not advance real context timers. Negative advances panic. ECS/STS packages expose individual operation interfaces, allowing small handwritten fakes instead of mocking an entire service. Runtime, retry, pagination and waiter tests use these helpers without real credentials or cloud accounts.

## 中文

注入使用 `sdktest.NewTransport` 的 `http.Client`，脚本化服务响应或传输错误。`Step.Check` 验证签名后的线协议参数，可拒绝请求。脚本和 header 被复制；检查函数共享，必须并发安全。脚本耗尽报错，绝不连接网络。传输统计尝试次数，但不保留敏感请求。`RoundTripperFunc` 支持自定义行为。`Clock.Sleep` 立即推进虚拟时间并尊重取消，但不推进真实 context 定时器。负数推进会 panic。ECS/STS 提供单操作接口，用小型手写 fake 替代整个服务的 mock。运行时、重试、分页和 waiter 测试使用这些 helper，无需真实凭据或云账号。
