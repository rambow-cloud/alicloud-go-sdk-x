# Issue 37: capability policies

- Actual issue [#37](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/37) follows #36 / PR #41 under #33.
- Implement the preceding [paired specification](../capability-policy.md) on `issue/37-capability-policies`, base `issue/36-batch-go-emission`.
- Unmerged dependencies are #41 -> #40 -> #39 -> #32.
- Initial four paginators, one waiter, explicit read retry, client token, validators, sensitivity and sparse naming policy do not imply all 579 emitted operations have reviewed capabilities.
- Keep #33/#38 and dependencies open.
