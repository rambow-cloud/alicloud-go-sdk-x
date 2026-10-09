# 官方端点规则

[English](endpoint-rules.md)

- 对应 issue #91；端点数据由官方语义 parser 从固定 Darabonba 源码提取。
- 产品 IR 和锁文件使用 schema v4；旧版本在写入前拒绝。上游源码与来源清单保持原样。
- 生成器管理公共目录 `endpoint/rules.gen.go`、Apache LICENSE 和 NOTICE；原创解析器与工具仍按 MIT 许可。

## 选择顺序

- `Config.Network` 随配置复制到服务和操作 `Options`；空值或 `public` 使用公网规则。自定义解析器同样会收到网络选项。
- `BaseEndpoint` 优先级最高，可以覆盖未知产品、地域或网络，但必须是合法的 HTTPS 源地址。
- 公网先查官方 `endpointMap` 准确映射，再使用源码的区域化或集中式构造规则。
- 例如，杭州 ECS 使用 `ecs-cn-hangzhou.aliyuncs.com`，杭州 VPC 使用 `vpc.aliyuncs.com`，ap-south-1 的 STS 使用 `sts.aliyuncs.com`。
- 区域规则接受单个由小写字母、数字和连字符组成的合法 DNS 标签。构造出地址不能证明该地域已部署服务或可访问。
- 非法地域标签和未知产品返回 `endpoint.ErrUnsupported`；解析不发 HTTP 请求，也不查询 DNS。
- `NewRules` 仍只解析准确规则，并复制输入；空网络与 `public` 使用同一个键。零值仅支持显式地址覆盖。

## 私网访问

- `Network: "vpc"` 仅选择已经审核的准确私网地址，不退回公网。
- 当前审核 ECS 11 个、STS 35 个、VPC 1 个（北京）私网规则；策略文件列出地域及依据 URL。
- VPC 账号侧仍可能需要 PrivateLink 配置；解析出地址不等于自动建立网络连接。
- 未列出的私网、后缀或地域与网络组合，须通过 `BaseEndpoint` 或自定义规则、解析器显式提供。
- 有意调整的行为：已审核私网规则优先于 DSL 公网映射，避免原 DSL 在设置网络选项后仍返回公网地址。

## 生成约束

- 前端要求唯一且为常量的规则和映射，校验准确调用参数，以及「显式地址 → 映射 → 规则」控制流。
- 控制流变化、非常量表达式、重复映射或不支持的规则会阻止投影。
- Go 生成在写入前校验来源坐标、地址、私网规则类型、重复项及依据。
- 端点策略绑定来源版本；不代表操作已审核，也不授予重试或分页能力。
- 目录由全部产品 IR 确定性生成，与产品产物一起完成写入前检查。
- 测试覆盖特殊映射、区域规则、私网与公网隔离、覆盖优先级、数据归属、非法来源和策略拒绝；反例验证不会写入部分服务或端点文件。
- 地址选择属于离线证据，未宣称私网真实连通性。

## 官方依据

- 固定的 `sources/darabonba/products/{ecs,sts,vpc}/main.tea`；初始化及 getEndpoint 来源坐标保存在 IR 和生成注释中。
- [官方构造规则](https://www.alibabacloud.com/help/en/sdk/developer-reference/endpoint-configuration)。
- [ECS 私网端点](https://www.alibabacloud.com/help/en/ecs/developer-reference/call-api-operations-over-the-internal-network)。
- [STS 端点](https://www.alibabacloud.com/help/en/ram/developer-reference/api-sts-2015-04-01-endpoint)。
- [VPC PrivateLink](https://www.alibabacloud.com/help/en/vpc/use-privatelink-to-access-vpc-openapi-over-private-network)。
