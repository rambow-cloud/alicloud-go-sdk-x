GitHub issue: #28.

## English

### Problem and evidence

Review of abcf961 found behavior defects and incomplete AWS Go SDK v2 design alignment. Design and acceptance were written before code in docs/aws-style-remediation.md.

### Scope and dependencies

Return the fetched page before stopping a repeated continuation; expose a safe default and explicit duplicate-stop opt-out. Use overflow-safe arithmetic in both page profiles. Generate dedicated paginator options and per-fetch NextPage options. Support multiple paginator/waiter declarations and token-only/page-only/dual profiles, validating exact selected schema fields.

Dependency: #27.

### Acceptance criteria

- [ ] Current-page delivery on repeated/cyclic tokens, empty token pages, cursor preservation on errors/cancellation, overflow boundaries, per-page overrides and input ownership. Isolated generated token-only client and multiple policies compile offline; incompatible/missing/conflicting declarations fail before writes.
- [ ] English-primary Go docs, deterministic offline Examples, equivalent bilingual docs and migration notes; offline regeneration, doccheck, language, vet, tests, Linux race and Windows CI.
- [ ] Record exact commit and CI evidence before closing.

### Affected areas

core, tools, ecs, vpc

## 中文

依据 abcf961 review，设计先写在 docs/aws-style-remediation.md，再实现；依赖 #27。

重复游标先交付当前页，默认安全停止且可显式禁用；统一页码防溢出；生成分页专属选项及 NextPage 覆盖；支持多分页/waiter 集合和纯 token/纯页码/双模式，校验选定字段。验收当前页/循环/空 token 页、失败取消游标、防溢出、选项/复制、离线生成编译及非法声明拒绝。

英文注释、离线 Example、中英文对应文档与迁移说明、再生成/文档/双语/vet/测试/Linux race/Windows CI 通过后记录提交证据再关闭。默认策略、ROA/body 和基准保持独立。
