# [Feature]: Implement a unified bounded waiter engine and ECS running waiter

## English

### Affected areas

waiter, ecs

### Dependencies

#5

### Problem and scope

Complete the shared runtime foundation before product generation. This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Provide typed polling with success, retry and failure acceptors.
- [ ] Bound total waiting time, propagate cancellation and distinguish waiter expiry from operation errors.
- [ ] Implement an ECS instance-running reference waiter with reviewed state rules.
- [ ] Test success, missing resources, terminal failure, cancellation and timeout using controllable time.

### Documentation and verification

English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.

## 中文

### 领域与依赖

领域：waiter, ecs；依赖：#5。

### 问题与范围

生成产品前完成共享基础。本 issue 交付以下能力，不宣称全量服务覆盖或真实云验证。

### 验收

- [ ] 提供类型化轮询及成功、重试、失败 acceptor。
- [ ] 限制总等待时间，传播取消，区分 waiter 超时和操作错误。
- [ ] 提供具有审核状态规则的 ECS Running waiter。
- [ ] 使用可控时间测试成功、缺失资源、终止失败、取消、超时。

### 文档与验证

英文为主的 Go 注释、外部可执行 Examples、对应双语指南、离线行为测试及相关 CI 检查。
