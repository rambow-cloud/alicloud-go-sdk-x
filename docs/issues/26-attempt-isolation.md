GitHub issue: #26.

## English

### Problem and evidence

Review of abcf961 found behavior defects and incomplete AWS Go SDK v2 design alignment. Design and acceptance were written before code in docs/aws-style-remediation.md.

### Scope and dependencies

Clear and commit decoded state per successful attempt. Reject successful short circuits without an output. Retry EOF/unexpected EOF and explicitly transient body reads only under the bounded idempotency policy. Preserve cancellation, JSON failures, atomic output and safe diagnostics.

Dependency: #25 (completed).

### Acceptance criteria

- [ ] Offline regressions reproduce stale middleware/retry output, incomplete successful short circuits, interrupted body retry/success, no-retry/non-idempotent/canceled/JSON failures, closure of response bodies and fresh metadata.
- [ ] English-primary Go docs, deterministic offline Examples, equivalent bilingual docs and migration notes; offline regeneration, doccheck, language, vet, tests, Linux race and Windows CI.
- [ ] Record exact commit and CI evidence before closing.

### Affected areas

core, transport

## 中文

依据 abcf961 review，设计先写在 docs/aws-style-remediation.md，再实现；依赖 #25 (completed)。

修复每尝试输出隔离，拒绝缺少输出的成功短路；EOF/明确瞬态响应读取仅在有界幂等策略下重试。回归覆盖旧输出、短路、中断成功/默认不重试/非幂等/取消/JSON、body 关闭及元数据。

英文注释、离线 Example、中英文对应文档与迁移说明、再生成/文档/双语/vet/测试/Linux race/Windows CI 通过后记录提交证据再关闭。默认策略、ROA/body 和基准保持独立。
