# Issue 37: capability policies / Issue 37：能力策略

## English

Actual issue [#37](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/37) follows
#36 / PR #41 under #33. Implement the preceding [paired specification](../capability-policy.md)
on `issue/37-capability-policies`, base `issue/36-batch-go-emission`. Unmerged dependencies
are #41 -> #40 -> #39 -> #32. Initial four paginators, one waiter, explicit read retry,
client token, validators, sensitivity and sparse naming policy do not imply all 579
emitted operations have reviewed capabilities. Keep #33/#38 and dependencies open.

## 中文

真实 issue [#37](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/37) 继 #36 /
PR #41，父任务 #33。先提交[双语规格](../capability-policy.md)，再在
`issue/37-capability-policies` 基于 `issue/36-batch-go-emission` 实现。未合并依赖为
#41 → #40 → #39 → #32。首批四分页、一 waiter、明确读重试、client token、校验、
敏感及命名策略不表示 579 输出操作都已审核能力；父任务/#38/依赖保持打开。
