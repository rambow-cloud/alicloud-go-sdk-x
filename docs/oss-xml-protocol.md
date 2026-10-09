# OSS XML and streaming follow-up

[中文](oss-xml-protocol.zh-CN.md)

- Issue #92. Research only; production source/IR/runtime has not changed.
- Official OSS product: oss-20190517 at ec489e5c3deae95496daae2b41503ac58b221adb; original main.tea SHA256 0713683b286a49e4e80fff023f61db426dc035fb8755a7b4616a82f65c22e6ff. 90 WithOptions functions hand off to execute, with SPI/GatewayOSS initialization and bucket hostMap.
- Source response modes: XML 79, JSON 3, none 4, string 1, binary 3. Request modes: XML 84, binary 3, JSON 2, multiFormData 1. These are source counts, not parser/lowering/emission acceptance.
- Official registry research resolved GatewayOSS 0.0.42; archive SHA1 0b31bfdd4c86a28a6c93d74c0622e3167558435f, SHA256 5bf16e8d5283b8174c34d17fd4cb085422c7ef64a8d129beb93dd404216b7cab. This research resolution is not a production pin. Additional OSSUtil/GatewayOSS_Util/Time imports and license/source evidence must be fixed before production generation.
- GatewayOSS has dedicated signing, bucket-host/subresource behavior, XML body/error parsing and stream/checksum rules. Do not route OSS through ACS3 by merely adding encoding/xml.
- Before code: commit paired route/acceptance, pin complete original source/helper/imports, extract exact helper/root/host/signing evidence into IR, then implement shared protocol helpers and generic Go emission. Keep unknown modes/selected unsupported operations rejected before writes.
- Establish stream transfer/close, context lifetime, request hashing/replay and response checksum semantics explicitly. Use independent signature/XML/stream ownership fixtures, synthetic external Examples and normal frontend/Go/CI gates. No OSS provisioning or live mutations.
- FC InvokeFunction needs its own readable-body/header transformation/lowercase error coverage; do not pretend buffered JSON/none closes binary streaming.

## First verification

- Add an isolated, pinned official-helper/model comparison under tools/ossxml. Use synthetic ACL, CORS and location XML; make no HTTP or credential calls.
- Assert structured-root and scalar-root representations separately. Compare unchanged helper conversion with explicit fixture normalization; do not introduce per-action production overlays.
- Record malformed XML handling. New runtime decoding must return an error rather than copy a helper's silent empty result.
- This comparison supplies source/representation evidence only. Production source pinning, semantic projection, signing, XML emission and streaming remain pending.
