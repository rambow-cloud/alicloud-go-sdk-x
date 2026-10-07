# 现有阿里云 Go SDK 调研

调研日期：2026-10-07。公开 issue 代表特定版本的用户报告，不等于已复现，
也不意味着当前版本仍有同一缺陷。以下分别记录观察、报告和本项目的设计判断。
源码链接指向上游默认分支，后续可能变化；涉及实现前须在对应 issue 固定版本和协议来源。

## 生命周期与范围

[V1 仓库](https://github.com/aliyun/alibaba-cloud-sdk-go)声明于 2025-03-01
停止支持，GitHub 显示仓库于 2026-02-11 归档。新项目不能建立在其维护承诺之上。
V2 产品模块见 [alibabacloud-go](https://github.com/alibabacloud-go)，
并非所有阿里云 Go SDK 都有相同设计。

## 证据与改进方向

| 观察或用户报告 | 证据及边界 | 本项目决策 |
| --- | --- | --- |
| 通用产品 API 的 Go 使用成本 | [ECS client.go](https://github.com/alibabacloud-go/ecs-20140526/blob/master/client/client.go)中可观察到请求/运行时 options 风格、指针配置和生成代码；所查方法没有 ctx 参数。不能推断整个 V2 没有 context 支持 | 新的阻塞 API 统一 ctx-first；类型模型尽量采用值，保留有意义的 optional |
| HTTP 客户端配置难以接入 | [V2 #72](https://github.com/aliyun/alibabacloud-go-sdk/issues/72)，2024-08-05，报告 Config.HttpClient 未定义；未在本地复现 | 标准 net/http 客户端可注入，配置与观测扩展有明确契约 |
| 依赖组合编译失败 | [OpenAPI #177](https://github.com/aliyun/darabonba-openapi/issues/177)，2025-02-17，报告特定版本找不到 tea.HttpClient | 核心零第三方运行时依赖；生成器与运行时分离；CI 验证支持的 Go 版本 |
| 共享状态的语义令人困惑 | [V2 #68](https://github.com/aliyun/alibabacloud-go-sdk/issues/68)，2024-03-29，报告 GetRpcHeaders 读取后清空字段 | 构造后私有配置，不修改调用者参数；未来客户端并发性由 race 测试验证 |
| Endpoint 扩展需求 | [V1 #612](https://github.com/aliyun/alibaba-cloud-sdk-go/issues/612)，2023-12-04，用户希望通过统一 resolver 接入内部域名 | 单独设计 endpoint resolver、显式覆盖和 HTTPS 校验；不猜测特殊地域规则 |
| 分页模型追不上服务 API | [V1 #669](https://github.com/aliyun/alibaba-cloud-sdk-go/issues/669)，2025-03-25，报告部分接口缺少 NextToken | 产品模型按版本维护，分页器检测重复 token，并支持 ctx 取消 |
| 生成注释影响工具链 | [OpenAPI #221](https://github.com/aliyun/darabonba-openapi/issues/221)，2025-07-04，描述 swag 解析 @param 失败；这不是 pkg.go.dev 故障报告 | 采用 Go 原生 doc comments，包文档、导出符号、可执行示例作为 CI 门禁 |
| 依赖许可证缺失报告 | [OpenAPI #225](https://github.com/aliyun/darabonba-openapi/issues/225)，2025-07-22，针对 gateway-pop 依赖的许可证；不作当前法律状态判断 | 根目录标准 LICENSE，新增依赖审查许可证，发布前确认模块归属 |

不能把旧迁移文档中的 timeout、retry、concurrency 描述当作所有最新 SDK 的行为。
[OSS Go SDK V2 指南](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/DEVGUIDE-CN.md)
已有 ctx-first API 和重试设计，是可以参考的正面样例。因此这里重点重建通用 OpenAPI
开发体验，不计划首版重做 OSS 对象存储的数据面 SDK。

## 优先级

先初始化开发流程、文档门禁和凭据/错误契约。接着以签名协议固定测试向量完成请求管线，
再用 ECS 一个只读 API 验证端到端设计。随后加入安全重试、临时凭据刷新、分页和生成器。
初始化不提供性能或体积数字；依赖图、编译时间和二进制大小的比较留给可复现的基准 issue。

文档基础规则来自 [Go doc comments](https://go.dev/doc/comment)、
[pkg.go.dev 索引说明](https://pkg.go.dev/about#adding-a-package)和
[许可证检测规则](https://pkg.go.dev/license-policy)。许可证无法识别时，pkg.go.dev
只展示有限模块信息；README 不能代替 Go 包及符号文档。
