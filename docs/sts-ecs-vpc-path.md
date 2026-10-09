# STS, ECS and VPC delivery path

[中文](sts-ecs-vpc-path.zh-CN.md)

## Authority

- The user revised the route on 2026-10-09: finish #60, then complete ECS and VPC acceptance before #61 publication.
- This replaces the earlier STS-only first-release schedule. Keep its accepted evidence and historical decisions.
- Use the same pinned official Darabonba DSL, semantic parser, complete IR, shared Go backend and runtime.
- Keep Go 1.27, JSON v2, AWS-style calling conventions and Alibaba-native wire semantics.
- Do not publish a tag or release as part of STS closeout.

## Order

1. Complete #60 against the latest accepted SDK revision. Record the actual reviewer type and task results; never label agent execution as independent human evidence.
2. Complete scoped ECS product acceptance #74: complete generated RPC inventory plus reviewed paginator, waiter, retry and consumer workloads.
3. Complete scoped VPC product acceptance #75: complete generated RPC inventory plus reviewed pagination, retry and consumer workloads.
4. Run #61 only after the required product gates pass. Publish the declared scope and inspect same-version pkg.go.dev pages.

## Product boundaries

- The current pinned baseline emits four STS, 283 ECS and 296 VPC actions. These are generation counts, not live coverage.
- STS includes all four actions, the provider/cache and native Profile composition. STS has no pagination or waiter.
- ECS starts with DescribeRegions, DescribeImages, DescribeInstances and DescribeInstanceStatus consumer contracts; retain native page/token modes and InstanceRunningWaiter semantics.
- VPC starts with DescribeVpcs and its native page-number paginator. Do not invent a token API or waiter.
- Review policies against complete IR. Unlisted operations remain unreviewed for retry and adapters; discover raw operations without per-operation authoring.
- Account for every discovered operation: lowered, emitted, compiled, independently tested and live are separate statuses. Unsupported protocol shapes have explicit reasons and fail selected generation before writes.
- Product acceptance must declare its workloads, required cases and exclusions before execution. Compilation alone does not complete a product.
- Existing #47/PR #48 provides a separate read-only ECS/VPC evidence track. Review its current facts before integration; do not promote empty pages or skipped waiter checks into continuation/transition success.
- Reuse accepted #55 renewal, #60 identity/source rehearsal and #68 native OAuth evidence unless relevant behavior changes.
- Live federation, new protocols, arbitrary resource writes, whole-cloud parity and benchmark #20 remain separate scopes.

## Evidence and coordination

- Product issues contain the acceptance matrix, dependencies, verification and paired documentation requirements.
- #61 depends on #60 and both product acceptance issues. Parent #57 and milestone v0.1.0 remain open through publication/indexing.
- Keep actual GitHub dependencies acyclic. Parent membership is not a blocking dependency for children.
- Maintain one issue status label and the matching Project status. Issues and PRs use English only.
- Human usability evidence remains distinct from agent consumer tests. The user chose agent acceptance for #60. Optional human UX is #76; only real human results update the human record.
- Source, runtime, policy and generated-file changes remain issue-driven. Update both language guides in the same change.
