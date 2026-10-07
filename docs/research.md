# SDK research / SDK 调研

[English](#english) | [中文](#中文)

## English

Research date: 2026-10-07. Reports concern specific upstream versions and are not local
reproductions or claims that every latest Alibaba SDK has the same issue.
V1 ended support on 2025-03-01 and was archived on 2026-02-11:
https://github.com/aliyun/alibaba-cloud-sdk-go. V2 products are under alibabacloud-go.

| Evidence | Design response |
| --- | --- |
| [V2 #72](https://github.com/aliyun/alibabacloud-go-sdk/issues/72): Config.HttpClient report | injectable net/http |
| [OpenAPI #177](https://github.com/aliyun/darabonba-openapi/issues/177): tea dependency compile report | standard-library core |
| [V2 #68](https://github.com/aliyun/alibabacloud-go-sdk/issues/68): header state mutation report | private immutable configuration |
| [V1 #612](https://github.com/aliyun/alibaba-cloud-sdk-go/issues/612): endpoint resolver request | public resolver contract |
| [V1 #669](https://github.com/aliyun/alibaba-cloud-sdk-go/issues/669): NextToken coverage report | unified engine, product policies |
| [OpenAPI #221](https://github.com/aliyun/darabonba-openapi/issues/221): swag comment parser report | native Go comments and doc checks |
| [OpenAPI #225](https://github.com/aliyun/darabonba-openapi/issues/225): dependency license report | license/provenance review |

ECS generated code illustrates pointer-heavy configuration and methods without explicit ctx
in the inspected examples: https://github.com/alibabacloud-go/ecs-20140526/blob/master/client/client.go.
This does not prove all V2 lacks context. OSS V2 provides good ctx-first patterns:
https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/DEVGUIDE-CN.md.
Do not generalize old migration docs to every current product or infer maturity from handwritten line count.

Use AWS v2's middleware, small testing interfaces and provider/cache contracts as design
references, preserving Alibaba-specific protocols. Complete the shared runtime before generation.
Do not claim performance/size improvements without pinned, reproducible benchmarks.

## 中文

调研日期 2026-10-07。报告针对上游具体版本，未在本地复现，也不宣称所有最新阿里云 SDK 有同样问题。
V1 于 2025-03-01 停止支持、2026-02-11 归档，来源为英文章节的官方仓库；V2 产品位于 alibabacloud-go。

| 证据 | 设计应对 |
| --- | --- |
| V2 #72：Config.HttpClient 报告 | 可注入 net/http |
| OpenAPI #177：Tea 依赖编译报告 | 标准库核心 |
| V2 #68：header 共享状态修改报告 | 私有、构造后不变的配置 |
| V1 #612：endpoint resolver 需求 | 公共 resolver 契约 |
| V1 #669：NextToken 覆盖报告 | 统一引擎、产品规则 |
| OpenAPI #221：swag 注释解析报告 | 原生 Go 注释和文档门禁 |
| OpenAPI #225：依赖许可证报告 | 许可证与来源审查 |

所有证据链接与英文表格对应。所查 ECS 生成代码体现较多指针配置和未显式带 ctx 的方法，但不能推断整个 V2
缺少 context。OSS V2 已有良好的 ctx-first 模式。不把旧迁移文档推广至所有产品，也不以手写代码量判断成熟度。
借鉴 AWS v2 的 middleware、小测试接口和 provider/cache 契约，同时保留阿里云协议特性。
生成前完成共享 runtime；未进行固定输入的可复现基准前不宣称性能或体积改善。
