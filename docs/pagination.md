# Unified pagination / 统一分页

## English

`pagination.Paginator[T]` uses a typed fetcher, comparable Cursor and explicit HasMore. Token and page-number adapters share the engine; adapters decide termination from service metadata. Empty pages may continue when a token exists. Fetch/context failures preserve the cursor; cycles and repeated cursors stop with ErrRepeatedCursor. Exhaustion returns ErrNoMorePages without fetching. Paginators belong to one consumer and are not concurrency safe. `ecs.NewDescribeInstancesPaginator` copies input/slices/options and defaults to token mode with MaxResults=10. Explicit PageNumber or PageSize selects legacy mode; modes cannot mix. Token completion ignores TotalCount because the service declares it meaningless in that mode. Legacy completion uses page size and TotalCount; an empty page stops. Original requests and returned pages are not reused as mutable fetch input. See the [official paging contract](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeinstances).

## 中文

`pagination.Paginator[T]` 使用类型化 fetcher、可比较 Cursor 和明确 HasMore。Token 和页码适配器共享引擎；适配器根据服务元数据决定终止。有 token 时空页仍可继续。Fetch/context 失败保留 cursor；循环和重复 cursor 以 ErrRepeatedCursor 终止。耗尽返回 ErrNoMorePages，不再获取。分页器属于单个消费者，不保证并发安全。`ecs.NewDescribeInstancesPaginator` 复制输入、slice、options，默认使用 MaxResults=10 的 token 模式。显式 PageNumber 或 PageSize 选择旧页码模式；禁止混用。Token 模式不使用 TotalCount，因为官方声明该模式下它无意义。旧页码模式根据 page size 和 TotalCount 终止；空页停止。原始请求和返回页不会被复用为可变 fetch 输入。参考[官方分页契约](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeinstances)。
