# Official Darabonba frontend migration

[中文](darabonba-migration.zh-CN.md)

- This records #31's five-operation compatibility bridge.
- The newer [product-generator-roadmap.md](product-generator-roadmap.md) governs subsequent work and supersedes conflicting snapshot/overlay prerequisites.
- Product discovery uses complete DSL, optional normalized metadata and automatic reachable models; #31 is not full-product acceptance.

- Use the official product DSL corpus at aliyun/alibabacloud-sdk and the official Darabonba parser as the build-time frontend.
- Keep our reviewed public API subset, IR, Go emitter and shared runtime.
- Node dependencies are development tools only; Go 1.27, direct JSON v2 and the standard-library runtime remain required.

- Delivery order: pin real ECS/STS/VPC source/license -> pin imported modules -> official syntax and semantic analysis -> recognize supported SDK request patterns -> deterministic protocol/model projection -> cross-check public metadata and reviewed policy -> existing IR/Go emission -> offline acceptance.
- Preserve ECS DescribeRegions, DescribeInstances, DescribeInstanceStatus, STS AssumeRole and VPC DescribeVpcs, including all existing paginator/waiter/middleware/error contracts.

- Lower selected functions and reachable models.
- Recognized OpenApi calls map to our runtime boundary; unsupported statements, selected bindings and type changes fail before output writes.
- A general DSL interpreter is outside this migration.
- Overlay policy retains Go names, presence, redaction, idempotency and waiters.

- Pin upstream revision, relative paths, SHA-256, license, parser/tool versions and resolved modules.
- Explicit import can access the network; generate/check and Go tests use local artifacts.
- CI re-exports parser projections offline.
- Dependency wildcards cannot trigger unpinned downloads during generation.

- Review DSL/metadata differences without automatic precedence.
- Distinguish hidden or unselected fields from supported public contracts.
- Record DSL optionality and API requiredness separately.
- Material unresolved selected conflicts block generation.
- Decisions record both revisions, facts, disposition and verification evidence.

- Use OpenAPI Explorer links and generated CLI examples with the authorized local profile for read-only behavior checks and sanitized reports.
- The user verifies browser results by opening documented links; do not claim unobserved UI results.
- Live checks stay outside unit tests.
- Missing authentication, permissions or an STS role is pending/skipped, never passed.
- Keep English/Chinese decisions and guides equivalent; issues and code comments use English.

- Acceptance: imported-module semantic parsing; real-source projections for all five operations; deterministic regeneration, unsupported behavior rejection and tamper checks; existing runtime/client tests and examples preserved; bilingual/pkg.go.dev docs, vet, Go tests, formatting and CI regeneration pass.
- Publish actual evidence.

- Implementation under [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31): [tool commands and bounded profile](../tools/darabonba/README.md), [source/module/license lock](../sources/darabonba/README.md), and [cross-source decisions and Explorer steps](darabonba-decisions.md).

- Recorded local acceptance, 2026-10-07 (Windows, Go 1.27.1, Node 22.21.1): pinned npm ci succeeded; official offline semantic projection check passed; frontend tests 14/14 and repository automation tests 11/11 passed; doccheck passed for 13 public packages, vet and all Go packages/examples passed; sdkgen check, project Go formatting, bilingual structure and diff whitespace checks passed.
- The contract comparison confirms all regenerated Go service files equal the prior backend output.
- Linux race and Windows CI run the same frontend plus Go gates; their results are tracked on the linked PR.
- Explorer browser verification is NOT RUN, with precise user steps in the decision guide.
