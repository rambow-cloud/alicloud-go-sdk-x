# OSS 来源核验

[English](oss-source-verification.md)

- 对应 [#92](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/92)，核对日期为 2026-10-10。
- 核对[暂存覆盖报告](research/oss-semantic-coverage.json)中的 8 个 XML 模型差异。
- DSL 版本为 `ec489e5c3deae95496daae2b41503ac58b221adb`。
- 原生参考为 `alibabacloud-gateway-oss-util` v0.0.6，哈希记录仍见[语义固定文件](../metadata/oss-semantic-pins.json)。
- 已阅读官方文档，并通过 Explorer 的公开元数据接口读取全部 8 个操作，均返回 HTTP 200。JSON 路径和响应哈希见[核验记录](acceptance/oss-source-live.json)。元数据接口证据与浏览器界面证据分别记录。
- 浏览器模型检查和浏览器真实调用仍为 **NOT RUN（未执行）**。网页工具无法读取交互页面，当前没有可用的浏览器控制工具。
- 本地 CLI 使用 aliyun 3.4.11、ossutil 2.4.0 和已有 oss-sftp OAuth Profile，发现北京地域的 1 个桶。全部查询为只读，关闭重试并设置有界超时。
- 修复候选 [#126](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/126) 后，SDK 对 ListBuckets、GetBucketInfo 和 GetBucketReplicationLocation 的原始线路核验**通过**，均只尝试一次，返回 HTTP 200/application/xml。这里使用公共底层运行时，不是生成的 OSS 客户端。
- 修复前，签名器拒绝原生临时 AccessKey ID 中的 `STS.` 句点，尚未发送 HTTP 就失败。核验记录绑定基线提交和修正后的签名器文件哈希；CLI 与 SDK 原始响应结构一致。
- 未修改来源字节、审核批准、模型或覆盖数字。仍为发现 90 个操作、降低 16 个、暂不支持 74 个；OSS Go 输出尚未交付。

## 文档发现与待核验项

| 操作 | 官方文档证据 | 真实核验结果 |
| --- | --- | --- |
| GetBucketCors | [GetBucketCors](https://help.aliyun.com/zh/oss/developer-reference/getbucketcors)说明了请求头条目；[PutBucketCors](https://help.aliyun.com/zh/oss/developer-reference/putbucketcors)的请求语法包含重复的 AllowedHeader 元素。这支持集合模型，但不是实际响应证据。 | **SKIP（跳过）**：HTTP 404 / NoSuchCORSConfiguration，没有可供核验的多请求头规则。 |
| GetBucketInfo | [GetBucketInfo](https://help.aliyun.com/zh/oss/developer-reference/getbucketinfo)使用 `BucketPolicy/LogBucket` 和 `LogPrefix`，固定 DSL 使用 `TargetBucket` 和 `TargetPrefix`。 | **字段名通过核验**：CLI 和 SDK 原始 XML 都有 LogBucket/LogPrefix，值均为空，没有 TargetBucket/TargetPrefix。非空日志配置行为仍未执行。 |
| GetBucketInventory | [GetBucketInventory](https://help.aliyun.com/en/oss/developer-reference/getbucketinventory)将 SSE-OSS 定义为容器，示例是空元素。 | **NOT RUN（未执行）**：没有发现清单配置 ID。空模型和空字符串是否等价需明确序列化决策，不能根据空列表判断。 |
| ListBucketInventory | [ListBucketInventory](https://help.aliyun.com/zh/oss/developer-reference/listbucketinventory)也将 SSE-OSS 定义为容器。 | **SKIP（跳过）**：HTTP 404 / NoSuchInventory，没有 SSE-OSS 清单配置可供核验。 |
| GetBucketReplicationLocation | [GetBucketReplicationLocation](https://help.aliyun.com/zh/oss/developer-reference/getbucketreplicationlocation)示例中每个 TransferTypes 只有一个 Type。 | **重复元素通过核验**：CLI 和 SDK 原始 XML 都有 21 项 TransferTypes，其中 8 项各含一个 Type，13 项各含两个。使用单值模型会丢失返回数据。 |
| GetBucketReplicationProgress | [GetBucketReplicationProgress](https://help.aliyun.com/zh/oss/developer-reference/getbucketreplicationprogress)要求 `rule-id`，示例只有一个 Rule。 | **NOT RUN（未执行）**：配置查询 GetBucketReplication 返回 HTTP 404 / NoSuchReplicationConfiguration，没有编造规则 ID。 |
| GetBucketWebsite | [GetBucketWebsite](https://help.aliyun.com/zh/oss/developer-reference/getbucketwebsite)将 HttpStatus 列为字符串，所查页面未定义 IndexDocument.Type。 | **SKIP（跳过）**：HTTP 404 / NoSuchWebsiteConfiguration。仅凭数字 XML 文本不能确定 Go 整数契约。 |
| ListBuckets | [ListBuckets](https://help.aliyun.com/en/oss/developer-reference/listbuckets)定义 ListAllMyBucketsResult、首字母大写字段和 Buckets/Bucket 包装层。 | **包装层和大小写通过核验**：CLI 和 SDK 原始 XML 一致，仅返回一个桶，没有分页字段。真实续页仍未执行；服务级地址和请求头降低能力仍需实现。 |

## Explorer 元数据核对结果

- 使用 `https://api.aliyun.com/meta/v1/products/Oss/versions/2019-05-17/apis/<Operation>/api.json?language=EN_US` 读取定义。完整 URL、哈希和 JSON 路径见核验记录，该公开接口无需浏览器登录。
- GetBucketCors：AllowedHeader 声明 `type: string`，同时有 `items.type: string` 和 `items.extendType: "true"`。这是不一致或扩展表示，不是标准数组声明；批准数组修正前需规范化并审核。
- GetBucketInfo：BucketPolicy 的内联字段为 LogBucket/LogPrefix，但 `$ref` 引用的 LoggingEnabled 字段为 TargetBucket/TargetPrefix/LoggingRole。两边事实都保留；本次原始响应和帮助文档支持该响应中的内联名称。
- GetBucketReplicationLocation：TransferTypes.Type 在元数据中是字符串数组，原始 XML 证实元素可以重复。
- GetBucketReplicationProgress：元数据中 Rule 是数组，没有已有复制规则可供真实核验。
- 两个清单操作引用的 SSEOSS 声明为 `type: string`，同时有空 `properties`；帮助文档将其描述为空容器，真实映射仍未验证。
- GetBucketWebsite：元数据将 HttpStatus 和 Type 声明为整数，DSL 声明为字符串，帮助文档也将 HttpStatus 列为字符串。这项三方差异仍保留。
- ListBuckets：元数据与实际响应的首字母大写节点及 Buckets/Bucket 包装层一致，但不能据此宣称全部桶字段和分页续页均已覆盖。
- 元数据可用于佐证或发现冲突，不替代完整官方 DSL/parser 路线，也不授权静默修改模型。

## Explorer 浏览器核验步骤

- 在已登录的浏览器中打开 [OpenAPI Explorer](https://api.aliyun.com/)，搜索 OSS 和操作名。
- 下表链接是待核验入口，**尚未确认路由有效**。链接不可用或产品目录没有该操作时，如实记录；不要用名称相似的控制面 API 替代 OSS 数据面操作。
- 使用已有账号、桶和桶所在地域，只执行下列 GET 读取操作。
- 清单和复制操作需要已有配置 ID；CORS、日志和网站操作需要已有配置。不要为了获取证据创建或修改配置。
- 先查看 API 模型，再在支持在线调试时查看响应。区分界面展示的是原始 XML、解码后的 JSON，还是模型或示例。
- 403、配置不存在或空结果不能验证有争议的字段结构。
- 记录操作和版本、UTC 请求时间、地域、HTTP 状态、响应格式及争议 XML 路径，保留脱敏的结构片段。不提交凭据、授权头、token 或真实账号和资源标识符。
- 浏览器模型证据、浏览器真实调用证据和本地 SDK/CLI 证据分别记录。看不到原始 XML 时，原始线路核验仍标记为 NOT RUN。

| 操作 | 待核验 Explorer 页面 | 所需已有资源 |
| --- | --- | --- |
| GetBucketCors | [打开](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketCors) | 包含多请求头 CORS 规则的桶 |
| GetBucketInfo | [打开](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketInfo) | 已开启日志的桶 |
| GetBucketInventory | [打开](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketInventory) | 桶和 SSE-OSS 清单配置 ID |
| ListBucketInventory | [打开](https://api.aliyun.com/api/Oss/2019-05-17/ListBucketInventory) | 包含 SSE-OSS 清单配置的桶 |
| GetBucketReplicationLocation | [打开](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketReplicationLocation) | 桶 |
| GetBucketReplicationProgress | [打开](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketReplicationProgress) | 桶和复制规则 ID |
| GetBucketWebsite | [打开](https://api.aliyun.com/api/Oss/2019-05-17/GetBucketWebsite) | 已有静态网站配置的桶 |
| ListBuckets | [打开](https://api.aliyun.com/api/Oss/2019-05-17/ListBuckets) | 名下已有桶的账号 |

## 修正前的门禁

- 本文提供候选修正依据，不批准运行时行为，也不直接解除生成器阻塞。
- 每项修正都需绑定来源版本、记录中英文决策并提供独立的离线 XML 测试，再重新生成。
- 保留完整 DSL 和上游原始字节。模型例外与序列化映射必须明确表达，并可复用。
- 其余 66 个操作当前受方法或正文格式、动态路径、正文模型缺失和服务级绑定限制；Explorer 证据不能代替这些生成器能力的实现。

## 验证要求

- 本分支仅增加文档和脱敏证据，执行文档语言与本地链接检查及 `git diff --check`。
- 签名器代码验证归属 #126；原始私有响应不纳入版本控制。没有为了补齐证据创建或修改云配置。
