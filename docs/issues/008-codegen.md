# Build product generation after foundation acceptance

## English

Foundation #19 passed on 28684e47ec1869dd88de0cc14faeec9bc7e6adb3. Follow the ordered implementation path in docs/generator.md.

- [ ] #21: [Feature]: Import pinned OpenAPI protocol metadata and validate generator IR
 - [ ] #22: [Feature]: Generate typed ECS and STS clients and reviewed paginator/waiter adapters
 - [ ] #23: [Maintenance]: Enforce deterministic regeneration and generator acceptance in CI

The first working profile generates four RPC operations, selected models, codecs, mock interfaces, reviewed pagination/waiter adapters, English Go docs, offline Examples and bilingual guides. Offline deterministic generation and CI drift checks are required. ROA/general body encoding and wider service coverage require later issues. Benchmarks remain #20.

### Affected areas

tools

## 中文

基础 #19 已验收；按 docs/generator.md 的顺序完成元数据/IR、客户端与适配器生成、再生成与 CI 验收：#21  #22  #23 。首版四个 RPC 操作及字段子集，复用共享基础能力，交付英文注释、离线 Example、双语指南及确定性门禁。ROA/body 和更多产品另建 issue，基准 #20 独立。
