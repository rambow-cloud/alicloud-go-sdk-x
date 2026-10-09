# FC ROA product generation

[中文](fc-roa-product.zh-CN.md)

- Issue #92. Depends on the explicit response-mode runtime stage in PR #109.
- Use the complete official `fc-20230330` DSL at `ec489e5c3deae95496daae2b41503ac58b221adb`, its original Teafile/catalog and already pinned import modules. Preserve all upstream bytes and notices.
- Official parser research found 73 WithOptions actions: 54 JSON responses, 18 none and one binary. Discovery must retain all actions; emission must report unsupported cases explicitly.
- Recognize exact path interpolation, header maps, guarded query fields, whole/body-field JSON conversion and callApi handoff. Do not accept arbitrary DSL statements or infer bindings from names.
- Build operation input facades from actual path parameters, request-model fields and headers. Keep source coordinates and exact wire names. Shared body models need separate operation output facades with common model definitions preserved.
- Empty/bodyless outputs retain operation metadata. Use explicit runtime none mode only from the source declaration. Binary remains unsupported until stream ownership and replay rules are implemented.
- Field Go comments describe actual path/query/header/body roles. Body is not a JSON wrapper member and Headers is not a header named headers; JSON outputs and metadata-only none outputs are documented separately.
- The pinned `ap-southeast-5` endpoint contains a trailing tab. Record the exact source hash, original value and reviewed trim decision; keep the source bytes unchanged and reject unapproved drift before writes.
- This stage is offline. No FC resources or live mutations are authorized by this plan.

## Verification

- Complete source discovery and deterministic parser/IR/generator checks. Selected binary or unknown transforms must fail before writes.
- Independent generated-wire fixtures for escaped multi-parameter paths, JSON body/container shape, headers, query absence/zero, JSON/none/error responses and input ownership.
- Required-field Examples use synthetic values and the offline transport. Package/symbol docs and paired product guides must be generated with source/licensing attribution.
- Run Node 22 frontend tests/check before Go doccheck, vet, tests and generation checks. Require exact-head Linux race and Windows CI.
- Counts for discovery, lowering, emission, compilation and live behavior stay separate. XML, binary streaming and additional signing profiles remain #92 follow-ups.

## Delivered offline scope

- IR/lock schema v5, profile openapi-json-v1. 73 discovered, 72 lowered/emitted/compiled, 331 emitted models and 72 runnable Examples. InvokeFunction is excluded with DSL_RESPONSE_BODY_PROFILE; no binary stream claim.
- Source manifest SHA256: fda7bc19fbbeb45b230c5dc9c4e69452a1dfe57331452f62dc653261b6d13254. Existing STS/ECS/VPC policy/translation bindings were re-bound to this expanded corpus without changing reviewed policy or translations. Their executable generated Go remains unchanged.
- Reuse fixtures rename the ROA product/action. Unknown statements, guards, transforms, query encodings, path bindings, XML modes and unsupported selections fail before emission; invalid endpoint approvals fail before IR writes.
- Field business prose: 1118/1591 fields have source English; all 1591 have generated Go contracts. No FC canonical prose corpus or reviewed Chinese business translation is added. Paired guides cover usage, behavior and sources. Missing business prose remains #93.
- Node 22: 98 tests covered; 90 passed initially and the affected 29-test set passed after fixture corrections. Frontend/check passed. Go full-suite failures were corrected and affected FC/codegen suites passed; all other packages passed in the initial run. Doccheck, vet, product-check, formatting, diff whitespace and language/local-link checks passed. Final-head Linux race/Windows evidence belongs to the linked PR.
- No FC live calls, resource creation or inferred pagination/waiter/retry policy. XML, binary streaming and additional signing profiles keep #92 open.
