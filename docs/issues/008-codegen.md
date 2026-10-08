# Build product generation after foundation acceptance

- Foundation #19 passed on 28684e47ec1869dd88de0cc14faeec9bc7e6adb3.
- Follow the ordered implementation path in docs/generator.md.

- [ ] #21: [Feature]: Import pinned OpenAPI protocol metadata and validate generator IR
- [ ] #22: [Feature]: Generate typed ECS and STS clients and reviewed paginator/waiter adapters
- [ ] #23: [Maintenance]: Enforce deterministic regeneration and generator acceptance in CI

- The first working profile generates four RPC operations, selected models, codecs, mock interfaces, reviewed pagination/waiter adapters, English Go docs, offline Examples and bilingual guides.
- Offline deterministic generation and CI drift checks are required.
- ROA/general body encoding and wider service coverage require later issues.
- Benchmarks remain #20.

### Affected areas

- tools
