# Product acceptance definition

## English

### Affected areas

core, testing, tools

### Problem and evidence

The user wants a usable modern replacement for Alibaba Cloud Go SDK v2. Existing foundation and generator acceptance is recorded in docs/foundation-acceptance.md and docs/generator-integration.md, but generated/compiled counts do not establish developer experience, live behavior or a release. The Smithy experiment is local and does not replace the accepted Darabonba route.

### Scope and dependencies

Establish docs/product-acceptance.md before additional implementation: stable criterion IDs, explicit candidate Beta products/actions/region, developer tasks, evidence/status rules and foundation/Beta/release gates. Link it from README, development-path and AGENTS.md. Preserve existing accepted evidence and scope; live #47 and benchmark #20 remain separate. This issue delivers the definition, not overall Beta acceptance.

### Acceptance criteria

- [ ] Equivalent English/Chinese product acceptance, including all eleven shared capabilities and generation/maintenance/documentation.
- [ ] Candidate Beta scope and reproducible consumer tasks, with proposed first-call target and comparison protocol.
- [ ] PASS/FAIL/SKIP/NOT RUN, evidence dimensions, source versions and blockers recorded without fabricated status.
- [ ] Record current gaps, including full-DSL STS provider composition, incomplete live waiter/token/AssumeRole evidence, comparative UX/performance, upstream update rehearsal and publication.
- [ ] Wire the agreement into project constraints and issue index; preserve production architecture.
- [ ] Bilingual/format/link/diff checks; documentation-only changes do not repeat unchanged SDK tests.

## 中文

### 领域、问题与范围

领域为 core、testing、tools。用户要求现代且好用的官方 v2 替代品；基础/生成记录存在，但生成/编译数量不足以证明体验、真实行为或发布。Smithy 仅本地实验，不改变现有 Darabonba 路线。

实现前建立 docs/product-acceptance.md，固定标准编号、候选 Beta 产品/操作/地域、用户任务、证据/状态规则和基础/Beta/发布门槛，并链接 README、开发路径及 AGENTS。保留已接受证据及范围；真实验证 #47、基准 #20 分开。本 issue 交付标准定义，不表示 Beta 通过。

### 验收

- [ ] 中英对应，覆盖十一项基础能力及生成/维护/文档。
- [ ] 明确候选 Beta 和可复现用户任务、建议首次调用目标及对比方法。
- [ ] 记录 PASS/FAIL/SKIP/NOT RUN、不同证据维度、来源版本和阻塞，不编造通过状态。
- [ ] 记录完整 DSL STS provider 接入、真实 waiter/token/AssumeRole、体验/性能、上游升级和发布缺口。
- [ ] 写入项目约束与 issue 索引，不改变生产架构。
- [ ] 双语/格式/链接/diff 检查；纯文档不重复未改 SDK 测试。
