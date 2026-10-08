# STS 来源修订演练

[English](sts-source-rehearsal.md)

- #60 在上述测试数据固定真实官方不可变修订、原始许可、blob/SHA256，将 2025-06-30 源码与 d2c0338636a58a6cafc5316d2ed5d158f1f5b162 比较。
- 这是仅 STS 来源、保留固定导入/语义解析器的限定范围的演练，不表示刚发布新 STS API 或升级所有产品。
- 候选 main/Teafile/ catalog 与生产字节相同；历史变化和对生产的无变化更新分别记录。

- 根目录完成 Node 22 前端检查/测试后执行上述命令；第一次 Go 生成应因旧来源绑定策略失败且不创建 service 输出，显式复制已审核策略后再生成/检查。
- 目录已存在时选择新目录。
- 工具写前核测试数据/候选已接受字节，用官方语义解析器将四个操作转换为 IR、清点模型，对比清单/ 绑定/类型/认证/说明/许可及确定性 IR，不改生产文件，不更新导入解析版本。

- 策略重新绑定显式审核为仅来源 manifest 哈希变化，候选 STS 输入字节及已审核准确路径相同；禁止静默批准 DSL 或无效策略变化。
- 编译时将当前运行时/credentials/endpoint/ 中间件/retry/internal signing/rpcmodel 源码、生成 STS、sdktest 复制到规范 module 路径的新独立模块并运行 Examples/测试，报告准确命令/结果，不从 IR 推断编译或真实覆盖。

- 另记录旧策略写前拒绝、非法/未知认证拒绝及原生产源码/IR/策略/生成物哈希未变的回滚证明。
- 合成不兼容测试数据与真实上游修订比较分别记载。

## 已审核的真实来源变更

- 历史 initializer 显式 v2，签名操作现在报告 DSL_PRODUCT_AUTH_INITIALIZER，不生成协议不兼容的 ACS3；历史 OIDC/SAML callApi 仍不支持。
- 当前候选继承 ACS3 默认且使用匿名 doRPCRequest。
- 原生模型/字段/类型/必需性与语义说明不变，说明坐标和 initializer 的 endpoint 映射改变。
- 所有历史地区的映射不升级为验收覆盖，必需真实地区仍为 cn-hangzhou。
- 上述共用命令的 --运行时自动复制独立运行时/契约并编译；旧策略和历史签名选择都需在 service 写入前失败。

## 可运行命令与示例

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
