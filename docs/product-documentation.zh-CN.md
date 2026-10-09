# 授权产品文档

[English](product-documentation.md)

- #93 [可选 canonical 说明](canonical-prose-enrichment.zh-CN.md)在表示和类型规范化后补充缺失的英文字段说明。DSL 原文、元数据补充、Go 行为注释及剩余缺口分别统计，保留准确来源指针和 Apache 归属。

- 阶段 #38 继 #37 / PR #42，在 `issue/38-licensed-product-docs` 叠加 `issue/37-capability-policies`；原依赖链现已[集成 main](generator-integration.zh-CN.md)。
- 本规格先于实现。
- 保持当前 RPC 支持范围及操作/模型数量，更多协议另行明确范围。

- 复用官方语义解析器的字段 description 字符串及操作 annotation token，从哈希固定的产品 DSL 提取；IR 增加说明/摘要及来源坐标，不另写 Tea 编译器，不要求逐操作文档补充配置。
- 上游 example 值仍仅保留坐标，不将账号、凭据、资源、CLI 命令或权限策略复制到可执行 Go Example；既有示例保持确定、离线且明确只演示调用，不冒充有效云参数。

- 为操作/字段输出英文为主的原生 Go 段落，保留本 SDK 的所有权、指针、重试、取消契约。
- 将 Markdown 链接/强调、标题、列表转成可读 Go 说明，保留安全 HTTPS 链接， 去掉 HTML/不安全 URL 标记，规范控制字符/换行并阻止编译器指令。
- 每行均为 // 注释； 原文不能生成代码，也不自动变成校验、必填、弃用或重试策略。
- 明确区分上游服务说明与 SDK 契约，并链接固定来源行。
- 缺失/空/纯非英文说明明确记录，不编造或自动翻译。

- 生成对应中英文指南，同步使用/行为契约、操作/来源索引及文档覆盖。
- 固定语料的语义说明为英文，不假设有固定中文翻译；两个语言指南指向同一授权来源/Go 文档， 明确此语言限制，不把英文摘录标成已翻译中文。
- 机器报告单独记录操作/字段说明覆盖、 缺失及归属，不将其混同于代码输出、能力审核或真实验收。

- 保留上游源码/许可原始字节；产品输出附版权、Apache-2.0 归属、明确转换声明及完整许可。
- 根 MIT 仍适用于原创运行时/工具，上游派生定义/说明不改标 MIT，不输出导入模块实现。
- 包通知及来源锁引用让分发后的生成包自身包含来源/许可信息。

- 验收覆盖官方语义解析器提取、真实语料/缺失/空注释、确定性、来源坐标、英文说明、 Markdown/HTML/指令安全转换、不复制原始示例或改变运行行为、许可/语言覆盖准确。
- 保护文件/产物清理测试同步覆盖文档/许可。
- 先运行前端/发现检查及相关 Node 测试， 再运行新旧生成检查、完整 Go 测试（含隔离编译/Example）、doccheck、vet、已跟踪格式、双语/空白检查及 Linux race/Windows CI。
- 不做真实调用、发布、索引或合并。
- 下方链接记录三份固定产品来源、上游许可证与完整 Apache 条款；这些资料于 2026-10-07 阅读。

## 参考资料

- [product corpus](https://github.com/aliyun/alibabacloud-sdk/tree/ec489e5c3deae95496daae2b41503ac58b221adb)
- [upstream license](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/LICENSE)
- [Apache-2.0 terms](https://www.apache.org/licenses/LICENSE-2.0.txt)
