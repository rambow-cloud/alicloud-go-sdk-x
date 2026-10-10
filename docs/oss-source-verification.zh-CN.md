# OSS 来源核验

[English](oss-source-verification.md)

- 对应 [#92](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/92)，核对日期为 2026-10-10。
- 核对[暂存覆盖报告](research/oss-semantic-coverage.json)中的 8 个 XML 模型差异。
- DSL 版本为 `ec489e5c3deae95496daae2b41503ac58b221adb`。
- 原生参考为 `alibabacloud-gateway-oss-util` v0.0.6，哈希记录仍见[语义固定文件](../metadata/oss-semantic-pins.json)。
- 已通过网页工具阅读官方文档。这些是当前文档，不是固定版本的真实响应，也不是 Explorer 核验结果。
- 网页工具无法读取 GetBucketCors 和 ListBuckets 的 Explorer 页面；当前也没有可用的浏览器控制工具。这不能证明 Explorer 不提供这些 API。
- 8 个操作的 Explorer 页面模型检查均为 **NOT RUN（未执行）**。
- 8 个操作的 Explorer 真实请求及 SDK 真实请求均为 **NOT RUN（未执行）**。
- 未修改来源字节、审核批准、模型或覆盖数字。仍为发现 90 个操作、降低 16 个、暂不支持 74 个；OSS Go 输出尚未交付。

## 文档发现与待核验项

| 操作 | 官方文档证据 | 仍需核验 |
| --- | --- | --- |
| GetBucketCors | [GetBucketCors](https://help.aliyun.com/zh/oss/developer-reference/getbucketcors)说明了请求头条目；[PutBucketCors](https://help.aliyun.com/zh/oss/developer-reference/putbucketcors)的请求语法包含重复的 AllowedHeader 元素。这支持集合模型，但不是实际响应证据。 | 读取已有且至少包含两个允许请求头的规则，检查 `CORSConfiguration/CORSRule/AllowedHeader` 是否重复。单个条目不能证明最多只返回一个。 |
| GetBucketInfo | [GetBucketInfo](https://help.aliyun.com/zh/oss/developer-reference/getbucketinfo)定义及示例使用 `BucketPolicy/LogBucket` 和 `LogPrefix`，固定 DSL 使用 `TargetBucket` 和 `TargetPrefix`。 | 读取已开启日志的桶，核对 `BucketInfo/Bucket/BucketPolicy` 下的准确元素名。不能再简单解释为旧 helper 缺少新版字段。 |
| GetBucketInventory | [GetBucketInventory](https://help.aliyun.com/en/oss/developer-reference/getbucketinventory)将 SSE-OSS 定义为容器，示例是空元素。 | 读取已有且采用 SSE-OSS 加密的清单配置，区分空元素与元素缺失。空模型和空字符串可能对应同一种 XML，需明确映射，不能直接改成文本字段。 |
| ListBucketInventory | [ListBucketInventory](https://help.aliyun.com/zh/oss/developer-reference/listbucketinventory)也将 SSE-OSS 定义为容器，但列表示例没有展示该元素。 | 读取包含已有 SSE-OSS 清单配置的列表，检查 `InventoryConfiguration[]/Destination/OSSBucketDestination/Encryption/SSE-OSS`。 |
| GetBucketReplicationLocation | [GetBucketReplicationLocation](https://help.aliyun.com/zh/oss/developer-reference/getbucketreplicationlocation)示例有多个 LocationTransferType，但每个 TransferTypes 中只有一个 Type。 | 检查 `LocationTransferType[]/TransferTypes/Type`。现有示例不能确定 Type 是单值还是可重复元素。 |
| GetBucketReplicationProgress | [GetBucketReplicationProgress](https://help.aliyun.com/zh/oss/developer-reference/getbucketreplicationprogress)要求 `rule-id`，示例只有一个 Rule；说明中的父节点名称也存在不一致。 | 读取已有复制规则。单个 Rule 既可能对应单模型，也可能对应只有一项的数组，不能据此确定最大数量。 |
| GetBucketWebsite | [GetBucketWebsite](https://help.aliyun.com/zh/oss/developer-reference/getbucketwebsite)明确将 ErrorDocument.HttpStatus 列为字符串；本次阅读的页面未定义 IndexDocument.Type。 | 读取已有静态网站配置，分别核对两个字段。XML 文本是数字不能证明 API 类型是整数。HttpStatus 的 DSL 字符串类型有文档依据，Type 仍未确定。 |
| ListBuckets | [ListBuckets](https://help.aliyun.com/en/oss/developer-reference/listbuckets)定义根节点 ListAllMyBucketsResult、首字母大写字段和 Buckets/Bucket 包装层，也说明部分分页元素可以省略。 | 检查真实 XML 的包装层、大小写和分页字段是否存在。界面展平后的 JSON 不能证明 XML 没有包装层。服务级地址和类型化请求头仍需单独补充生成器能力。 |

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
| GetBucketCors | [打开](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketCors) | 包含多请求头 CORS 规则的桶 |
| GetBucketInfo | [打开](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketInfo) | 已开启日志的桶 |
| GetBucketInventory | [打开](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketInventory) | 桶和 SSE-OSS 清单配置 ID |
| ListBucketInventory | [打开](https://api.aliyun.com/api/OSS/2019-05-17/ListBucketInventory) | 包含 SSE-OSS 清单配置的桶 |
| GetBucketReplicationLocation | [打开](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketReplicationLocation) | 桶 |
| GetBucketReplicationProgress | [打开](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketReplicationProgress) | 桶和复制规则 ID |
| GetBucketWebsite | [打开](https://api.aliyun.com/api/OSS/2019-05-17/GetBucketWebsite) | 已有静态网站配置的桶 |
| ListBuckets | [打开](https://api.aliyun.com/api/OSS/2019-05-17/ListBuckets) | 名下已有桶的账号 |

## 修正前的门禁

- 本文提供候选修正依据，不批准运行时行为，也不直接解除生成器阻塞。
- 每项修正都需绑定来源版本、记录中英文决策并提供独立的离线 XML 测试，再重新生成。
- 保留完整 DSL 和上游原始字节。模型例外与序列化映射必须明确表达，并可复用。
- 其余 66 个操作当前受方法或正文格式、动态路径、正文模型缺失和服务级绑定限制；Explorer 证据不能代替这些生成器能力的实现。

## 验证要求

- 本次仅增加文档，执行文档语言与本地链接检查及 `git diff --check`。
- 文档修改无需运行 Go 或前端测试，也不需要调用真实云 API。
