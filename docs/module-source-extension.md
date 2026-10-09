# Explicit Darabonba module extension

[中文](module-source-extension.zh-CN.md)

- Issue #92. Existing product import requires every transitive dependency to be pinned already.
- Use `tools/darabonba/import-modules.cjs` to add reviewed missing modules first, then `import-product.cjs` to add the product.
- The module importer never resolves wildcard versions. A reviewed plan supplies exact versions, archive SHA1/SHA256, official archive URLs and license evidence, bound to the current source manifest hash.

## Commands

```powershell
node tools/darabonba/import-modules.cjs metadata/module-imports/oss.json
# Optional: use previously downloaded, checksum-verified archives instead.
node tools/darabonba/import-modules.cjs metadata/module-imports/oss.json <archive-directory>
```

- Choose one command. Local archive names are `<scope>_<name>_<version>.tgz`.
- The network form downloads only the exact approved archives. Generation remains offline. Existing modules cannot be replaced, and repeated import fails before writes.
- Unknown imports or missing license evidence fail before writes. License evidence must be an existing pinned source-repository notice or an exact preserved archive README declaring Apache-2.0/MIT. This does not establish an unrelated native runtime's license.
- Archive size/expanded size are bounded. Reject escaping/absolute/Windows-reserved paths, case collisions, links and unsupported file types.
- Validate all current and new module dependencies before writes. Refresh every product's complete transitive map; preserve original module/source/notice bytes and write the manifest last.
- The final filesystem write phase is not transactional. A failure can leave a partial corpus; repair it under the issue before generation. Source verification must reject incomplete state.

## OSS plan and limits

- [Reviewed import plan](../metadata/module-imports/oss.json): GatewayOSS 0.0.42, OSSUtil 0.1.10, GatewayOSS_Util 0.0.8 and Time 0.0.2.
- Registry archive SHA1 and SHA256 were checked during research. The plan is an explicit future import selection, not proof the production corpus already contains these modules.
- Gateway module source notices reuse the pinned Alibaba Cloud gateway repository's Apache evidence. OSSUtil/Time preserve their archive Apache declarations. Native Go helper source is not copied or relicensed by this plan.
- Applying the plan changes the source manifest. Review and rebind affected policy/translation/projection hashes, regenerate all current IR/prose/Go outputs, and run Node 22 frontend tests/check plus sdkgen product-check and Go gates under the issue before merge.
- This change implements the reusable importer only. The production source corpus/IR/output remains unchanged. OSS product registration, exact XML traits, signing, emission and streaming are pending.
- [OSS protocol route](oss-xml-protocol.md) and [XML representation evidence](../tools/ossxml/README.md) define the next gates.
