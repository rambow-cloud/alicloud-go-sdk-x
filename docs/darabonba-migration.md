# Official Darabonba frontend migration / 官方 Darabonba 前端迁移

## English

This records #31's five-operation compatibility bridge. The newer
[product-generator-roadmap.md](product-generator-roadmap.md) governs subsequent work and
supersedes conflicting snapshot/overlay prerequisites. Product discovery uses complete DSL,
optional normalized metadata and automatic reachable models; #31 is not full-product acceptance.

Use the official product DSL corpus at aliyun/alibabacloud-sdk and the official
Darabonba parser as the build-time frontend. Keep our reviewed public API subset,
IR, Go emitter and shared runtime. Node dependencies are development tools only;
Go 1.27, direct JSON v2 and the standard-library runtime remain required.

Delivery order: pin real ECS/STS/VPC source/license -> pin imported modules ->
official syntax and semantic analysis -> recognize supported SDK request patterns
-> deterministic protocol/model projection -> cross-check public metadata and
reviewed policy -> existing IR/Go emission -> offline acceptance. Preserve ECS
DescribeRegions, DescribeInstances, DescribeInstanceStatus, STS AssumeRole and
VPC DescribeVpcs, including all existing paginator/waiter/middleware/error contracts.

Lower selected functions and reachable models. Recognized OpenApi calls map to
our runtime boundary; unsupported statements, selected bindings and type changes
fail before output writes. A general DSL interpreter is outside this migration.
Overlay policy retains Go names, presence, redaction, idempotency and waiters.

Pin upstream revision, relative paths, SHA-256, license, parser/tool versions and
resolved modules. Explicit import can access the network; generate/check and Go
tests use local artifacts. CI re-exports parser projections offline. Dependency
wildcards cannot trigger unpinned downloads during generation.

Review DSL/metadata differences without automatic precedence. Distinguish hidden
or unselected fields from supported public contracts. Record DSL optionality and
API requiredness separately. Material unresolved selected conflicts block generation.
Decisions record both revisions, facts, disposition and verification evidence.

Use OpenAPI Explorer links and generated CLI examples with the authorized local
profile for read-only behavior checks and sanitized reports. The user verifies
browser results by opening documented links; do not claim unobserved UI results.
Live checks stay outside unit tests. Missing authentication, permissions or an STS
role is pending/skipped, never passed. Keep English/Chinese decisions and guides
equivalent; issues and code comments are English-primary.

Acceptance: imported-module semantic parsing; real-source projections for all five
operations; deterministic regeneration, unsupported behavior rejection and tamper
checks; existing runtime/client tests and examples preserved; bilingual/pkg.go.dev
docs, vet, Go tests, formatting and CI regeneration pass. Publish actual evidence.

Implementation under [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31):
[tool commands and bounded profile](../tools/darabonba/README.md),
[source/module/license lock](../sources/darabonba/README.md), and
[cross-source decisions and Explorer steps](darabonba-decisions.md).

Recorded local acceptance, 2026-10-07 (Windows, Go 1.27.1, Node 22.21.1): pinned
npm ci succeeded; official offline semantic projection check passed; frontend tests
14/14 and repository automation tests 11/11 passed; doccheck passed for 13 public
packages, vet and all Go packages/examples passed; sdkgen check, project Go formatting,
bilingual structure and diff whitespace checks passed. The contract comparison confirms
all regenerated Go service files equal the prior backend output. Linux race and Windows
CI run the same frontend plus Go gates; their results are tracked on the linked PR.
Explorer browser verification is NOT RUN, with precise user steps in the decision guide.

## 中文

本文件记录 #31 五操作兼容桥；后续以更新的
[product-generator-roadmap.md](product-generator-roadmap.md) 为准，优先于冲突 snapshot/
overlay 前置要求。产品发现使用完整 DSL、可选规范化元数据及自动可达模型，#31 不等于全量验收。

以 aliyun/alibabacloud-sdk 的官方产品 DSL 和 Darabonba parser 为构建期前端，保留
本项目审核后的公开 API 子集、IR、Go 后端及共享运行时。Node 依赖仅用于开发工具；
继续要求 Go 1.27、直接 JSON v2 和标准库核心。

交付顺序：固定真实 ECS/STS/VPC 源码与许可证 → 固定导入模块 → 官方语法/语义检查
→ 识别支持的 SDK 请求模式 → 确定性协议/模型投影 → 对照公共元数据和审核策略 →
现有 IR/Go 输出 → 离线验收。保留 ECS DescribeRegions、DescribeInstances、
DescribeInstanceStatus、STS AssumeRole、VPC DescribeVpcs，以及现有分页/waiter/
middleware/错误契约。

仅降低选定函数与可达模型，识别的 OpenApi 调用映射到现有运行时边界；不支持的语句、
选定绑定与类型变化在写文件前明确失败。本次不实现通用 DSL 解释器。Overlay 保留
Go 命名、缺失、脱敏、幂等性及 waiter 策略。

固定上游 revision、相对路径、SHA-256、许可证、parser/工具版本与实际解析模块。
显式 import 可联网，generate/check/Go 测试使用本地产物，CI 离线重导出 parser 投影；
生成过程中不按通配符下载未固定依赖。

DSL/元数据偏差经审核，不自动决定来源优先级。区分隐藏/未选字段和已支持公开契约，
分别记录 DSL 可选性与 API 必填性。选定字段有未决实质冲突时阻止生成，决策记录双方
版本、事实、处理方式与验证证据。

使用 OpenAPI Explorer 链接及其生成 CLI 示例，在已授权本地 Profile 下只读校验并
脱敏报告；浏览器结果由用户打开明确链接验证，不宣称未观察的 UI 结果。真实调用不
进入单测，缺少认证、权限或 STS role 时记为 pending/skip，不能记为通过。决策与指南
中英文对应，issue 与代码注释以英文为主。

验收：完整导入模块语义解析；五个真实操作投影；确定性再生成、不支持行为拒绝及篡改
检查；保留现有运行时/客户端测试和示例；双语/pkg.go.dev 文档、vet、Go 测试、格式与
CI 再生成通过，发布实际证据。

实现跟踪 [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31)：
[工具命令/有界模式](../tools/darabonba/README.md)、
[来源/模块/许可锁](../sources/darabonba/README.md)、
[跨来源决策及 Explorer 步骤](darabonba-decisions.md)。

2026-10-07 本地验收（Windows、Go 1.27.1、Node 22.21.1）：固定 npm ci 成功，官方
离线语义投影检查通过，前端测试 14/14、仓库自动化测试 11/11；13 个公共包 doccheck、
vet、全部 Go 包/Example、sdkgen check、项目 Go 格式、双语结构及 diff 空白检查通过。
契约比较确认所有再生成 Go 服务文件与旧后端输出一致。Linux race 和 Windows CI
执行相同前端及 Go 门禁，结果在关联 PR 跟踪。Explorer 浏览器核验为 NOT RUN，
决策指南提供准确用户步骤。
