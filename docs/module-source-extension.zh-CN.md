# 显式扩展 Darabonba 模块来源

[English](module-source-extension.md)

- OSS 模块和完整产品源码现已按 [OSS 语义 IR](oss-semantic-ir.zh-CN.md)固定为暂存输入。原模块扩展计划保留导入前的来源锁，属于历史证据，不应在已经扩展的来源锁上重复执行。

- 对应 #92。现有产品导入工具要求所有传递依赖都已固定。
- 先用 `tools/darabonba/import-modules.cjs` 导入审核过的缺失模块，再用 `import-product.cjs` 添加产品。
- 模块导入器不解析通配版本。审核计划明确给出准确版本、归档 SHA1/SHA256、官方归档地址和许可证据，并绑定当前来源 manifest 哈希。

## 命令

```powershell
node tools/darabonba/import-modules.cjs metadata/module-imports/oss.json
# 也可以使用此前已下载并校验过的本地归档。
node tools/darabonba/import-modules.cjs metadata/module-imports/oss.json <archive-directory>
```

- 两条命令选择一条执行。本地归档文件名为 `<scope>_<name>_<version>.tgz`。
- 联网方式只下载计划中的准确归档，生成过程仍离线运行。不允许替换已有模块；重复导入在写入前失败。
- 未知依赖或缺少许可证据时，在写入前失败。许可证据须为现有固定源码仓库的许可通知，或原样保留且声明 Apache-2.0/MIT 的归档 README。这不代表其他原生运行时的许可已经确认。
- 限制归档及解压后的大小，拒绝越界路径、绝对路径、Windows 保留名称、大小写冲突、链接和不支持的文件类型。
- 写入前检查现有及新增模块的全部依赖。更新所有产品的完整传递映射，保留原始模块、来源和许可字节，最后写 manifest。
- 最终文件系统写入不是事务。失败可能留下部分来源；须在 issue 下修复后再生成，来源验证必须拒绝不完整状态。

## OSS 计划及范围

- [审核导入计划](../metadata/module-imports/oss.json)：GatewayOSS 0.0.42、OSSUtil 0.1.10、GatewayOSS_Util 0.0.8、Time 0.0.2。
- 调研时已检查 registry 归档 SHA1 和 SHA256；计划只是明确的后续导入选择，不代表生产来源已包含这些模块。
- Gateway 模块沿用固定阿里云 gateway 源码仓库的 Apache 许可证据；OSSUtil/Time 保留归档中的 Apache 声明。本计划不复制原生 Go helper 源码，也不重新标注其许可。
- 执行计划会改变来源 manifest。合并前须在 issue 下审核并重绑受影响的策略、翻译和投影哈希，重新生成当前所有 IR、说明和 Go 产物，再执行 Node 22 前端测试/检查、sdkgen product-check 及 Go 门禁。
- 本变更只实现可复用的导入工具，尚未修改生产来源、IR 或生成产物。OSS 产品注册、准确 XML 特征、签名、生成和流式处理仍待完成。
- 下一步遵循[OSS 协议路线](oss-xml-protocol.zh-CN.md)及[XML 表示差异证据](../tools/ossxml/README.zh-CN.md)。
