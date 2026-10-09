# 补齐 VPC RPC 生成支持

[English](vpc-rpc-completion.md)

## 范围与顺序

- Issue：[#85](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/85)，依赖 #83，是 #61 的前置项。
- 扩展共用的官方 Darabonba parser → IR → Go 流程，支持固定来源中剩余的 7 个 VPC 操作。
- 4 个 VPN 写操作分别使用 URL query 和表单 body。识别准确的 bodyFlat/query/parseToMap 流程，保留字段位置和从 1 开始的数组索引。
- DescribeVpnGatewayAvailableZones 使用整模型 query 转换和 GET，保留原生 HTTP 方法及字段名。
- GrantInstanceToVbr 和 RevokeInstanceFromVbr 显式使用 simple 数组编码。公开输入保持结构化，字符串元素以逗号拼接后再进行 URL 编码。
- 不按操作名称推断规则。未知流程和含糊的绑定在写文件前拒绝。
- 新的字段位置、编码和 HTTP 方法采用产品 IR 及锁文件 schema v3；旧后端在写文件前拒绝不兼容的 IR。
- 重试保持显式启用，上游来源和少量能力策略不变；保留输入副本及上下文取消行为。

## 验收

- 固定 VPC 的发现、降低、生成和编译均达到 403；ECS 保持 380，STS 保持 4 个操作。
- 使用官方来源和重命名夹具验证共用生成规则；反例覆盖错误的转换、body 流程、HTTP 方法及重复绑定。
- 离线签名请求验证 query/body 分离、表单类型、请求体签名、GET、simple 数组、缺失与空值、显式零值、嵌套副本及 HTTP 前取消。
- 在独立消费者模块中对比固定版本的官方表单及 simple helper，无需账号或网络。
- 增加输出确定的外部 Example、英文 Go 注释和配对产品指南。
- 先运行 Node 22 检查及测试，再重新生成，执行 product-check、文档检查、vet、Go 测试及格式检查；CI 保留 Linux race 和 Windows 检查。
- CI 配置确定后，在同一干净的最终任务提交上刷新所有独立消费者记录，保留历史真实调用证据。
- 生成与编译范围不代表全部操作已经真实云验收；发布及索引仍由 #61 跟踪。

## 状态

- VPC 发现、降低、生成均达到 403，包含 1,728 个完整模型，固定来源已无剩余降低缺口。ECS 为 380 个操作、2,053 个模型；STS 为 4 个操作、19 个模型。
- Node 22 前端及 IR 检查、84 项测试通过。product-check、文档检查（15 个包）、vet、根模块测试及 Example 通过。最后补充的版本及位置拒绝测试也已通过。
- 原有公开模型字段类型保持不变：ECS 2,053 个、STS 19 个、VPC 1,671 个模型；上游来源和能力策略未变。
- 格式检查、25 项自动化测试、配对语言及本地链接检查通过，历史真实调用范围保留。
- 消费者共同任务提交为 `72d92d11b5066eb8b6b38a90cf2734c91ef4a1ea`。两个独立模块的 vet、测试及离线程序通过。STS 包含 11 项原有用例及 2 项官方 helper 对比；ECS 为 10 项用例、14 个子用例；VPC 为 8 项用例、10 个子用例。
- 官方 simple 数组及表单展开对比通过，覆盖 nil/空数组、嵌套字段、显式零值及 false、特殊字符。这是 helper 对比，不代表官方 VPC 全部操作的兼容性。
- 机器记录固定到同一任务提交，保留历史真实调用提交。合并前要求最终 PR 的 Linux race、Windows 和自动化检查通过。

- CI 评审修复：运行 37923233731 在 vet 阶段拒绝了反例测试中的重复 JSON tag。已改用不同模型路径在表单展开后产生碰撞，完整 vet 和 RPC 编码测试通过。SDK 及生成产物未变，消费者证据已刷新到上方修正后的任务提交。
