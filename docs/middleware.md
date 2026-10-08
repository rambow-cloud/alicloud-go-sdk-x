# Middleware

[中文](middleware.zh-CN.md)

- Serialize runs once between Initialize and Build.
- Generated clients provide an owned typed Exchange.Input; Initialize and pre-next Serialize hooks can change that model before validation/encoding.
- Region routing may be changed through Exchange.Region; explicit call options take precedence over the original input region.
- Deserialize hooks can inspect or change typed Exchange.Output after next.
- A replacement must be the exact non-nil operation output pointer type.
- Successful short circuits must supply that output; otherwise ErrIncompleteOperation is returned.
- Output is committed only after all enclosing hooks succeed.
- Wire Invoke has nil Input and a typed decoded Output.

- Use `middleware.Registration` and `middleware.Func` to decorate a client.
- Initialize surrounds the whole operation; Build runs once; Finalize surrounds signing and transport on every attempt; Deserialize surrounds response decoding on every attempt.
- Registration order is outer-to-inner.
- IDs must be nonempty and unique per stage.
- Client construction copies registrations; implementations remain shared and must be concurrency safe.
- Call `next` at most once.
- Exchange is owned by one operation and must not escape; its request is nil before Build and response is nil before transport.
- Context decoration is propagated.
- Returning an error short-circuits the pipeline.
- Do not retain or log raw requests or credentials.
