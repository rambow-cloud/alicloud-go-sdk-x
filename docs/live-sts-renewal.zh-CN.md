# 真实 STS 续期

[English](live-sts-renewal.md)

### 范围与前置条件

- 用户于 2026-10-08 授权本地沙箱外验证真实 STS 续期，并创建专用验证角色。
- 基线为 f4cebddd3660d0d09d2b56e6319f5895c2f4d4a5，显式使用本地 Aliyun CLI oss-sftp Profile 和 cn-hangzhou。
- 带超时的 CLI 预检确认 IdentityType=Account，主账号不能直接 AssumeRole。
- 创建唯一命名、无控制台登录的 RAM 用户、一枚临时 AccessKey 和只可扮演一个新角色的策略； 角色仅信任该用户，另一个策略仅允许 ecs:DescribeRegions。
- 认证秘密/token 只在进程内存， 不进参数/文件；密钥标识仅在清理时用作删除参数。
- 私有清理记录只跟踪本次创建对象，测试后删除密钥、授权、用户、角色及两份策略。
- 不改变既有身份、权限、资源；用户新指令仅对此专用设置替代旧“不创建”限制。

- 先建立独立 issue/分支。
- 执行一次带超时的 CLI GetCallerIdentity 预检，本地处理响应，账号身份仅存忽略的私有初始化资料，公开只展示安全身份类型/状态。
- 允许 CLI 已有 OAuth 刷新，本地检查管理 Profile 缓存期限必须覆盖创建/清理， 否则需重新登录。
- 主账号 Profile 仅用于创建/清理专用设置；生成 STS 使用显式 RAM 用户来源。
- 权限失败如实记录，不扩大权限、不用未改变的凭据重复尝试。

### 验证顺序

1. 生成 service/sts 使用独立来源；NewAssumeRoleProviderFromClient 使用授权角色、会话及
   可选 identity，DurationSeconds=900，采用真实时间。只统计签发次数，不输出密钥、token、
   header/body、角色/账号/资源标识或原始错误。
2. 包装 credentials.Cache，获取有效角色快照，调用生成 ECS DescribeRegions；验证 AssumeRole
   返回的身份/元数据以及重复读取不再签发。ECS 权限失败与 STS 失败
   分别记录，不为通过验收扩大权限。
3. 显式 Invalidate 一次，验证第二次真实签发、内存比较凭据变化的布尔值、有效过期时间和角色
   认证读取。此项称为强制刷新，不称为自然续期。
4. Cache 保持空闲，等待第二份凭据的真实服务端过期时间加小余量。使用可中断、最多 60 秒
   的等待间隔，持续报告进度。然后不手动失效、不替换时间、不写重复签发循环，直接读取/
   使用 cache；验证新签发、新未来过期时间、不返回旧过期快照及生成客户端认证读取成功。
5. 保留独立来源身份、输入所有权和结构化安全错误。只公开脱敏时间/次数/状态、SDK/CLI/Go
   版本及提交；一次性脚本、本地报告仅放忽略的 .git/。

- 固定完整 DSL STS 当前只输出 AssumeRole，没有 GetCallerIdentity；角色身份断言使用 AssumedRoleUser ARN，认证消费者使用生成 ECS。
- CLI 预检单独记录，不表示存在生成的 GetCallerIdentity。
- [角色扮演前提](https://www.alibabacloud.com/help/en/ram/user-guide/assume-a-ram-role) 说明专用 RAM 用户来源与单角色 sts:AssumeRole 策略的必要性。

- [官方 STS FAQ](https://www.alibabacloud.com/help/en/ram/support/faq-about-ram-roles-and-sts-tokens) 规定最短 900 秒，签发新 token 不会撤销旧 token。
- 本实验只证明限定真实签发/复用/强制刷新/ 自然到期续期，不证明 SDK 原生 Profile/OAuth 刷新、所有角色/信任场景、独立体验或整体 Beta。
- Explorer 浏览器证据单独记录，可打开 [AssumeRole](https://api.aliyun.com/api/Sts/2015-04-01/AssumeRole) 核查角色/会话/时长字段，不分享返回凭据。

### 验证记录

- [#55](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/55)，2026-10-08。
- SDK 提交 f4cebddd3660d0d09d2b56e6319f5895c2f4d4a5；Go 1.27.1 / Windows amd64，Aliyun CLI 3.4.11，cn-hangzhou。
- 开始时间为 2026-10-08T09:40:30.8039616Z，来源和消费者分别显式配置凭据提供者，无生产代码、生成产物或依赖变更。

| 用例               | 结果    | 证据                                                                                                                    |
| ------------------ | ------- | ----------------------------------------------------------------------------------------------------------------------- |
| CLI 管理来源预检   | PASS    | 主账号身份、有效 OAuth 管理快照，期限覆盖创建/清理                                                                      |
| 专用设置           | PASS    | 一个无控制台登录用户、一枚临时来源密钥、一个角色、两份窄策略，回读确认只信任该用户                                      |
| 真实签发与角色使用 | PASS    | 生成 AssumeRole、匹配返回角色身份、有效密钥/token/未来期限、HTTP 200/request ID/一次尝试、生成 ECS 读取成功             |
| 缓存复用           | PASS    | 再次读取凭据/ECS 仍只有一次签发                                                                                         |
| 强制刷新           | PASS    | 一次 Invalidate 引发第二次真实签发，密钥/token 变化，角色认证读取成功                                                   |
| 自然到期续期       | PASS    | 空闲至 2026-10-08T09:55:30Z 加 5 秒，下一次读取自动第三次签发，新期限 2026-10-08T10:10:34Z，角色签名的生成 ECS 读取成功 |
| 来源/输入所有权    | PASS    | 独立来源和原始输入不变，总计三次签发、四次 ECS 读取                                                                     |
| 清理               | PASS    | 临时密钥删除，两份授权解除，用户/角色及两份自定义策略删除，不保留测试身份                                               |
| Explorer 浏览器    | NOT RUN | 与 SDK/CLI 分开记录，人工链接见上文                                                                                     |

- 自然续期完成时间 2026-10-08T09:55:35.4615676Z，不使用虚拟时钟、再次 Invalidate 或手写签发循环。
- 强制刷新与自然续期分别观察；本结果不证明提前/后台刷新、真实并发刷新、所有角色/信任/条件、 SDK 原生 Profile/OAuth 刷新、独立体验或整体 Beta，离线契约继续保留各自证据。

- 一次性本地脚本 SHA256（不作为生产 SDK 发布）：Go 40267af9bf8aa2afa07f708799c637690201bc557129477623522e2682efd0c6；PowerShell bac2ff7b24a8ee264da3f9c1106854766d34a31146361f15a617b2b80b949b12。
- 文档/GitHub 证据不含密钥、token、账号/角色/用户标识或原始响应体。
