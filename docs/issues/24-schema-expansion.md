GitHub issue: #24.

# [Feature]: Extend RPC generation with scalar presence, object lists and local references

## English

### Problem and evidence

The first generator only covers selected ECS/STS shapes. VPC DescribeVpcs introduces meaningful absent/false boolean filters, repeatList tag objects, nested responses and page-only pagination. Source: https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs. Design and acceptance: docs/generator-expansion.md (written before code).

### Scope

Implement optional scalar pointers and booleans, named repeatList input models with recursive copying, nested response projections with JSON v2 methods, bounded offline local schema references, model dependency checks and a page-only paginator profile. Keep existing APIs and protocol tests. No generator service branches.

### Dependencies

#23

### Acceptance criteria

- [ ] Test explicit false/empty/zero versus nil, input pointer/slice/model copying, nested response round trips and atomic failure, local-ref chains/escapes/dangling/cyclic/external/sibling errors, schema drift and page bounds/overflow. English Go docs, offline Examples and equivalent bilingual docs; regeneration and full CI.
- [ ] Record exact commit and successful CI evidence before closure; no live account tests.

### Affected areas

tools

## 中文

首版生成器仅覆盖 ECS/STS 子集，VPC 带来布尔缺失/false、tag 对象数组、嵌套响应和纯页码需求；设计先于代码，见 docs/generator-expansion.md。依赖 #23。

实现指针标量/布尔、递归复制的具名 repeatList 输入模型、嵌套响应 JSON v2、离线有界本地引用、模型依赖及纯页码 profile。验收存在语义、复制、原子 JSON、引用失败、schema 漂移和页边界/溢出，保留既有 API，不添加服务分支。

关闭前记录真实提交与成功 CI，不使用真实账号测试。ROA/body、基准 #20、发布为后续阶段。
