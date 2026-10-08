# 无请求模型 STS 生成

[English](sts-requestless.md)

- #58 将固定的官方 GetCallerIdentity 函数（只有运行时参数）转换为 IR。请求根为 `{"kind":"empty"}`，没有绑定、不伪造 DSL 模型。
- Go 后端输出空 `GetCallerIdentityInput struct{}`，保留统一 context/input/options 调用及窄测试替身接口。
- nil 和零值不发送 query 成员，完整原生响应及 Metadata 强类型保留；指针区分缺失与显式空值。

- 该操作采用签名 RPC：即使无需 RAM 操作权限，也须显式注册来源凭据提供者；构造不读凭据。
- 保留取消/超时与结构化服务错误，默认错误不显示原始服务说明。
- JSON v2 拒绝重复名称、 容忍新增响应字段。
- 无请求函数只支持审核的运行时签名、空 OpenApiRequest、常量协议及 callApi handoff，额外字段/行为继续写前拒绝。

- 经官方前端/发现及 sdkgen product-generate 生成，再执行两套生成检查、doccheck、vet、 测试/Examples 和 Linux race/Windows CI。
- 生成指南/覆盖与操作 Example 同步，独立离线线路/完整响应/错误/取消契约在 service/sts；空 Go 输入不增加官方模型数。
- 真实身份证据由 #60 单独验收，本项不调用云。

- 同一官方模式还自动支持固定 VPC ListGeographicSubRegions，追加生成完整类型/API/ Example 及覆盖，不改变已有操作/重试策略。
- 其可选能力仍为未审核，不扩张 v0.1.0 STS 验收。
