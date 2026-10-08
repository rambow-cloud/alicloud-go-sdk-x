# ACS3 signing

[中文](signing.zh-CN.md)

- The internal signer implements ACS3-HMAC-SHA256 for the exact transmitted method, escaped path, canonical query, selected headers and SHA256 body hash.
- RPC uses `/`; ROA paths preserve segment escaping, including encoded slashes.
- Query spaces use `%20`, never `+`; repeated values are sorted deterministically.
- Host, content-type and all x-acs headers participate.
- STS tokens are signed and stale tokens are removed on rotation.
- Every runtime attempt gets a fresh cryptographic nonce, timestamp and credential snapshot.
- The fixed-vector test matches the [official signature fixture](https://help.aliyun.com/zh/sdk/product-overview/v3-request-structure-and-signature); it requires no account.
- OSS/SLS and other custom signing protocols are outside scope.
- Signing is internal; middleware modifying signed values must do so before calling next in Finalize.
