# 固定协议元数据

[English](README.md)

- 生产 manifest 同时固定官方语义解析器的 dsl.json 投影、来源锁及审核偏差决策。
- 只刷新元数据后也需重投影 DSL 并审核差异，再生成 Go；见 [Darabonba 工具](../tools/darabonba/README.zh-CN.md)。

- 每个产品目录包含 manifest.json、审核 overlay.json 和仅协议操作快照。
- manifest 记录官方 URL、 获取时间、原始源与快照 SHA-256。
- 生成离线验证快照 hash；原始 hash 是来源记录，不保证当前远端响应相同。

- 快照来自阿里云公开元数据 API 的协议结构事实，不包含上游说明和示例，不表示上游文档许可证。
- 原创补充配置和生成文档遵循仓库 MIT LICENSE；[设计与支持范围](../docs/generator.zh-CN.md)。

- 显式刷新 ECS 的联网命令见本页，会替换固定源文件。
- STS 使用 Sts、2015-04-01、sts、AssumeRole。
- VPC 使用 Vpc、2016-04-28、vpc、DescribeVpcs；补充配置选择指针布尔过滤、Tag repeatList 对象、 嵌套响应模型和纯页码分页器，见 [VPC 指南](../docs/vpc.zh-CN.md)。
- 非空 components.schemas 定义保留用于审核本地引用，不包含说明文字。
- `-raw-dir PATH` 离线提取已下载的 operation-name.json，并以文件修改时间记录获取时间。
- 先在 issue 下评审元数据及补充配置差异再生成。
- 产品/style 经过人工审核：单操作 API 不含产品 Info/style， 导入仅支持 RPC。
- 不导入凭据、账号响应、上游说明文字或推测端点规则。

## 可运行命令与示例

```sh
go run ./internal/cmd/sdkgen import -product Ecs -version 2014-05-26 -package ecs -operations DescribeRegions,DescribeInstances,DescribeInstanceStatus -out metadata/ecs
```
