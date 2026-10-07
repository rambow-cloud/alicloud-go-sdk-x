# Middleware / 中间件

## English

Use `middleware.Registration` and `middleware.Func` to decorate a client. Initialize surrounds the whole operation; Build runs once; Finalize surrounds signing and transport on every attempt; Deserialize surrounds response decoding on every attempt. Registration order is outer-to-inner. IDs must be nonempty and unique per stage. Client construction copies registrations; implementations remain shared and must be concurrency safe. Call `next` at most once. Exchange is owned by one operation and must not escape; its request is nil before Build and response is nil before transport. Context decoration is propagated. Returning an error short-circuits the pipeline. Do not retain or log raw requests or credentials.

## 中文

通过 `middleware.Registration` 和 `middleware.Func` 装饰客户端。Initialize 包围整个操作；Build 执行一次；Finalize 在每次尝试中包围签名和传输；Deserialize 在每次尝试中包围响应解码。注册顺序为从外到内。ID 必须非空且在同一阶段唯一。客户端构造复制注册项；实现仍共享，必须保证并发安全。最多调用一次 `next`。Exchange 属于一个操作，不得逸出；Build 前 request 为 nil，传输前 response 为 nil。装饰后的 context 会继续传播。返回错误会短路管线。不要保留或记录原始请求或凭据。
