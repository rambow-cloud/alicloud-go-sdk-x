# Foundation acceptance mapping

[中文](foundation-acceptance.zh-CN.md)

- Issue #19 gates generator development on actual behavior, not interface presence. #3–#7 and #9–#18 deliver the foundation and language policy.
- GitHub records accepted-commit evidence and CI runs; this guide maps criteria to reviewable artifacts.

| Criterion               | Evidence                                                                                                                                    |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Protocol                | official ACS3 golden vector; ECS/STS reviewed wire fixtures and sources                                                                     |
| Credentials             | coalescing, cancellation isolation, early refresh, expired-source rejection, invalidation, source deadline                                  |
| Retry                   | opt-in/default, unsafe-write rejection, budget/jitter/Retry-After, fresh signatures, canceled backoff                                       |
| HTTP                    | input/config copies, output publication on success, response closure, redirects, size/UTF-8/duplicate JSON failures                         |
| Paginator/waiter        | cursor cycles, empty pages, every-ID Running, terminal states, total/caller deadlines                                                       |
| Mock/testing            | individual operation APIs, function fakes, sdktest script/clock                                                                             |
| Tracing                 | hierarchy, propagation, attempt status; no query/body/credential/error text                                                                 |
| Integration             | TestFoundationSTSCacheRetryPaginationWaiterAndTelemetry: chain → source cache → STS → role cache → ECS → retry → paginator → waiter → spans |
| pkg.go.dev              | twelve public packages with doc.go, exported Go docs, external Examples, canonical module and MIT license                                   |
| Language/dependencies   | paired guides/matrix; Markdown gate; doccheck JSON v2 and standard-library core dependency gate                                             |
| Portability/concurrency | local checks and accepted-commit Linux race/Windows CI                                                                                      |

- Tests are offline with placeholder credentials.
- They verify selected reference operations, not every service protocol or real-cloud behavior.
- Generator #8 can become ready only after this gate passes.
- Benchmarks #20 remain independent.
- Local checks imply neither release tags nor pkg.go.dev indexing.
