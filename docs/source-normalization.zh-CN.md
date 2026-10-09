# 来源规范化

[English](source-normalization.md)

- 当前客户端路径以[服务整合 #81](service-consolidation.zh-CN.md)为准：旧 services/ 包和 Go 输出器已移除，generate/check 均检查完整产品。下文旧流程仅保留历史证据。

- 阶段 [#34](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/34) 在 #31 五操作兼容桥上实现[权威路线](product-generator-roadmap.zh-CN.md)第一阶段。
- 完整官方 DSL 和语义语义解析器为主；可选固定 CLI 元数据经版本化适配，保留来源、中英文说明、CLI/backend 属性，和线属性分开。
- 不增加运行时依赖或公共 API。

- Canonical ECS DescribeImages/DescribeRegions/DescribeInstances、version.json 及 Apache-2.0 LICENSE 固定在 [sources/openapi-meta](../sources/openapi-meta/README.zh-CN.md)。
- 来源锁在桥接投影写入前检查 revision URL、SHA-256、路径和清单。
- Node 前端对存在的可选补充核对选定 DSL 输入/协议；Go 独立对照选定快照及投影，检查索引叶子和显式兼容例外。
- Go 生成读取固定 DSL 投影，不需要 Node 或 canonical 样本。

| 来源表示                                                              | 规范化语义                                                   |
| --------------------------------------------------------------------- | ------------------------------------------------------------ |
| name: region_id、raw_name: RegionId、options: --biz-region-id         | API 字段名准确为 RegionId，CLI 名称/选项保存为属性           |
| bool/int、location、required、param_style: repeatList、element.fields | 布尔/整数、query 绑定、来源必填性和重复结构模型              |
| Key/Value 与已弃用 key/value                                          | 大小写不同成员，不合并或重命名                               |
| Filter.1.Key/Value 到 Filter.4.Key/Value                              | DSL Filter 数组 Key/Value 叶子的八个字面绑定，不推断长度上限 |
| Images 数组并带 itemName: Image                                       | Images 对象内 Image 数组，递归恢复嵌套包装                   |
| backendName/nullToEmpty/valueMapping                                  | 保留来源注解，不自动转换 SDK 数据                            |
| 顶层 GET\|POST 与 operation.method POST                               | CLI 允许方法单独保留，实际 POST 配方交叉核对                 |
| operation_type: read                                                  | 来源注解，不能证明重试安全                                   |

- 索引接受没有前导零的正十进制位置，限制结构嵌套深度，成员大小写准确。
- 同一根字段同时直接/展开绑定、投影漏绑/伪造、类型/大小写漂移、索引必填性不一致都在写输出前失败。
- 没有 API 必填证据的必填索引根也失败。
- 既有 API 必填/DSL 可选批准不变； DSL 可选不证明服务端接受省略。

- 来源评审发现真实兼容例外：DescribeInstances.Tag 元数据包含当前 DSL 没有的可选 Tag[].key/value。
- 双语决策文档及机器策略明确列出这两个可选字符串 metadata-only 路径，继续不进入当前 Go 公开子集。
- 新缺字段、大小写变化、成员新必填均拒绝，不忽略大小写或推断线别名。
- Filter 八个绑定对应后，从 DSL 独有批准清单移除。

- 兼容桥仍按既有选择策略输出五操作。
- 无需逐操作快照的全量操作/模型发现属于 #35， 批量输出属于 #36。
- 本阶段不表示二者完成或所有 canonical 响应字段已真实验证。
- 选定响应/模型字段继续执行 Go 交叉核对；适配器测试比较真实 DescribeImages 嵌套包装与官方语义解析器输出。

- 安装工具后在根目录执行本文件所列的投影、生成、检查、npm test、sdkgen check、 doccheck、vet 和 Go 测试命令；均离线。
- 测试包含真实大小写/风格/必填/包装数据、 不合法元数据和来源修改、超出样本索引的位置、检查前不写文件。
- 既有跨后端测试逐字节检查所有生成 Go 文件，公共 Example/包契约不应改变。
- 保留 Go 1.27、直接 JSON v2、英文为主注释。

- 本阶段浏览器/真实核验为 **NOT RUN**。
- 进一步证据可打开下方 Explorer 页面：DescribeImages 检查请求 Tag/Filter 和响应 Images.Image 包装；DescribeInstances 检查 Filter 与原生分页。
- 在双语决策文档中把 UI 提示、CLI 本地校验、授权真实 HTTP 调用响应分别记录。
- 此前 #30 读取属于历史证据，不能证明本次规范化变更。

## 可运行命令与示例

```sh
node tools/darabonba/frontend.cjs generate
go run ./internal/cmd/sdkgen generate
node tools/darabonba/frontend.cjs check
cd tools/darabonba
npm test
cd ../..
go run ./internal/cmd/sdkgen check
go run ./internal/cmd/doccheck
go vet ./...
go test ./...
```

## 参考资料

- [DescribeImages](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeImages)
- [DescribeInstances](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstances)
