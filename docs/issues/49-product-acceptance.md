# Product acceptance definition

### Affected areas

- core, testing, tools

### Problem and evidence

- The user wants a usable modern replacement for Alibaba Cloud Go SDK v2.
- Existing foundation and generator acceptance is recorded in docs/foundation-acceptance.md and docs/generator-integration.md, but generated/compiled counts do not establish developer experience, live behavior or a release.
- The Smithy experiment is local and does not replace the accepted Darabonba route.

### Scope and dependencies

- Establish docs/product-acceptance.md before additional implementation: stable criterion IDs, explicit candidate Beta products/actions/region, developer tasks, evidence/status rules and foundation/Beta/release gates.
- Link it from README, development-path and AGENTS.md.
- Preserve existing accepted evidence and scope; live #47 and benchmark #20 remain separate.
- This issue delivers the definition, not overall Beta acceptance.

### Acceptance criteria

- [ ] Equivalent English/Chinese product acceptance, including all eleven shared capabilities and generation/maintenance/documentation.
- [ ] Candidate Beta scope and reproducible consumer tasks, with proposed first-call target and comparison protocol.
- [ ] PASS/FAIL/SKIP/NOT RUN, evidence dimensions, source versions and blockers recorded without fabricated status.
- [ ] Record current gaps, including full-DSL STS provider composition, incomplete live waiter/token/AssumeRole evidence, comparative UX/performance, upstream update rehearsal and publication.
- [ ] Wire the agreement into project constraints and issue index; preserve production architecture.
- [ ] Bilingual/format/link/diff checks; documentation-only changes do not repeat unchanged SDK tests.
