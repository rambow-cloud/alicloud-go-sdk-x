# SDK 调研

[English](research.md)

- V1 仓库：[aliyun/alibaba-cloud-sdk-go](https://github.com/aliyun/alibaba-cloud-sdk-go)。

- 调研日期 2026-10-07。
- 报告针对上游具体版本，未在本地复现，也不表示所有最新阿里云 SDK 有同样问题。
- V1 于 2025-03-01 停止支持、2026-02-11 归档，来源为本文件所列的官方仓库；V2 产品位于 alibabacloud-go。

| 证据                             | 设计应对               |
| -------------------------------- | ---------------------- |
| V2 #72：Config.HttpClient 报告   | 可注入 net/http        |
| OpenAPI #177：Tea 依赖编译报告   | 标准库核心             |
| V2 #68：header 共享状态修改报告  | 私有、构造后不变的配置 |
| V1 #612：endpoint 端点解析器需求 | 公共端点解析器契约     |
| V1 #669：NextToken 覆盖报告      | 统一引擎、产品规则     |
| OpenAPI #221：swag 注释解析报告  | 原生 Go 注释和文档门禁 |
| OpenAPI #225：依赖许可证报告     | 许可证与来源审查       |

- 以下记录来自上游报告；本项目没有把这些报告当作本地复现结果。

| 上游证据                                                                                   | 本项目的设计选择                     |
| ------------------------------------------------------------------------------------------ | ------------------------------------ |
| [V2 #72](https://github.com/aliyun/alibabacloud-go-sdk/issues/72)：HTTP 客户端配置问题     | 可注入标准 net/http 客户端           |
| [OpenAPI #177](https://github.com/aliyun/darabonba-openapi/issues/177)：Tea 依赖编译问题   | 核心只依赖标准库                     |
| [V2 #68](https://github.com/aliyun/alibabacloud-go-sdk/issues/68)：请求头状态被修改        | 构造后保持配置私有且不可变           |
| [V1 #612](https://github.com/aliyun/alibaba-cloud-sdk-go/issues/612)：端点解析扩展需求     | 提供公开端点解析接口                 |
| [V1 #669](https://github.com/aliyun/alibaba-cloud-sdk-go/issues/669)：NextToken 支持不完整 | 统一分页引擎，保留各产品原生分页规则 |
| [OpenAPI #221](https://github.com/aliyun/darabonba-openapi/issues/221)：注释解析问题       | 原生 Go 注释与文档检查               |
| [OpenAPI #225](https://github.com/aliyun/darabonba-openapi/issues/225)：依赖许可证问题     | 保留来源与许可证证据                 |

- 所查 ECS 生成代码体现较多指针配置和未显式带 ctx 的方法，但不能推断整个 V2 缺少 context。
- OSS V2 已有良好的 ctx-first 模式。
- 不把旧迁移文档推广至所有产品，也不以手写代码量判断成熟度。
- 借鉴 AWS v2 的中间件、小测试接口和凭据提供者与缓存契约，同时保留阿里云协议特性。
- 生成前完成公共运行时；未进行固定输入的可复现基准前不表示性能或体积改善。

## 参考资料

- [V2 #72](https://github.com/aliyun/alibabacloud-go-sdk/issues/72)
- [OpenAPI #177](https://github.com/aliyun/darabonba-openapi/issues/177)
- [V2 #68](https://github.com/aliyun/alibabacloud-go-sdk/issues/68)
- [V1 #612](https://github.com/aliyun/alibaba-cloud-sdk-go/issues/612)
- [V1 #669](https://github.com/aliyun/alibaba-cloud-sdk-go/issues/669)
- [OpenAPI #221](https://github.com/aliyun/darabonba-openapi/issues/221)
- [OpenAPI #225](https://github.com/aliyun/darabonba-openapi/issues/225)

## 来源与证据链接

- [参考链接](https://github.com/alibabacloud-go/ecs-20140526/blob/master/client/client.go)
- [参考链接](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/DEVGUIDE-CN.md)
