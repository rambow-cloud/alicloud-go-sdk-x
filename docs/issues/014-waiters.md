# [Feature]: Implement a unified bounded waiter engine and ECS running waiter

### Affected areas

- waiter, ecs

### Dependencies

- #5

### Problem and scope

- Complete the shared runtime foundation before product generation.
- This issue delivers the capability described below; no full service coverage or live-cloud validation is claimed.

### Acceptance criteria

- [ ] Provide typed polling with success, retry and failure acceptors.
- [ ] Bound total waiting time, propagate cancellation and distinguish waiter expiry from operation errors.
- [ ] Implement an ECS instance-running reference waiter with reviewed state rules.
- [ ] Test success, missing resources, terminal failure, cancellation and timeout using controllable time.

### Documentation and verification

- English-primary Go comments, external executable Examples, equivalent English/Chinese guides, offline behavior tests and relevant CI checks.
