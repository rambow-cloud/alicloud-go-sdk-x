# Pinned protocol metadata / 固定协议元数据

## English

Production manifests also pin the official parser's dsl.json projection, source lock
and reviewed divergence decisions. Refreshing metadata alone requires re-projecting
DSL and reviewing differences before Go generation; see [Darabonba tools](../tools/darabonba/README.md).

Each product directory contains manifest.json, a reviewed overlay.json and protocol-only
operation snapshots. Official source URLs, retrieval times, raw-source SHA-256 and
snapshot SHA-256 are in the manifest. Generation verifies snapshot hashes offline;
the raw hash is provenance, not a promise that today's remote response is identical.

Snapshots derive structural protocol facts from Alibaba Cloud's public metadata API.
Upstream descriptions/examples are excluded and no upstream documentation license is
asserted. Original overlays and generated documentation follow the repository MIT
LICENSE. See [the design and supported profile](../docs/generator.md).

To refresh ECS explicitly (network access, replaces pinned source files):

```sh
go run ./internal/cmd/sdkgen import -product Ecs -version 2014-05-26 -package ecs -operations DescribeRegions,DescribeInstances,DescribeInstanceStatus -out metadata/ecs
```

For STS use product Sts, version 2015-04-01, package sts and operation AssumeRole.
For VPC use product Vpc, version 2016-04-28, package vpc and operation DescribeVpcs.
The VPC overlay selects pointer boolean filters, Tag repeatList objects, nested response
models and a page-only paginator; see [the VPC guide](../docs/vpc.md). Nonempty
components.schemas definitions are retained for reviewed local references, without prose.
`-raw-dir PATH` extracts already downloaded operation-name.json files offline and
records their modification time as retrieval time. Review the metadata and overlay
diff under an issue before regenerating. Product/style selection is reviewed: the
operation API does not itself carry product Info/style; import supports RPC only.
Do not import credentials, account responses, upstream prose or guessed endpoint rules.

## 中文

生产 manifest 同时固定官方 parser 的 dsl.json 投影、来源锁及审核偏差决策。只刷新
元数据后也需重投影 DSL 并审核差异，再生成 Go；见 [Darabonba 工具](../tools/darabonba/README.md)。

每个产品目录包含 manifest.json、审核 overlay.json 和仅协议操作快照。manifest 记录官方 URL、
获取时间、原始源与快照 SHA-256。生成离线验证快照 hash；原始 hash 是来源记录，不保证当前远端响应相同。

快照来自阿里云公开元数据 API 的协议结构事实，不包含上游说明和示例，不宣称上游文档许可证。
原创 overlay 和生成文档遵循仓库 MIT LICENSE；[设计与支持范围](../docs/generator.md)。

显式刷新 ECS 的联网命令见英文章节，会替换固定源文件。STS 使用 Sts、2015-04-01、sts、AssumeRole。
VPC 使用 Vpc、2016-04-28、vpc、DescribeVpcs；overlay 选择指针布尔过滤、Tag repeatList 对象、
嵌套响应模型和纯页码 paginator，见 [VPC 指南](../docs/vpc.md)。非空 components.schemas 定义
保留用于审核本地引用，不包含说明文字。
`-raw-dir PATH` 离线提取已下载的 operation-name.json，并以文件修改时间记录获取时间。
先在 issue 下评审元数据及 overlay 差异再生成。产品/style 经过人工审核：单操作 API 不含产品 Info/style，
导入仅支持 RPC。不导入凭据、账号响应、上游说明文字或推测端点规则。
