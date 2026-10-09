# 产品能力策略

[English](capability-policy.md)

- 当前 #88 扩展支持原生 int64/字符串 token 分页限制、明确排除的废弃分页输入、同一操作的多个生命周期 waiter 及原生单 ID waiter。具体范围见[能力扩展](product-capability-expansion.zh-CN.md)；未列出的行为仍为未审核。

- [评审修复 #44](capability-role-review.zh-CN.md) 拒绝同一模型内的分页请求页码/页大小/limit、 响应页码/页大小/总数、状态等待器请求页码/页大小及成员 ID/状态重叠；不同请求/响应模型或独立适配器之间的同名路径仍有效。
- 路径存在且类型匹配不足以证明行为角色各不相同。

- 阶段 #37 继 #36 / PR #41，在 `issue/37-capability-policies` 叠加 `issue/36-batch-go-emission`；原依赖链及后续 #44 现已[集成 main](generator-integration.zh-CN.md)。
- 本规格先于代码。
- `policies/` 为少量审核例外，不是全部 API/模型/字段清单；绑定产品、版本和来源 manifest 哈希，各操作策略有证据。
- 无策略表示未审核、不生成能力、不被 Standard 重试。
- 未知/ 不支持操作、路径、类型、名称、模式、约束在写前失败；缺策略不妨碍完整支持 Go 输出。

- 首批为 ECS DescribeInstances（默认 token，显式页码字段选页码）、DescribeInstanceStatus、 DescribeImages 及 VPC DescribeVpcs 分页；ECS InstanceRunningWaiter；上述四查询及 DescribeRegions 的读取幂等策略；AllocateDedicatedHosts ClientToken 生成/校验；STS AssumeRole 敏感模型格式化保护。
- 其余操作保持未审核，报告区分已审核策略、生成适配器、 保守重试及真实验收。

- 分页策略声明准确线游标路径、集合路径、limit/页码/大小/总数、适配器默认值及上限， 对完整 IR 解析为带 nil 保护的强类型访问，不在运行时猜字段。
- 构造器遵循 NewOperationPaginator 并为无效选项返回 error，HasMorePages/NextPage(ctx, 操作选项) 复用公共引擎。
- 构造/每次抓取深复制输入，复制选项且单页覆盖不延续。
- 错误/取消不推进； 重复 token 默认先交付当页再停止，关闭保护可循环。
- 空 token 页仍有游标则继续，不以 TotalCount 判断 token 结束。
- 页码按审核总数/集合判断，缺总数、负元数据、页码不符或响应页大小超上限失败，空集合终止；算术受原生 int32 上限保护。
- 双模式拒绝非 nil 的 token/limit 与页码/大小混用，包括显式空/零指针。
- 状态分页刻意使用 50，区别于服务默认 10；其余首批默认 10，状态/VPC 上限 50，镜像/实例上限 100。

- Waiter 策略声明 ID 输入、原生集合及成员 ID/状态、首页字段、ID 上限、成功/重试状态与证据；复用总时长/退避引擎及可复用 Wait/WaitForOutput。
- InstanceRunningWaiter 要求 1..50 不重复 ID，固定页 1/大小 50，所有请求 ID 都 Running 才成功；缺 ID 或 Pending/Starting/Stopping/Stopped 重试，重复请求 ID、缺/未知状态、nil 结果、API 错误失败。
- 可覆盖 acceptor，但不能把失败请求/取消/超时变成成功。
- 每次等待/轮询独立复制输入/选项，不可变状态等待器支持并发等待。

- 策略明确读取可重放，Standard 仍需显式启用；不由 CLI operation_type 或 HTTP verb 推断安全。
- AllocateDedicatedHosts 即使有 ClientToken 仍不被自动重试，本阶段保守不提升带 token 写操作的重试权限。
- 缺 token 时在SDK 持有的输入副本上、hooks 前生成一次随机 ASCII token，保留显式 token，执行前及 hooks 后校验非空 ASCII/长度不超过 64；调用者输入不变，各独立调用生成不同 token，刻意重复操作应显式传同一个 token。

- 生成公开强类型 ValidateOperationInput，支持稀疏数值边界、ASCII/字符串长度、数组数量/非空 ID、混合模式排斥。
- nil 表示空请求，不把可选 API 字段变必填；复制输入与 Initialize 后序列化前都校验，诊断仅包含路径/规则，不带值。
- 显式零指针违反审核的正数下限；无约束零/false/空仍准确编码。
- 必填/约束声明仅覆盖已记录规则。

- 敏感模型值/指针的 String/GoString 返回固定类型/已隐藏标记，覆盖普通 fmt 及 %#v， 隐藏整个审核模型，不猜子集。
- JSON/直接字段访问仍明确返回原数据，不增加 logger。
- 标记 STS 请求/响应体/credentials/响应 envelope。
- 命名例外只改 Go 名，首例为原生 DescribeInstanceStatus 的 InstanceId 数组改 InstanceIDs，API 字段名仍准确保留。

- 验收包括确定性、独立策略加载及无效 schema/来源/类型/路径/命名的写前失败、多页 token/页码测试替身、循环/空 token 页、失败/取消/nil 容器/元数据校验/输入选项隔离； 状态等待器全 ID/缺失/重复/状态/超时/错误/自定义 acceptor/并发；签名 token/存在语义/ validator、显式读重试/写不重试、fmt 隐藏；各适配器/策略离线 Example、对应指南/ 包注释。
- 实现后各执行一次前端检查/47 测试、两种 Go 再生成检查、格式/doccheck/vet/ 完整 Go 测试，Linux race/Windows CI；不做真实调用或宣称全策略覆盖。

- 证据于 2026-10-07 阅读，链接与本文件所列的五个官方 API 页面一致，另有固定 DSL/已验收参考策略。
- 网页说明作为独立行为证据，不成为可变生成输入；状态等待器接受条件及保守写重试是本 SDK 的审核策略，不是官方状态等待器声明。

## 参考资料

- [DescribeInstances](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-describeinstances)
- [DescribeInstanceStatus](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-describeinstancestatus)
- [DescribeImages](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-describeimages)
- [DescribeVpcs](https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs)
- [AllocateDedicatedHosts](https://help.aliyun.com/en/ecs/developer-reference/api-ecs-2014-05-26-allocatededicatedhosts)
