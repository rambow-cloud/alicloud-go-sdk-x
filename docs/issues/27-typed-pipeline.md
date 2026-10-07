GitHub issue: #27.

## English

### Problem and evidence

Review of abcf961 found behavior defects and incomplete AWS Go SDK v2 design alignment. Design and acceptance were written before code in docs/aws-style-remediation.md.

### Scope and dependencies

Introduce owned model input/output and a Serialize stage; execute Initialize/Serialize/Build once and Finalize/Deserialize per attempt. Generate concrete service Options, NewFromConfig, configuration snapshots and isolated per-call configuration including retry overrides. Retain New(Config) and wire Invoke; document migration.

Dependency: #26.

### Acceptance criteria

- [ ] Typed hooks alter copied inputs before signed encoding; output hooks publish atomically. Verify stage order, retry counts, invalid replacement/short circuits, input ownership, configuration snapshot copying, per-call retries/transports/endpoints and existing protocol/STS/OTel fixtures.
- [ ] English-primary Go docs, deterministic offline Examples, equivalent bilingual docs and migration notes; offline regeneration, doccheck, language, vet, tests, Linux race and Windows CI.
- [ ] Record exact commit and CI evidence before closing.

### Affected areas

core, middleware, tools

## 中文

依据 abcf961 review，设计先写在 docs/aws-style-remediation.md，再实现；依赖 #26。

增加自有输入/输出与 Serialize 阶段；生成独立服务 Options、NewFromConfig、配置快照和隔离调用覆盖（含 retry）。保留 New(Config)/线 Invoke。验收类型 hook 修改及签名编码、输出原子、顺序/重试计数、非法替换/短路、所有权、配置复制与既有协议/STS/OTel。

英文注释、离线 Example、中英文对应文档与迁移说明、再生成/文档/双语/vet/测试/Linux race/Windows CI 通过后记录提交证据再关闭。默认策略、ROA/body 和基准保持独立。
