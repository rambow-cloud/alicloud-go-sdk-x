# DSL 与元数据决策

[English](darabonba-decisions.md)

- Issue [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31)，2026-10-07。
- DSL revision 为 ec489e5c3deae95496daae2b41503ac58b221adb，公共元数据版本及原始/提取 SHA-256 在 metadata/{ecs,sts,vpc}/manifest.json。
- 机器审核记录在 metadata/darabonba-decisions.json，由各产品 manifest 固定；前端和 Go 检查均拒绝新输入范围/必填差异，选定线类型、路径、绑定变化也失败。
- 修改批准记录必须有 issue 证据并同步本文对应语言。
- 批准的 DSL 独有字段必须保持可选；必填差异只允许 API 必填、DSL 可选，反转方向需要建立新的支持策略。

| 偏差                                                                                        | 审核处理                                                                                                                  | 行为证据                                                   |
| ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| ECS/VPC DSL 多出 OwnerAccount、OwnerId、ResourceOwnerAccount、ResourceOwnerId，公共快照没有 | 不进入公开子集；DSL 出现不等于支持契约                                                                                    | 源码比较；未真实调用 owner 覆盖                            |
| DescribeInstances DSL 使用重复 Filter 模型，快照列 Filter.1.Key/Value 到 Filter.4.Key/Value | #34 规范化八个绑定，移除 Filter 的 DSL 独有批准；保留旧 Go 子集，不由索引推断服务上限                                     | 两来源及离线叶子/类型/大小写/必填测试；Filter 未真实测试   |
| DescribeInstances.Tag 元数据包含当前 DSL 没有的可选小写 key/value                           | metadataOnlyFields 明确批准可选字符串 Tag[].key 和 Tag[].value，保留大小写、不进入公共子集；其他缺成员或类型/必填变化失败 | 固定 canonical/公共快照和语义 DSL 对照；未真实调用弃用字段 |
| DescribeInstances、DescribeInstanceStatus、DescribeVpcs 的 RegionId 在 DSL 可选、元数据必填 | 保留 API 必填及既有配置地域回退                                                                                           | 离线缺地域检查和此前 #30 配置地域真实读取                  |
| AssumeRole 的 RoleArn/RoleSessionName 在 DSL 可选、元数据必填                               | 保留 HTTP 前本地必填校验；DSL 表达可 unset，不表示省略后成功                                                              | 既有 STS 离线测试；未提供明确 role，真实调用跳过           |
| 元数据允许 GET/POST，DSL 固定 POST                                                          | 使用 DSL POST，保留既有签名 RPC 编码                                                                                      | 既有签名协议测试和此前 #30 读取；不公开 GET                |
| DSL 响应包裹 headers/statusCode/body，元数据描述 body                                       | 投影 body，由共同运行时提供响应元数据/结构化错误                                                                          | 既有响应/错误/中间件测试和此前 #30 读取                    |

- 以上是明确的 SDK 契约决策，不表示隐藏字段无效，也不表示已实验确认服务端必填性。
- API 必填规则来自固定公共元数据，属于保守客户端校验。
- 全产品及模块通过官方语义解析，IR 转换仅支持已声明模式。
- Go 数值位宽、nullable/缺值、JSON 字符串数组、时间转换及 API 约束仍由审核后的元数据/补充配置管理，DSL 说明不自动转为 validator。

- 新增实质冲突时：保留两个版本，以同参数通过 Explorer/其 CLI 示例复现，区分服务端行为和 CLI 本地校验，记录脱敏结果及日期，决定支持契约，更新批准记录、测试、双语文档后再生成。
- 不静默选择某个来源以掩盖选定线类型/路径冲突。

- 打开参考资料中列出的五个准确 Explorer 链接，使用相同授权账号和地域：DescribeRegions 检查公开字段，不发 owner 覆盖；DescribeInstances 检查 RegionId 和 Filter 可见性，分别验证 MaxResults=10 与 PageNumber=1/PageSize=10；DescribeInstanceStatus 检查 RegionId 必填和页响应字段；DescribeVpcs 检查 RegionId 及原生页元数据；AssumeRole 检查两个必填字段，成功调试需明确授权目标 role。

- 浏览器核验状态为 **NOT RUN**。
- UI schema 提示、CLI 示例、CLI 本地校验和实际 HTTP 响应分别记录。
- 此前真实 SDK/CLI 证据见 [live-validation.md](live-validation.zh-CN.md)，早于本次前端迁移，未验证 owner、Filter 或省略必填参数。
- 测试 TestOfficialDSLJoinsAllProductsWithoutChangingGoContracts 将再生成 Go 服务代码与旧元数据后端逐字节比较，不把旧运行记录当作本次迁移或 Explorer 浏览器的新通过结果。

## 参考资料

- [DescribeRegions](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeRegions)
- [DescribeInstances](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstances)
- [DescribeInstanceStatus](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstanceStatus)
- [DescribeVpcs](https://api.aliyun.com/api/Vpc/2016-04-28/DescribeVpcs)
- [AssumeRole](https://api.aliyun.com/api/Sts/2015-04-01/AssumeRole)
