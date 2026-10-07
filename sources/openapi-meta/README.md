# Canonical metadata fixtures / Canonical 元数据样本

## English

Issue [#34](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/34) pins three ECS
operation fixtures, version.json and the original Apache-2.0 LICENSE from
[aliyun/aliyun-openapi-meta](https://github.com/aliyun/aliyun-openapi-meta), revision
51286a65c79d008436eb314e636f9c9ad4b1ca08. manifest.json records exact file URLs and
SHA-256. Source/license bytes are retained, including upstream newline conventions;
this guide is project-authored. The upstream repository warns that its structure is
unstable and currently intended for CLI builds. These fixtures test adapter version 1;
they are optional enrichment, not the operation inventory or a prerequisite per API.

From the repository root, explicit import accesses the network:

```sh
node tools/darabonba/import-canonical.cjs
```

Import fetches all artifacts before writing. Filesystem failure can leave a partial
update; repair it before generation. Normalization/checks use local bytes and reject
modified, unlisted, unsafe or symlinked source files. Review a new revision/license/
adapter under an issue; never refresh metadata during ordinary generation.

See [source normalization](../../docs/source-normalization.md) for exact wire case,
indexed bindings, itemName wrappers, source attributes and evidence limits. Neither
CLI examples nor backend annotations establish retry safety or actual HTTP behavior.
No credential or account data is included.

## 中文

Issue [#34](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/34) 固定
[aliyun/aliyun-openapi-meta](https://github.com/aliyun/aliyun-openapi-meta) 的三个 ECS
操作样本、version.json 和原始 Apache-2.0 LICENSE，revision 为
51286a65c79d008436eb314e636f9c9ad4b1ca08。manifest.json 记录准确 URL 和 SHA-256。
来源/许可保留原字节及上游换行；本指南为项目编写。上游声明结构不稳定、目前用于
CLI 构建。本样本验证版本 1 适配器，作为可选补充，不是操作清单或逐 API 前置要求。

仓库根目录执行英文章节的 `import-canonical.cjs` 才会联网；先获取全部产物再写入。
文件系统失败可能留下部分更新，生成前须修复。规范化/检查使用本地字节，拒绝修改、
未列出、不安全或符号链接来源。新 revision、许可、适配器均在 issue 下审核，普通
生成不刷新元数据。

准确线字段大小写、索引绑定、itemName 包装、来源属性及证据边界见
[来源规范化](../../docs/source-normalization.md)。CLI 示例或 backend 注解不能证明
重试安全和真实 HTTP 行为。样本没有凭据或账号数据。
