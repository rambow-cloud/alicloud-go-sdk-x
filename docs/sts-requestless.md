# Requestless STS generation / 无请求模型 STS 生成

## English

Issue #58 lowers the pinned official GetCallerIdentity function with only a runtime
parameter. Its IR request root is `{"kind":"empty"}` with no bindings or fabricated
DSL model. The Go backend emits `GetCallerIdentityInput struct{}` to preserve
`client.GetCallerIdentity(ctx, input, optFns...)` and the small GetCallerIdentityAPI.
Nil and zero-value input both send no query members. The complete native response
and Metadata remain typed; pointers distinguish absent and explicit empty fields.

This is a signed RPC action: explicitly register a source provider, including when
the account needs no RAM permission for this operation. Construction never reads
credentials. Context cancellation/deadlines and structured service errors survive;
default error text omits raw service messages. JSON v2 rejects duplicate names and
tolerates unknown response members. Requestless functions must use the exact reviewed
runtime signature, empty OpenApiRequest, constant protocol and callApi handoff;
extra behavior/fields remain unsupported before writes.

Generate through the official frontend/discovery and sdkgen product-generate, then
check both generation paths, doccheck, vet, tests/Examples, Linux race and Windows CI.
The generated product guide/coverage and ExampleClient_GetCallerIdentity document
the operation; independent offline wire/output/error/context tests live under
service/sts. Empty Go inputs do not inflate official reachable model counts.
Real identity evidence is accepted separately under #60; no live call is made here.

The same source pattern also automatically admits pinned VPC
ListGeographicSubRegions. Its generated complete types/API/Example and coverage
are additive; existing operations and retry policies are unchanged. It stays
unreviewed for optional capabilities and does not expand v0.1.0 STS acceptance.

## 中文

#58 降低固定官方 GetCallerIdentity 的仅 runtime 参数函数，IR 请求根为
`{"kind":"empty"}`，没有绑定、不伪造 DSL 模型。Go 后端输出空
`GetCallerIdentityInput struct{}`，保留统一 context/input/options 调用及窄 mock 接口。
nil 和零值不发送 query 成员，完整原生响应及 Metadata 强类型保留；指针区分缺失与显式空值。

该操作采用签名 RPC：即使无需 RAM 操作权限，也须显式注册来源 provider；构造不读凭据。
保留取消/超时与结构化服务错误，默认错误不显示原始服务说明。JSON v2 拒绝重复名称、
容忍新增响应字段。无请求函数只支持审核的 runtime 签名、空 OpenApiRequest、常量协议及
callApi handoff，额外字段/行为继续写前拒绝。

经官方前端/发现及 sdkgen product-generate 生成，再执行两套生成检查、doccheck、vet、
测试/Examples 和 Linux race/Windows CI。生成指南/覆盖与操作 Example 同步，独立离线
线路/完整响应/错误/取消契约在 service/sts；空 Go 输入不增加官方模型数。真实身份证据
由 #60 单独验收，本项不调用云。

同一官方模式还自动支持固定 VPC ListGeographicSubRegions，追加生成完整类型/API/
Example 及覆盖，不改变已有操作/重试策略。其可选能力仍为未审核，不扩张 v0.1.0 STS 验收。
