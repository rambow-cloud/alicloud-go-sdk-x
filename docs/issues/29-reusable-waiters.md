GitHub issue: #29.

## English

### Problem and evidence

Review of abcf961 found behavior defects and incomplete AWS Go SDK v2 design alignment. Design and acceptance were written before code in docs/aws-style-remediation.md.

### Scope and dependencies

Bind only API/options at waiter construction; accept owned input on each Wait/WaitForOutput. Generate dedicated options, ClientOptions and an overridable reviewed acceptor. Wait returns error, WaitForOutput returns success output. Preserve bounded all-ID/page checks, context errors and immutable concurrent waiter use.

Dependency: #28.

### Acceptance criteria

- [ ] Wait and WaitForOutput examples, two independent input sets and concurrent waits, option overrides/defaults, poll ownership, acceptor customization, missing/unknown/duplicate states, cancellation, timeout and error wrapping. Update every caller, public docs and equivalent bilingual migration guidance.
- [ ] English-primary Go docs, deterministic offline Examples, equivalent bilingual docs and migration notes; offline regeneration, doccheck, language, vet, tests, Linux race and Windows CI.
- [ ] Record exact commit and CI evidence before closing.

### Affected areas

waiter, tools, ecs

## 中文

依据 abcf961 review，设计先写在 docs/aws-style-remediation.md，再实现；依赖 #28。

构造只绑定 API/选项，Wait/WaitForOutput 每次接收并复制输入；生成专属选项/ClientOptions/可覆盖审核 acceptor，Wait 返回 error，WaitForOutput 返回成功输出。验收可复用/并发独立输入、选项及复制、acceptor、缺失/未知/重复状态、取消/超时/错误，更新调用者及双语迁移。

英文注释、离线 Example、中英文对应文档与迁移说明、再生成/文档/双语/vet/测试/Linux race/Windows CI 通过后记录提交证据再关闭。默认策略、ROA/body 和基准保持独立。
