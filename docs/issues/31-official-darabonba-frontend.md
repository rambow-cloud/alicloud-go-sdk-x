# Official Darabonba frontend / 官方 Darabonba 前端

## English

Tracks [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31), depending on
accepted #29/#30. The development path was established in
[darabonba-migration.md](../darabonba-migration.md) before implementation.

Problem: metadata-only generation lacks official SDK request binding and model
evidence. Reuse real official ECS/STS/VPC DSL, official parser/import tools and pinned
modules, retaining our IR, Go backend and unified runtime. Scope is the existing five
operations; unsupported selected behavior must fail before writes.

Acceptance: full imported-module semantic checking; deterministic offline projection
and regeneration; reviewed metadata/DSL decisions; checksum and unsupported-pattern
rejection; unchanged Go APIs and runtime contracts; Go 1.27/JSON v2, pkg.go.dev comments,
Examples and paired docs; frontend checks/tests, Go gates and CI on Linux/Windows.

Implementation: official parser frontend and build-only tool lock, pinned source/module/
license evidence, protocol/binding/model projections, strict cross-source IR checks,
machine-readable divergence approvals, and CI frontend regeneration. Source and
projection corruption, changed types/bindings/protocol/requiredness, missing operations
and unreviewed inputs are exercised before output writes. Prior runtime behavior is
retained and checked against the metadata-only backend.

Evidence and remaining browser verification are recorded in
[decisions](../darabonba-decisions.md). Benchmark #20 and broader DSL profiles remain
separate. Current state and verification results are maintained on the GitHub issue.

## 中文

跟踪 [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31)，依赖已验收的
#29/#30；实现前已建立[迁移路径](../darabonba-migration.md)。

问题是纯元数据生成缺少官方 SDK 请求绑定及模型证据。复用真实官方 ECS/STS/VPC
DSL、parser/导入工具及固定模块，保留本项目 IR、Go 后端、统一运行时；范围为现有
五个操作，不支持的选定行为在写文件前失败。

验收包含完整导入模块语义检查、确定性离线投影/再生成、审核后的元数据/DSL 决策、
哈希和不支持模式拒绝、不改变 Go API/运行时契约、Go 1.27/JSON v2、pkg.go.dev 注释、
Example/双语文档，以及前端/Go 门禁和 Linux/Windows CI。

实现包含官方 parser 前端、构建工具锁、来源/模块/许可证据、协议/绑定/模型投影、
跨来源 IR 严格检查、机器可读偏差批准和 CI 前端再生成。测试覆盖来源/投影篡改、
类型/绑定/协议/必填变化、操作缺失及未审核字段，在写入前拒绝；与旧元数据后端
对比确保保留运行时行为。

证据及未完成浏览器核验见[决策](../darabonba-decisions.md)。基准 #20 及更多 DSL
模式独立跟踪，当前状态和验证结果维护在 GitHub issue。
