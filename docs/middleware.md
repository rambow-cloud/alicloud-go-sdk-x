# Middleware / 中间件

## English

Serialize runs once between Initialize and Build. Generated clients provide an owned
typed Exchange.Input; Initialize and pre-next Serialize hooks can change that model
before validation/encoding. Region routing may be changed through Exchange.Region;
explicit call options take precedence over the original input region. Deserialize
hooks can inspect or change typed Exchange.Output after next. A replacement must be
the exact non-nil operation output pointer type. Successful short circuits must supply
that output; otherwise ErrIncompleteOperation is returned. Output is committed only
after all enclosing hooks succeed. Wire Invoke has nil Input and a typed decoded Output.

Use `middleware.Registration` and `middleware.Func` to decorate a client. Initialize surrounds the whole operation; Build runs once; Finalize surrounds signing and transport on every attempt; Deserialize surrounds response decoding on every attempt. Registration order is outer-to-inner. IDs must be nonempty and unique per stage. Client construction copies registrations; implementations remain shared and must be concurrency safe. Call `next` at most once. Exchange is owned by one operation and must not escape; its request is nil before Build and response is nil before transport. Context decoration is propagated. Returning an error short-circuits the pipeline. Do not retain or log raw requests or credentials.

## 中文

Serialize 在 Initialize 与 Build 之间执行一次。生成客户端提供自有的类型化 Exchange.Input；
Initialize 及调用 next 前的 Serialize hook 可在校验/编码前修改模型。地域路由可通过
Exchange.Region 调整；显式调用选项优先于原始输入地域。Deserialize hook 在 next 后检查或修改
Exchange.Output，替换必须为操作的准确非 nil 输出指针类型。成功短路必须提供该输出，否则返回
ErrIncompleteOperation。所有外层 hook 成功后才发布输出；线 Invoke 的 Input 为 nil，Output 为类型化解码结果。

通过 `middleware.Registration` 和 `middleware.Func` 装饰客户端。Initialize 包围整个操作；Build 执行一次；Finalize 在每次尝试中包围签名和传输；Deserialize 在每次尝试中包围响应解码。注册顺序为从外到内。ID 必须非空且在同一阶段唯一。客户端构造复制注册项；实现仍共享，必须保证并发安全。最多调用一次 `next`。Exchange 属于一个操作，不得逸出；Build 前 request 为 nil，传输前 response 为 nil。装饰后的 context 会继续传播。返回错误会短路管线。不要保留或记录原始请求或凭据。
