# Official source lock

[中文](README.zh-CN.md)

- Products come from [aliyun/alibabacloud-sdk](https://github.com/aliyun/alibabacloud-sdk/tree/ec489e5c3deae95496daae2b41503ac58b221adb) at commit ec489e5c3deae95496daae2b41503ac58b221adb.
- Keep main.tea, Teafile and api-info.json bytes intact for ECS 2014-05-26, STS 2015-04-01, VPC 2016-04-28 and FC 2023-03-30.
- The upstream [Apache notice](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/LICENSE) is preserved as LICENSE.upstream.
- Product source descriptions are part of this licensed corpus; public metadata snapshots separately exclude descriptions/examples.
- Product comments now reuse licensed parser descriptions/annotations under #38; paired usage/contracts/source indexes record language and prose coverage.
- Source example values remain excluded from executable Examples.
- Package LICENSE/NOTICE preserve Apache terms, copyright and transformation attribution; original runtime/tooling use MIT.

- The lock records product paths, original URLs, per-file SHA-256, parser version, module registry URLs, archive SHA-1/SHA-256 and resolved scope/name/version.
- Product .libraries.json maps wildcard specs to local versioned modules; generation performs no registry calls.
- Imported module artifacts, README and notices remain byte-exact.
- Our bilingual guides explain their use; language checks distinguish these preserved third-party artifacts from project-authored Markdown.

| Module                    | Pinned version |
| ------------------------- | -------------- |
| alibabacloud Credential   | 0.5.17         |
| alibabacloud EndpointUtil | 0.2.1          |
| alibabacloud GatewayPOP   | 0.1.4          |
| alibabacloud GatewaySPI   | 0.0.15         |
| alibabacloud OpenApi      | 0.3.23         |
| alibabacloud OpenApiUtil  | 0.2.11         |
| alibabacloud Paginator    | 0.0.3          |
| darabonba Array           | 0.1.1          |
| darabonba EncodeUtil      | 0.0.6          |
| darabonba Map             | 0.0.5          |
| darabonba SignatureUtil   | 0.0.11         |
| darabonba String          | 0.0.13         |
| darabonba Util            | 0.2.19         |
| darabonba XML             | 0.1.14         |

- License evidence is stored in licenses/\*.NOTICE with file SHA-256 and upstream Git blob identities in the manifest.
- These contain original LICENSE or README license declarations; full Apache terms are included in the gateway/paginator/crypto notices.
- Module source repositories declare Apache-2.0.
- Credential's registry archive supplies no license declaration; its referenced Node runtime declares MIT, recorded only as related-runtime evidence with sourceSPDX=NOASSERTION.
- That evidence does not prove the registry DSL's license.
- Preserve this provenance distinction for release review; do not relabel the entire corpus MIT or infer one implementation's license for another.
- Build-tool package licenses/integrities are separately recorded in package-lock.json.

- Complete product/import semantic analysis does not mean every operation is lowered or every helper is executed.
- The compatibility bridge recognizes five SDK functions; the first accepted product backend emitted 579 supported RPC operations using our runtime. [Review/integration evidence](../../docs/generator-integration.md) records scope and remaining unsupported behavior.
- Module sources supply parser declarations and type checks.
- The pinned Paginator module is a params factory declaration, not our unified paginator/waiter engine.
- No imported Go Tea runtime implementation is added to the SDK.

- See [tools](../../tools/darabonba/README.md), [migration](../../docs/darabonba-migration.md) and [decisions](../../docs/darabonba-decisions.md).
