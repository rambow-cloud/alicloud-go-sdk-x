# [Feature]: Generate typed ECS and STS clients and reviewed paginator/waiter adapters

- GitHub issue: #22.

# [Feature]: Generate typed ECS and STS clients and reviewed paginator/waiter adapters

### Problem and evidence

- Foundation #19 passed.
- The four reference operations remain handwritten.
- Official source: https://help.aliyun.com/zh/sdk/product-overview/openapi-metadata/.
- Development path: docs/generator.md; parent #8.

### Scope

- Emit formatted models, request codecs, operation methods, minimal mock interfaces, client construction, English Go docs, deterministic offline Examples and bilingual product guides from validated IR.
- Generate reviewed pagination/waiter adapters and sensitive model redaction.
- Replace four handwritten reference operations while retaining prose-only validation extensions.
- Delegate all shared capabilities to the foundation runtime; preserve current public API and endpoint scope.

### Dependencies

- #21

### Acceptance criteria

- [ ] Existing ECS/STS protocol fixtures and paginator/waiter/STS helper tests must pass unchanged. Test operation encoding, shared runtime idempotency and a synthetic service. No live accounts or network calls in generated Examples.
- [ ] Original English-primary code comments and equivalent English/Chinese Markdown updated together.
- [ ] Record exact commit and relevant successful verification before closure.

### Affected areas

- tools, ecs, sts
