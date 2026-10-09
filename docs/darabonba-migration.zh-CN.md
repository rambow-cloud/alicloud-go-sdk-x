# 官方 Darabonba 前端迁移

[English](darabonba-migration.md)

- 当前客户端路径以[服务整合 #81](service-consolidation.zh-CN.md)为准：旧 services/ 包和 Go 输出器已移除，generate/check 均检查完整产品。下文旧流程仅保留历史证据。

- 本文件记录 #31 五操作兼容桥；后续以更新的 [product-generator-roadmap.md](product-generator-roadmap.zh-CN.md) 为准，优先于冲突快照/ 补充配置前置要求。
- 产品发现使用完整 DSL、可选规范化元数据及自动可达模型，#31 不等于全量验收。

- 以 aliyun/alibabacloud-sdk 的官方产品 DSL 和 Darabonba 语义解析器为构建期前端，保留本项目审核后的公开 API 子集、IR、Go 后端及共享运行时。
- Node 依赖仅用于开发工具； 继续要求 Go 1.27、直接 JSON v2 和标准库核心。

- 交付顺序：固定真实 ECS/STS/VPC 源码与许可证 → 固定导入模块 → 官方语法/语义检查 → 识别支持的 SDK 请求模式 → 确定性协议/模型投影 → 对照公共元数据和审核策略 → 现有 IR/Go 输出 → 离线验收。
- 保留 ECS DescribeRegions、DescribeInstances、 DescribeInstanceStatus、STS AssumeRole、VPC DescribeVpcs，以及现有分页/状态等待器/ 中间件/错误契约。

- 仅将以下内容转换为 IR：选定函数与可达模型，识别的 OpenApi 调用映射到现有运行时边界；不支持的语句、 选定绑定与类型变化在写文件前明确失败。
- 本次不实现通用 DSL 解释器。
- Overlay 保留 Go 命名、缺失、脱敏、幂等性及状态等待器策略。

- 固定上游 revision、相对路径、SHA-256、许可证、语义解析器/工具版本与实际解析模块。
- 显式 import 可联网，generate/check/Go 测试使用本地产物，CI 离线重导出语义解析器投影； 生成过程中不按通配符下载未固定依赖。

- DSL/元数据偏差经审核，不自动决定来源优先级。
- 区分隐藏/未选字段和已支持公开契约， 分别记录 DSL 可选性与 API 必填性。
- 选定字段有未决实质冲突时阻止生成，决策记录双方版本、事实、处理方式与验证证据。

- 使用 OpenAPI Explorer 链接及其生成 CLI 示例，在已授权本地 Profile 下只读校验并脱敏报告；浏览器结果由用户打开明确链接验证，不表示未观察的 UI 结果。
- 真实调用不进入单测，缺少认证、权限或 STS role 时记为 pending/skip，不能记为通过。
- 决策和指南分别提供中英文文件；issue 和代码注释只使用英文。

- 验收：完整导入模块语义解析；五个真实操作投影；确定性再生成、不支持行为拒绝及篡改检查；保留现有运行时/客户端测试和示例；双语/pkg.go.dev 文档、vet、Go 测试、格式与 CI 再生成通过，发布实际证据。

- 实现跟踪 [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31)： [工具命令与支持范围](../tools/darabonba/README.zh-CN.md)、 [来源/模块/许可锁](../sources/darabonba/README.zh-CN.md)、 [跨来源决策及 Explorer 步骤](darabonba-decisions.zh-CN.md)。

- 2026-10-07 本地验收（Windows、Go 1.27.1、Node 22.21.1）：固定 npm ci 成功，官方离线语义投影检查通过，前端测试 14/14、仓库自动化测试 11/11；13 个公共包 doccheck、 vet、全部 Go 包/Example、sdkgen check、项目 Go 格式、双语结构及 diff 空白检查通过。
- 契约比较确认所有再生成 Go 服务文件与旧后端输出一致。
- Linux race 和 Windows CI 执行相同前端及 Go 门禁，结果在关联 PR 跟踪。
- Explorer 浏览器核验为 NOT RUN， 决策指南提供准确用户步骤。
