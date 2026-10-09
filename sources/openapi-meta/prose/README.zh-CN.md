# 可选字段说明来源

[English](README.md)

- 对应 #93，见[设计与验收](../../../docs/canonical-prose-enrichment.zh-CN.md)。
- 固定 `aliyun/aliyun-openapi-meta` 提交 `51286a65c79d008436eb314e636f9c9ad4b1ca08`，Apache-2.0；保留原始来源与 LICENSE 字节。
- 自动对应全部 787 个 STS、ECS、VPC DSL 操作，取得 778 份元数据；固定版本缺少 9 个操作，已记入 manifest.json。
- 只补充可选的信息性说明，不影响完整 DSL 发现、原生模型或运行策略。
- 上游结构不稳定，面向 CLI 构建；匹配前规范化索引输入和 itemName 包装，明确报告差异，不静默采用。
- 父目录原有规范化夹具保持原样。
- `node tools/darabonba/prose.cjs generate` 输出确定性的本地说明投影；`check` 在不访问网络的情况下核对产物。
- 保留来源哈希、字段坐标、JSON 指针及排除原因，不把上游账号或资源示例当作可执行 SDK 示例。
- 普通生成不更新来源提交；导入和来源修改须先审核 issue。
