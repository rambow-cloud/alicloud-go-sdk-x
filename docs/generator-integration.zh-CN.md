# 生成器评审与 main 集成

[English](generator-integration.md)

- 用户授权的评审与合并于 2026-10-08（北京时间）完成。
- 兼容桥、五个产品生成阶段和评审修复已集成到 `main`，实现合并提交为 `19e9107e876b2309808511ad0dbced633cb54d7d`。
- 父 issue [#33](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/33) 跟踪此验收及文档。
- 旧 issue 草稿和规格中的分支/依赖描述属于实现历史，不表示仍待合并。

## 合并与 CI 记录

| Issue | 评审 PR                                                          | 评审提交  | main 合并提交 | 通过的 CI                                                                                 |
| ----- | ---------------------------------------------------------------- | --------- | ------------- | ----------------------------------------------------------------------------------------- |
| #31   | [#32](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/32) | `92095fe` | `33bfa4a`     | [37614002853](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37614002853) |
| #34   | [#39](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/39) | `292bb5d` | `1ca9cf8`     | [37615648634](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37615648634) |
| #35   | [#40](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/40) | `7427e10` | `59508cc`     | [37623726048](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37623726048) |
| #36   | [#41](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/41) | `4afa9ff` | `1f1ed96`     | [37629756588](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37629756588) |
| #37   | [#42](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/42) | `629e810` | `fd27fe9`     | [37635577018](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37635577018) |
| #38   | [#43](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/43) | `7098116` | `5f507a5`     | [37645170995](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37645170995) |
| #44   | [#45](https://github.com/rambow-cloud/alicloud-go-sdk-x/pull/45) | `859f8f6` | `19e9107`     | [37699342435](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/runs/37699342435) |

- 评审覆盖来源/导入固定、严格的 IR 转换检查、元数据表示规范化、完整发现/可达 IR、批量强类型 RPC 输出、输入副本与产物保护、原生能力、授权说明和通知。
- 各 PR 的历史评审记录以英文为主，并附中文解释。
- #44 的角色重叠缺口已由 #45 修复：同一模型内的冲突在渲染前失败，跨模型 token 同名及独立适配器共用字段仍合法。
- 当前固定协议范围没有剩余合并阻塞。

- 使用 merge commit 保留各 issue 提交，按依赖顺序合并，并将后续 PR 基分支改为 `main`。
- 各次 main 文件树均与评审提交相同；最终实现文件树为 `11e36c09e3493c40c5f07f8f998fe1ac946e9173`，等于 #45 已验证内容。
- PR 证据含 Linux race、Windows、Node/前端、Go 测试/Example、doccheck、vet 和再生成。
- 仅取消本批仍运行且已被替代的 main push CI，保留最终快照检查；文档集成更新单独执行双语/ 格式检查及 PR CI，不重复未改变的本地 SDK 测试。

- 接受的路线仍为完整固定官方 Darabonba DSL → 官方语义解析器 → 规范化 IR → 本项目 Go 后端 → 公共运行时，无逐操作元数据白名单或 Tea 运行时依赖。
- 可选元数据补充表示，少量策略定义审核行为，不逐项重述模型/字段。

## 当时的覆盖范围

| 产品 | 已发现 | 已输出 | 不支持 | 输出模型 | 英文操作说明 | 英文字段说明 |
| ---- | ------ | ------ | ------ | -------- | ------------ | ------------ |
| ECS  | 380    | 283    | 97     | 1,453    | 274/283      | 2,401/6,476  |
| VPC  | 403    | 295    | 108    | 1,240    | 295/295      | 3,482/6,489  |
| STS  | 4      | 1      | 3      | 5        | 1/1          | 14/19        |
| 合计 | 787    | 579    | 208    | 2,698    | 570/579      | 5,897/12,984 |

- `service/` 提供完整支持模型，`services/` 保留五操作参考桥。
- 四个原生分页器、一个 ECS running 状态等待器和七项已审核操作策略已验收，572 个输出操作的相关能力仍未审核。
- 请求或响应字段/原生游标采用阿里云语义，调用范式参考 AWS；重试显式启用且保守。
- 英文 pkg.go.dev 注释、离线 Example、Apache 条款/通知和双语使用/契约/来源指南已交付，缺说明及缺中文语义翻译明确报告。

- 此验收覆盖固定 RPC、编译和离线契约；更多协议/产品、能力审核、基准 #20、真实 Explorer 证据、版本发布及 pkg.go.dev 索引另行验收。
- 来源许可及发布边界保留在来源/发布指南中。
