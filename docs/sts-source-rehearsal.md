# STS source revision rehearsal / STS 来源修订演练

## English

Issue #60 uses real immutable official revisions, original license bytes and blob/
SHA256 provenance under `tools/darabonba/fixtures/sts-revisions`. It compares the
2025-06-30 source to repository revision d2c0338636a58a6cafc5316d2ed5d158f1f5b162.
This is a bounded STS-source-only rehearsal with accepted pinned imports/parser,
not a claim that a newer STS API was just released or that all products were upgraded.
The candidate STS main/Teafile/catalog match production bytes; historical evolution
and this no-change production update are reported separately.

From the root, after Node 22 frontend checks/tests:

```powershell
node tools/darabonba/rehearse-sts.cjs .git/sts-source-rehearsal
go run ./internal/cmd/sdkgen product-generate -root .git/sts-source-rehearsal/candidate
# Expected failure: stale source-bound policy; no service output may exist.
go run ./internal/cmd/sdkgen product-generate -root .git/sts-source-rehearsal/baseline -operations sts/AssumeRole
# Expected failure: historical v2 initializer; no service output.
Copy-Item .git/sts-source-rehearsal/reviewed-sts-policy.json .git/sts-source-rehearsal/candidate/policies/sts.json
go run ./internal/cmd/sdkgen product-generate -root .git/sts-source-rehearsal/candidate
go run ./internal/cmd/sdkgen product-check -root .git/sts-source-rehearsal/candidate
node tools/darabonba/rehearse-sts.cjs --runtime .git/sts-source-rehearsal/candidate
go -C .git/sts-source-rehearsal/candidate test ./service/sts
```

Choose a new output directory if it already exists. The tool verifies fixtures and
accepted candidate bytes before writing, lowers all four actions with the official
parser, accounts for complete models, compares inventory/bindings/types/auth/prose/
licenses and checks deterministic IR. No production file is changed. Imported module
versions remain pinned; import resolution updates are outside this rehearsal.

The policy rebind is explicitly reviewed: only the source-manifest hash changes,
with byte-identical candidate STS inputs and unchanged reviewed native field paths.
Do not silently rebind a changed DSL or invalid policy. To compile candidate output,
use --runtime to copy the current runtime/credentials/endpoint/middleware/retry/internal
signing/rpcmodel sources into a fresh standalone canonical module, add generated STS,
sdktest and independent contracts, and run its Examples/tests. Record exact commands/result
in the acceptance report. Compilation and live coverage are not inferred from IR.

Also record expected-negative stale-policy failure before writes, malformed/unknown
authentication selection rejection, and rollback by proving the original source/IR/
policy/generated artifacts' hashes unchanged. Synthetic incompatible fixtures remain
distinct from the real upstream revision comparison.

## 中文

#60 在上述 fixture 固定真实官方不可变修订、原始许可、blob/SHA256，将 2025-06-30
源码与 d2c0338636a58a6cafc5316d2ed5d158f1f5b162 比较。这是仅 STS 来源、保留固定
导入/parser 的有界演练，不表示刚发布新 STS API 或升级所有产品。候选 main/Teafile/
catalog 与生产字节相同；历史变化和对生产的无变化更新分别记录。

根目录完成 Node 22 前端检查/测试后执行上述命令；第一次 Go 生成应因旧来源绑定策略
失败且不创建 service 输出，显式复制已审核策略后再生成/检查。目录已存在时选择新目录。
工具写前核 fixture/候选已接受字节，用官方 parser 降低四操作、清点模型，对比清单/
绑定/类型/认证/说明/许可及确定性 IR，不改生产文件，不更新导入解析版本。

策略重新绑定显式审核为仅来源 manifest 哈希变化，候选 STS 输入字节及已审核准确路径
相同；禁止静默批准 DSL 或无效策略变化。编译时将当前 runtime/credentials/endpoint/
middleware/retry/internal signing/rpcmodel 源码、生成 STS、sdktest 复制到规范 module
路径的新独立模块并运行 Examples/测试，报告准确命令/结果，不从 IR 推断编译或真实覆盖。

另记录旧策略写前拒绝、非法/未知认证拒绝及原生产源码/IR/策略/生成物哈希未变的回滚
证明。合成不兼容 fixture 与真实上游修订比较分别记载。

## Reviewed real drift / 已审核真实漂移

### English

The historical initializer explicitly sets v2; signed actions now report DSL_PRODUCT_AUTH_INITIALIZER rather than being emitted with incompatible ACS3. Historical OIDC/SAML callApi remains unsupported. Current candidate inherits the ACS3 default and uses anonymous doRPCRequest. Native models/fields/types/requiredness and semantic prose are unchanged; prose coordinates and initializer endpoint mappings change. Endpoint-map parity across all historical regions is not promoted to accepted coverage; required live region remains cn-hangzhou. The runtime-fixture command in the shared snippet automates standalone copying/compilation; both old-policy and historical signed selections must fail before service writes.

### 中文

历史 initializer 显式 v2，签名操作现在报告 DSL_PRODUCT_AUTH_INITIALIZER，不生成协议不兼容的 ACS3；历史 OIDC/SAML callApi 仍不支持。当前候选继承 ACS3 默认且使用匿名 doRPCRequest。原生模型/字段/类型/必需性与语义说明不变，说明坐标和 initializer 的 endpoint 映射改变。所有历史地区的映射不升级为验收覆盖，必需真实地区仍为 cn-hangzhou。上述共用命令的 --runtime 自动复制独立 runtime/契约并编译；旧策略和历史签名选择都需在 service 写入前失败。
