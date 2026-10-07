# Sparse capability policies / 稀疏能力策略

## English

Follow [#37 specification](../docs/capability-policy.md). Policies bind product/version/
source-lock hash and record evidence; the generator resolves exact wire paths against
complete IR. They supply reviewed paginator/waiter/retry/client-token/validator/sensitive/
naming behavior, not per-field model declarations. Missing policy is allowed and means
unreviewed/no Standard retry. Generated coverage records policy file hashes and individual
capabilities. Do not edit generated adapters directly or infer support from names.

Initial policies cover seven actions: four paginators, one reusable waiter, five
idempotent reads, one conservatively non-retrying ClientToken write and STS sensitivity.
This is not full capability or live acceptance. Evidence links are review records,
not mutable network generation inputs. All generation and Examples run offline.

## 中文

遵循 [#37 规格](../docs/capability-policy.md)。策略绑定产品/版本/来源锁哈希及证据，
生成器按完整 IR 解析准确线路径。只补审核分页/waiter/重试/client token/validator/
敏感/命名行为，不逐字段声明模型；缺策略允许且表示未审核/不被 Standard 重试。
生成覆盖记录策略文件哈希及每操作能力。不直接编辑生成适配器或凭名称猜支持。

首批七操作：四分页、一可复用 waiter、五幂等读取、一保守不重试的 ClientToken 写操作
及 STS 敏感保护。不是完整能力/真实验收；证据链接供审核，不作为可变联网生成输入，
所有生成及 Example 离线运行。
