# OSS checksum foundation

[中文](oss-checksums.zh-CN.md)

- Issue #92; follows the internal XML codec and OSS4 signer. Keep operation selection and integrity policy explicit in future official DSL/Gateway lowering.
- Add stdlib-only, cancellable CRC64 and Content-MD5 calculations for owned byte input. Preserve finalized CRC seed semantics for append/continuation; return no partial checksum on cancellation. Do not retain/mutate input or add native SDK dependencies.
- Bind CRC arithmetic to complete official Go SDK source/tests at `2614e2cfdc3618f204cee8bb247f9359e8e509aa`. Native `TestCRC64` uses Go `crc64.ECMA` with vectors `123456789` -> `995dc9bbdf1939fa` and `This is a test of the emergency broadcast system.` -> `27db187fc15bbc72`. Reuse factual vectors, not native implementation.
- Verify server CRC as an unsigned decimal uint64. Missing, nondecimal and overflow values fail explicitly; mismatches return a redacted sentinel. Never interpret ETag as a checksum or silently treat a missing header as verified.
- Accept input strings with leading zeros if they parse as uint64; reject whitespace/sign/hex syntax. Format no raw server value or body in errors. A zero-length body preserves the CRC seed and has the standard empty-body Content-MD5.
- Test independent official vectors, append/chunk continuation, zero input, MD5/base64 known vectors, input ownership, invalid/overflow/mismatch redaction and cancellation including mid-calculation. Add an account-free external Example, paired docs and Go/doc/vet/format checks; require exact-head CI before merge.
- This stage does not enable upload/download checks automatically, verify partial/range downloads, combine independently hashed multipart parts, read unbounded streams or change retry/replay. Stream EOF/early-close/context/ownership contracts and operation-specific source policy remain separate.
- Sources: [official CRC64 protocol](https://www.alibabacloud.com/help/en/oss/user-guide/check-data-transmission-integrity-by-using-crc-64), [fixed native test](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/2614e2cfdc3618f204cee8bb247f9359e8e509aa/oss/utils_crc_test.go), complete pinned GatewayOSS 0.0.42 (`main.tea` SHA256 `71caf417a396688b8a4a38b0e0f431d5adfb7ae043d98fa504f07ac1b319f44c`). Keep CRC/header absence policy distinct from checksum arithmetic.

## Verification

- Implementation and checks pending.
