# STS 角色扮演

[English](sts.md)

- 完整 service/sts 客户端使用 NewAssumeRoleProviderFromClient，见 [完整 DSL 凭据指南](sts-credentials.zh-CN.md)。
- 下述构造器保留原 services/sts 参考桥契约。

- `sts.New` 用独立来源凭据构造强类型 AssumeRole 客户端。
- `sts.AssumeRoleAPI` 支持小型 fake。
- 输入验证会话语法、最小时长、external/来源身份和 policy JSON；授权及角色最大有效期由服务决定。
- 可选 ExternalID 支持混淆代理保护。
- 响应把 Expiration 解析为 RFC3339 时间，格式化隐藏凭据。
- `stscreds.NewAssumeRoleProvider` 复制输入/options，验证非空密钥/token 和未来过期时间。
- 使用 `credentials.NewCache` 包装以同步提前刷新。
- STS 来源凭据提供者必须与扮演角色凭据提供者分离，避免递归读取；辅助组件不修改来源。
- Context 失败可识别。
- AssumeRole 保守视为非幂等，Standard 不重试发放 token。
- 只覆盖此 STS 操作，不表示真实账号验证。
- 协议：[官方 AssumeRole 文档](https://help.aliyun.com/zh/ram/developer-reference/api-sts-2015-04-01-assumerole)。

- 操作/模型/接口由生成器生成，仅说明中的校验规则保留手写。
- [生成字段指南](generated/sts.md)。
