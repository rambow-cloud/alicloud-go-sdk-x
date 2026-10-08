# Offline testing

[中文](testing.zh-CN.md)

- Inject an `http.Client` with `sdktest.NewTransport` to script service responses or transport errors. `Step.Check` asserts signed wire parameters and can reject requests.
- Scripts and headers are copied; checks remain shared and must be concurrency safe.
- Exhaustion fails and never opens a network connection.
- The transport counts attempts but does not retain sensitive requests. `RoundTripperFunc` supports custom behavior. `Clock.Sleep` advances virtual time immediately and honors cancellation; it does not advance real context timers.
- Negative advances panic.
- ECS/STS packages expose individual operation interfaces, allowing small handwritten fakes instead of mocking an entire service.
- Runtime, retry, pagination and waiter tests use these helpers without real credentials or cloud accounts.
