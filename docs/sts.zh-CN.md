# STS 使用指南

[English](sts.md)

- 导入 `service/sts`；#81 已移除旧 `services/sts` 包。
- 官方 Darabonba/parser → 完整 IR → 公共 Go 输出器生成 AssumeRole、GetCallerIdentity、AssumeRoleWithOIDC 和 AssumeRoleWithSAML。
- 使用 `sts.NewFromConfig`，将生成的 AssumeRoleInput 传给 `stscreds.NewAssumeRoleProvider`；NewAssumeRoleProviderFromClient 转发到同一实现。
- 可复用辅助组件复制指针及选项，验证审核过的角色、会话和 policy 规则，解析 RFC3339 过期时间，并拒绝不完整或已过期的凭据。
- 来源 provider 与角色 provider 保持分离；使用 credentials.Cache 管理同步、有界的续期。
- 签名操作需要显式配置或加载器选出的 provider；审核过的匿名操作不读取签名凭据，客户端构造时不发起凭据 HTTP 请求。
- 可以识别取消及结构化服务错误；Standard 不重试 AssumeRole 凭据发放。
- 详见[生成指南](products/sts.zh-CN.md)、[角色组合](sts-credentials.zh-CN.md)、[消费者验收](sts-consumer-acceptance.zh-CN.md)和[迁移说明](service-consolidation.zh-CN.md)。成功的真实联邦认证仍不在当前已验收范围内。
