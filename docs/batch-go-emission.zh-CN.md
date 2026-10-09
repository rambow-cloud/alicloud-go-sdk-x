# 批量 Go 输出

[English](batch-go-emission.md)

- 当前生成范围见 [RPC 扩展 #83](dsl-rpc-expansion.zh-CN.md)：ECS 380/380、VPC 396/403。下方原有数量及消费者记录对应当时验收的提交。

- 阶段 [#36](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/36) 在 #35 / PR #40 之后读取完整且哈希固定的 `models/*/ir.json`；本规格先于实现。
- 官方 DSL/语义 IR 即可生成，不依赖旧元数据、逐操作字段/模型选择或手写文档补充配置。

- 实现前规格提交为 `ce18878`。
- 当前输出 ECS 283/380 操作及 1,453 模型、VPC 295/403 及 1,240 模型、STS 1/4 及 5 模型，总计 579 操作、2,698 模型。
- 每个输出操作有确定性外部 Example。
- 本地完整 Go 测试包含临时独立 module 编译及签名 HTTP/中间件契约；doccheck 已通过 16 公共包。
- Linux race/Windows 验收在关联 PR 单独记录。
- 渲染报告刻意将编译/真实/策略标为未评估，不把依赖机器环境的编译成功写入确定性源码产物。
- 仓库根目录生成、只读检查及严格选择检查命令见本文件的命令与参考资料。

- 完整生成包使用单数 `service/<product>`，遵循 AWS 风格的服务导入。
- #81 已移除复数 services/ 兼容桥及输出器，统一使用 service/ 客户端，见[迁移说明](service-consolidation.zh-CN.md)。
- 迁移需明确修改导入、可选标量指针、原生响应容器； DSL 字符串字段（例如 JSON 编码的 IDs）仍传字符串。
- 不表示上游源码兼容；新产品能力适配器属于 #37。

- 自动输出每个已转换为 IR 的操作及完整可达模型、Input/Output、以 context 为首个参数的方法、服务 Options/NewFromConfig、小操作 API 接口。
- Input 为请求根；Output 为 HTTP JSON 响应体加运行时 Metadata。
- DSL 响应 envelope 模型也保留，但不是方法返回包装。
- 可选标量/模型使用指针：nil 省略，非 nil 保留零/false/空字符串。
- 数组/map 保留元素类型、数值宽度和准确线大小写；首字母缩写/匿名类型名确定性，命名冲突失败而非丢字段。
- DSL 可选性不等于 API 必填，本阶段不猜必填校验或重试安全。

- 私有标准库 RPC 辅助组件在 Initialize 前深复制所有指针、slice/map，使 hooks 得到独占模型；递归编码保留点分成员、从 1 起数组索引、原字符串，省略 nil。
- context、 签名、endpoint、hooks、响应限制、结构化错误及重试复用公共运行时；所有新操作在 #37 审核策略前保守禁止重试，不记录凭据/请求响应体。

- 编码参考：[官方 Go Query 实现](https://github.com/alibabacloud-go/openapi-util/blob/master/service/service.go) 递归展开对象成员，数组使用从 1 起索引；于 2026-10-07 阅读作为行为证据，不作为固定构建/运行时依赖。
- 本后端为独立标准库实现，并提供离线契约测试。

- 离线命令为 `sdkgen product-generate`、只读 `product-check`。
- 显式 `-operations product/Name,...` 验证支持情况，不悄悄缩小生成产品集合。
- 不支持/未知选择、哈希损坏、无效形状/协议、命名冲突在写前失败；预检全部输出并拒绝链接/ 无生成标记文件。
- 新旧生成各自管理产物。
- 确定性报告区分发现/转换为 IR/生成代码和不支持原因；编译和真实验证是独立证据，不因渲染成功自动宣称。

- 验收包括确定性再生成、真实产品完整数量、临时独立 module 编译、写前失败/过期产物清理、离线签名 HTTP 的存在语义/重复及嵌套参数/大小写/完整响应容器/hooks 复制/调用选项隔离/取消/错误、小型接口测试替身、英文 Go 文档及确定性外部操作 Example 和对应中文指南、Go 1.27/JSON v2、前端检查测试、格式/doccheck/两种再生成检查/ vet/测试及 Linux race/Windows CI。
- 无需真实云调用，仍有不支持操作时不表示 ECS/ VPC/STS 全覆盖。

- 分页历史：共享引擎在 main 直接提交 `61d2581`（issue #7），原生生成适配器/选项提交 `89e1d07`（issue #28），没有独立历史 PR。
- 新产品完整分页器/状态等待器策略/ 适配器在本次输出 PR 后的 #37 PR 评审；明确输入/输出游标、结果集合路径、页大小/ 默认值及终止规则，不凭字段或操作名猜能力。

## 可运行命令与示例

```sh
go run ./internal/cmd/sdkgen product-generate
go run ./internal/cmd/sdkgen product-check
go run ./internal/cmd/sdkgen product-check -operations ecs/DescribeImages,sts/AssumeRole
```
