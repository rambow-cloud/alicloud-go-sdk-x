# OSS XML representation comparison

[中文](README.zh-CN.md)

- Issue: #92. This module tests official source representations with synthetic XML.
- Run `go -C tools/ossxml test ./...` and `go -C tools/ossxml vet ./...` from the repository root.
- No network calls occur in tests; the first module download needs network access. No account, Profile or credentials are used.
- Dependencies belong to this isolated comparison module. The SDK runtime module has no Tea, official SDK or third-party XML dependency.

## Pinned inputs

- [OSS product DSL](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/oss-20190517/main.tea): the complete DSL remains the intended production schema source.
- [OSS Teafile](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/oss-20190517/Teafile) declares Go product release v2.0.1. Its [generated model source](https://github.com/alibabacloud-go/oss-20190517/blob/935681aa139c270f2b6743e57a3c0dab548f39b4/client/client.go) supplies the concrete response types in this comparison.
- The researched GatewayOSS 0.0.42 imports GatewayOSS_Util 0.0.8; that module declares Go helper v0.0.6. Its [registered XML roots](https://github.com/alibabacloud-go/alibabacloud-gateway-oss-util/blob/cd82cbd16bcb3f0988e125ee63679e887867f225/client/structs.go) and [parse entry point](https://github.com/alibabacloud-go/alibabacloud-gateway-oss-util/blob/cd82cbd16bcb3f0988e125ee63679e887867f225/client/client.go) supply the helper representation.
- The helper uses [tea-xml v1.1.3](https://github.com/alibabacloud-go/tea-xml/blob/v1.1.3/service/service.go). Exact module versions and checksum-database hashes are in go.mod/go.sum.
- This intentionally compares the product's declared model release with the researched gateway's declared helper release. It does not assert that an unchanged installation of OSS v2.0.1 uses helper v0.0.6, or that all official releases behave alike.

## Verified distinctions

- ACL and CORS: the helper retains a structured root wrapper. Passing it unchanged to the declared response model produces a body with missing nested fields. Explicit fixture unwrapping preserves those fields.
- Location: the scalar root already matches a DSL field. Keeping that wrapper preserves the value; decoding into the body directly with encoding/xml loses the root text.
- Malformed XML: the pinned helper suppresses the parse error and returns an empty map. The standard XML decoder returns an error.
- The external Example is deterministic and account-free. CI tests this isolated module on Windows and Linux race.

## Generator decision

- Normalize helper/model representations before deciding whether sources conflict.
- Derive root and scalar/structured distinctions from pinned native declarations and the complete semantic model. Do not infer roots from operation names or hand-author a permanent per-operation table.
- These tests contain explicit fixtures, not a production normalizer. Complete source/import/license pinning, no-write drift rejection, XML protocol emission, OSS signing and streaming are still pending.
- Production XML parsing must preserve fields and return malformed-body errors. Any intentional difference from the upstream helper must be recorded with source and behavior evidence.
- Follow the paired [OSS route](../../docs/oss-xml-protocol.md).
