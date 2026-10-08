# Canonical metadata fixtures

[中文](README.zh-CN.md)

- Issue [#34](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/34) pins three ECS operation fixtures, version.json and the original Apache-2.0 LICENSE from [aliyun/aliyun-openapi-meta](https://github.com/aliyun/aliyun-openapi-meta), revision 51286a65c79d008436eb314e636f9c9ad4b1ca08. manifest.json records exact file URLs and SHA-256.
- Source/license bytes are retained, including upstream newline conventions; this guide is project-authored.
- The upstream repository warns that its structure is unstable and currently intended for CLI builds.
- These fixtures test adapter version 1; they are optional enrichment, not the operation inventory or a prerequisite per API.

- From the repository root, explicit import accesses the network:

```sh
node tools/darabonba/import-canonical.cjs
```

- Import fetches all artifacts before writing.
- Filesystem failure can leave a partial update; repair it before generation.
- Normalization/checks use local bytes and reject modified, unlisted, unsafe or symlinked source files.
- Review a new revision/license/ adapter under an issue; never refresh metadata during ordinary generation.

- See [source normalization](../../docs/source-normalization.md) for exact wire case, indexed bindings, itemName wrappers, source attributes and evidence limits.
- Neither CLI examples nor backend annotations establish retry safety or actual HTTP behavior.
- No credential or account data is included.
