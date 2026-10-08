# [Feature]: Extend RPC generation with scalar presence, object lists and local references

- GitHub issue: #24.

# [Feature]: Extend RPC generation with scalar presence, object lists and local references

### Problem and evidence

- The first generator only covers selected ECS/STS shapes.
- VPC DescribeVpcs introduces meaningful absent/false boolean filters, repeatList tag objects, nested responses and page-only pagination.
- Source: https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs.
- Design and acceptance: docs/generator-expansion.md (written before code).

### Scope

- Implement optional scalar pointers and booleans, named repeatList input models with recursive copying, nested response projections with JSON v2 methods, bounded offline local schema references, model dependency checks and a page-only paginator profile.
- Keep existing APIs and protocol tests.
- No generator service branches.

### Dependencies

- #23

### Acceptance criteria

- [ ] Test explicit false/empty/zero versus nil, input pointer/slice/model copying, nested response round trips and atomic failure, local-ref chains/escapes/dangling/cyclic/external/sibling errors, schema drift and page bounds/overflow. English Go docs, offline Examples and equivalent bilingual docs; regeneration and full CI.
- [ ] Record exact commit and successful CI evidence before closure; no live account tests.

### Affected areas

- tools
