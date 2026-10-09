# Issue index

[中文](README.zh-CN.md)

- [Full-DSL live evidence](../product-live-validation.md): Historical full-DSL ECS/VPC read-only evidence (#47): two image pages; empty instance/VPC pages; live waiter skipped. Product acceptance #74/#75 remains separate.

- The first release is [v0.1.0 STS](../sts-v0.1.0.md), tracked by [milestone v0.1.0](https://github.com/rambow-cloud/alicloud-go-sdk-x/milestone/4) and [Project 3](https://github.com/orgs/rambow-cloud/projects/3).
- Native child issues belong to #57; parent membership is not a blocking dependency.

| Issue | v0.1.0 work                                                                 | Dependencies                                              |
| ----- | --------------------------------------------------------------------------- | --------------------------------------------------------- |
| #57   | [Scoped STS release parent](57-sts-v010-roadmap.md)                         | Aggregates #58-#61; remains open through release/indexing |
| #58   | [Requestless GetCallerIdentity](58-requestless-sts-generation.md)           | Accepted generator baseline                               |
| #59   | [Anonymous OIDC/SAML RPC](59-anonymous-sts-generation.md)                   | Accepted generator baseline; independent of #58           |
| #60   | [Four-operation acceptance and source rehearsal](60-sts-v010-acceptance.md) | #58, #59, accepted #51/#53/#55                            |
| #61   | [Release and pkg.go.dev](61-sts-v010-release.md)                            | #60                                                       |

- GitHub owns current state; this table records historical and current dependencies, not current blocking state.
- The authoritative full-DSL route is tracked by #33, with #34 -> #35 -> #36 -> #37 -> #38. #31 remains the compatibility bridge.
- Older accepted foundation/generator issues are history, not new per-operation prerequisites.

| Issue | Capability                                                                                   | Dependencies                                                        |
| ----- | -------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| #10   | [Docs]: Define the runtime-first development path and bilingual documentation policy         | None                                                                |
| #11   | [Feature]: Implement a shared staged middleware pipeline                                     | None                                                                |
| #12   | [Feature]: Implement replaceable endpoint resolution with explicit rules                     | None                                                                |
| #13   | [Feature]: Provide structured operation errors and response metadata                         | None                                                                |
| #14   | [Feature]: Implement a unified bounded waiter engine and ECS running waiter                  | #5                                                                  |
| #15   | [Feature]: Provide small mock interfaces and deterministic testing helpers                   | None                                                                |
| #16   | [Feature]: Add a typed STS AssumeRole client and credential provider helper                  | #3, #9                                                              |
| #17   | [Feature]: Add optional OpenTelemetry operation and attempt instrumentation                  | #3, #11, #13                                                        |
| #18   | [Feature]: Add an explicit composable credential chain                                       | None                                                                |
| #19   | [Maintenance]: Validate the unified runtime foundation before enabling generator development | #3, #4, #5, #6, #7, #9, #10, #11, #12, #13, #14, #15, #16, #17, #18 |
| #20   | [Maintenance]: Establish reproducible runtime and dependency comparison benchmarks           | #19                                                                 |
| #21   | Pinned metadata importer and validated IR                                                    | #19                                                                 |
| #22   | Generated clients, models and paginator/waiter adapters                                      | #21                                                                 |
| #23   | Deterministic regeneration and CI acceptance                                                 | #22                                                                 |
| #24   | Scalar presence, object lists, local references and page-only generation                     | #23                                                                 |
| #25   | Generated VPC DescribeVpcs client and paginator                                              | #24                                                                 |
| #26   | Isolated attempt output and interrupted response retries                                     | #25                                                                 |
| #27   | Typed middleware and concrete service Options                                                | #26                                                                 |
| #28   | Native paginator options and multiple policy profiles                                        | #27                                                                 |
| #29   | Reusable Wait/WaitForOutput and waiter options                                               | #28                                                                 |
| #30   | Live local-profile reads and Explorer CLI comparison                                         | #29                                                                 |
| #31   | Official Darabonba semantic frontend and reviewed DSL/metadata decisions                     | #29, #30                                                            |
| #33   | Full-DSL product generator roadmap (parent)                                                  | Stages #34-#38                                                      |
| #34   | Source representation normalization                                                          | #31                                                                 |
| #35   | Complete DSL discovery, reachable IR and coverage                                            | #34                                                                 |
| #36   | Batch Go emission from product IR                                                            | #35                                                                 |
| #37   | Sparse capability policies                                                                   | #36                                                                 |
| #38   | Licensed bilingual/pkg.go.dev documentation automation                                       | #37                                                                 |
| #44   | Reject aliased pagination and waiter policy roles                                            | #38                                                                 |
| #49   | Scoped product Beta/release criteria and developer experience definition                     | #33, #44; #47 evidence tracked separately                           |
| #51   | Full-DSL STS client -> provider/cache -> generated consumer composition                      | #49; accepted foundation and generator                              |
| #53   | STS-first application guidance and explicit credential-provider configuration                | #49, #51                                                            |
| #55   | Dedicated-role live STS issuance, cache reuse, forced refresh and real expiry renewal        | #49, #51, #53                                                       |
