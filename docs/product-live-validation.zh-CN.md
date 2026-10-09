# 完整 DSL SDK 真实调用验收

[English](product-live-validation.md)

- 对应 [#47](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/47)。
- 本文保留历史只读证据，于 2026-10-09 评审后合入。
- 本次使用静态凭据快照，不替代后来原生 Profile/OAuth 的证据，也不代表 #74/#75 产品验收已完成。

### 范围与顺序

本验收计划先于执行建立，针对基线 `dde481ce4da680f964ee3bbd1458048e6ed981bc`
上的完整 DSL `service/ecs` 和 `service/vpc` 包。此前[真实验证记录](live-validation.md)
覆盖有界 `services/` 兼容桥，不能证明新包已真实验收。依赖 #33、#44 已完成；先建立独立
GitHub issue，再在 issue 分支执行一次性脚本。若发现 runtime/generator 缺陷，修改生产
代码前明确对应 issue 范围。

仅使用用户授权的本地 Aliyun CLI `oss-sftp` Profile 及其配置地域。先执行官方 CLI 参考，
允许其已有 OAuth 刷新流程，再将有效本地凭据快照显式注入 SDK static provider。
检查期限，不输出凭据、不通过命令参数传递凭据；不为 SDK 新增原生 Profile 加载或 OAuth 刷新。

### 验收标准

- 同参数比较 ECS DescribeRegions、DescribeImages、DescribeInstances 原生 token/页码模式、
  DescribeInstanceStatus 及 VPC DescribeVpcs。
- 使用审核范围内的小分页大小，每种最多两页。比较选定类型字段、可选字段存在性、原生游标/
  页元数据以及 HTTP/request ID/尝试元数据；等值比较排除每调用不同的 request ID 和时间。
  区分解码/协议失败、权限失败及可能的资源变化。
- 检查输入所有权与预取消 context。若有界状态样本含已有 Running 实例，使用其标识和有界
  超时验证 Wait/WaitForOutput，否则记录跳过。不为获取更多分页或状态转换创建、修改资源。
- 一次性脚本及原始 CLI 响应仅保留在忽略的 `.git/`。公开脱敏数量、通过/失败/跳过、
  版本、源码提交及限制；不包含凭据、账号/资源标识、名称及原始响应体。
- 发布双语证据与关联 issue 的 PR；明确第一页结束及跳过项。这些读取不证明全部 579 个操作、
  真实重试故障、凭据刷新、tracing、STS AssumeRole 或 waiter 状态转换。

### 网页核验

SDK/CLI 执行和 OpenAPI Explorer 网页证据分别记录。可选人工检查时，登录同一账号、选择
Profile 地域并打开配对英文指南中的五个准确 API 链接；检查 CLI Example、对应 RegionId 与原生
token/页码参数、成功响应容器和游标元数据。网页结果不计入自动化 SDK/CLI 验收，只有用户
独立提供后才记录。

### 证据

2026-10-08 执行，SDK 源码提交为 `dde481ce4da680f964ee3bbd1458048e6ed981bc`，
Go 1.27.1 windows/amd64、Aliyun CLI 3.4.11，显式选择 cn-hangzhou 的 `oss-sftp`
OAuth Profile。本次验收仅修改文档；一次性脚本导入新 `service/ecs`、`service/vpc`，
不是旧 `services/` 兼容桥。凭据为检查期限后的本地 static 快照，未新增公共 provider。

最初脚本选择了旧 `sftp-oss` Profile，其缓存凭据已过期；CLI 在 OAuth token 刷新时
初始化失败（HTTP 400），没有取得 ECS 响应。检查后确认用户新登录的是 `oss-sftp`，
成功执行前已修正计划、issue 和脚本。这是本地验收 Profile 选择修正，不是生产 SDK
修复；公开记录不含凭据值和原始诊断。

| 检查                             | 结果   | 证据及限制                                                                                               |
| -------------------------------- | ------ | -------------------------------------------------------------------------------------------------------- |
| ECS DescribeRegions              | 通过   | 33 个地域；选定线路径字段及标量存在性与 CLI 相符                                                         |
| ECS DescribeImages 页码分页      | 通过   | ImageOwnerAlias=system、PageSize=2；两页共四个不同镜像标识，more=true；原生页元数据及选定标量与 CLI 相符 |
| ECS DescribeInstances token 分页 | 通过   | MaxResults=1；第一页为空且结束，未覆盖真实 token 跨页                                                    |
| ECS DescribeInstances 页码分页   | 通过   | PageSize=1；第一页为空且结束                                                                             |
| ECS DescribeInstanceStatus 分页  | 通过   | PageSize=1；第一页为空且结束                                                                             |
| VPC DescribeVpcs 分页            | 通过   | PageSize=1；第一页为空且结束                                                                             |
| 请求及分页输入所有权             | 通过   | 所有成功读取/分页前后，调用者输入序列化内容保持一致                                                      |
| 预取消 context                   | 通过   | ECS/VPC 操作及镜像分页器保留 errors.Is(context.Canceled)，分页游标未被消费                               |
| ECS Wait / WaitForOutput         | 跳过   | 有界状态样本中没有 Running 实例                                                                          |
| STS AssumeRole                   | 跳过   | 未提供明确目标 role                                                                                      |
| OpenAPI Explorer 网页 UI         | 未执行 | 仅 SDK/CLI 证据；人工核验链接见上文                                                                      |

每个成功 SDK 响应均为 HTTP 200、非空 request ID 且与类型化 RequestId 相符、恰好一次
尝试。等值比较独立解码的 CLI 线路径字段和类型化 SDK 输出的 JSON 投影，保留选定标量的
存在性；request ID 仅在单次响应内部检查，不在不同请求之间比对。页内投影排序，opaque
token 只比续取存在性；下一页设计为 CLI/SDK 各用自己的 token，但账号实例响应没有续取，
因此这一路径未在真实环境执行。每个分页器最多两页，检查原生续取及跨页重复标识，不遍历
完整公共镜像目录。

本次证明 DescribeImages 的真实页码续取，以及账号样本读取的空结果解码/分页结束；不证明
非空实例/VPC 模型覆盖、实例 token 跨页、waiter 轮询或状态转换、自动重试、凭据刷新、
tracing、其他地域、全部生成操作或 pkg.go.dev 发布。原始参考及脚本仅在忽略的
`.git/product-live-validation/`，未创建或修改云资源。
