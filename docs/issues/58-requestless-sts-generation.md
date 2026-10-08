## English

### Problem and evidence

Pinned STS main.tea:500 exposes getCallerIdentityWithOptions(runtime) without a request-model parameter. Coverage reports DSL_OPERATION_SIGNATURE.

### Scope and dependencies

Member of #57; accepted generator baseline #35/#36/#38. Support this reviewed signature in semantic projection/normalization, complete IR and Go emission without per-operation field overlays or handwritten generated clients. Preserve AWS-style context/input/options API conventions and define/document the empty input representation. Keep exact response fields, signed AK protocol behavior, mock interface and Metadata. No dependency on the parent issue's closure.

### Acceptance

- [ ] GetCallerIdentity lowers/emits deterministically with complete reachable models.
- [ ] Independent requestless/invalid-signature, wire/signing/empty-query/response/absence/error/context/ownership contracts pass; unsupported signatures fail before writes.
- [ ] Generated external Example, English symbol docs and paired guide/source/coverage report are complete.
- [ ] Existing ECS/VPC/STS outputs and signed-operation credential requirements are preserved.

### Verification

Node 22 frontend checks/tests; sdkgen check and product-check; doccheck, format, vet, Go tests, Linux race/Windows CI. No live cloud calls in unit tests.

## 中文

官方 main.tea:500 无请求模型参数导致 DSL_OPERATION_SIGNATURE。归属 #57，基于已接受 #35/#36/#38，不依赖父项关闭。完善前端/IR/生成器的无请求签名和空输入表示，保留签名、准确响应、mock/Metadata/AWS 调用范式；不逐字段 overlay、不手写生成客户端。确定性生成及独立线路/签名/存在/错误/取消/所有权和非法签名写前失败测试、外部 Example、英文注释/双语指南/来源/覆盖必需；不破坏其他产品和签名凭据要求。执行上述 Node/Go/CI 门禁，不在单元测试调用云。
