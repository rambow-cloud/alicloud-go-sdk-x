# AWS 风格 review 修复路径

[English](aws-style-remediation.md)

- 本路径写在 abcf961 的 review 修复实现之前。
- Go 调用范式参考 AWS v2，操作名、请求或响应字段、token 值及页语义遵循阿里云 OpenAPI。
- 保留 Go 1.27、JSON v2、显式凭据、默认不重试、30 秒操作期限、HTTPS 规则和标准库核心。
- 本项目独立，不承诺与两个上游 SDK 源码兼容。

| 顺序 | Issue 范围                 | 验收                                                                                                                                           |
| ---- | -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| 1    | 尝试隔离及中断响应重试     | 不发布旧输出/不完整短路成功；EOF/读取失败仅幂等设有次数和时间上限的重试；取消/非幂等/JSON 边界保留                                             |
| 2    | 强类型中间件和服务 Options | Initialize/Serialize/Build 一次，Finalize/Deserialize 每尝试；输入/输出所有权；NewFromConfig、独立服务 Options；调用重试/HTTP 传输实现覆盖隔离 |
| 3    | 分页及生成策略集合         | 重复游标先返回当前页且停止策略可配置；页码防溢出；专属 options、NextPage 调用选项；多个纯 token/纯页码/双模式适配器且不造字段                  |
| 4    | 可复用状态等待器 API       | Wait/WaitForOutput 接收输入；并发等待独立；专属选项、默认 acceptor 可覆盖；保留全 ID 安全及取消                                                |

- 代码前已创建真实 GitHub issues，依赖顺序为 #26 → #27 → #28 → #29。
- 每 issue 独立分支/提交、测试及对应双语文档。
- 通过模板再生成，不直接编辑生成文件。
- 全量本地门禁和同一提交的 Linux race/Windows CI 通过后关闭。
- 基准、ROA/body、默认凭据发现及默认重试/超时策略调整独立跟踪。

- Review 已复现跨重试旧解码输出、重复 token 丢当前页、ECS 旧分页整数溢出及 io.ErrUnexpectedEOF 不重试；回归直接覆盖这些触发条件，并保留生成客户端和 ECS/STS/VPC 契约。

- 迁移说明区分增量及早期 v0 分页/状态等待器修改。
- 保留 New(Config)，NewFromConfig 提供服务函数式选项及校验。
- Token 不解析，空条目不能单独结束 token 分页；纯页码保留服务字段。
- Paginator ClientOptions 每页生效，NextPage 的当前选项优先。
- 重复游标默认先交付当前页再安全停止， 显式禁用防护可能允许循环；失败/取消不前进。
- Wait 返回 error，WaitForOutput 返回成功输出。
- 每次等待及轮询复制输入，覆盖不得修改共享配置。
- AWS 来源与所列参考资料一致，阿里云协议证据固定在 metadata/。

## 参考资料

- [AWS 分页器](https://github.com/aws/aws-sdk-go-v2/blob/main/service/ec2/api_op_DescribeInstances.go)
- [AWS 状态等待器](https://github.com/aws/aws-sdk-go-v2/blob/main/service/ec2/api_op_DescribeInstanceStatus.go)
- [AWS middleware](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/middleware.html)
