# 发布前消费者证据更新

[English](release-consumer-refresh.md)

- 对应 #61，前置产品 issue #60、#74、#75 已关闭。
- 发布前，使用集成后的 SDK 重新验证独立模块中的 STS、ECS、VPC 消费者；保留旧记录和真实云调用、来源验证的原始提交。
- 本次由实现代理验收。自动测试耗时不代表人工使用体验；可选的独立人工评审 #76 仍未执行。

## 候选版本及结果

- SDK 提交：`0365dd8733ffa6aa86061b226d5e8f79d83e70d0`；Go 1.27.1，Windows/amd64。
- 两个独立模块的 `go vet ./...`、`go test -json -count=1 ./...` 和无需账号的 `go run .` 均于 2026-10-09 16:04 UTC 通过。
- STS：五项必需任务、十一项必需测试及两项额外协议一致性测试通过。
- ECS：十项必需测试、十四项子测试通过。VPC：八项测试、十项子测试通过。
- 当前记录：[STS](acceptance/sts-agent-result.json)、[ECS](acceptance/ecs-product-result.json)、[VPC](acceptance/vpc-product-result.json)。
- 历史记录：[STS](acceptance/sts-agent-result.historical.json)、[ECS](acceptance/ecs-product-result.historical.json)、[VPC](acceptance/vpc-product-result.historical.json)。早期收尾文档描述的是这些历史执行。
- 真实云调用和来源证据保留当时实际验证的提交。[本轮现有资源证据](live-resource-followup.zh-CN.md) 单独记录；本次消费者更新未调用云 API。
- 生成覆盖为 STS 4、ECS 380、VPC 403 个操作。这不代表全部操作已通过真实云验收，也不代表全部能力策略已经审核。

## 剩余发布门禁

- 只读发布检查要求已接受的源码、工作负载行为没有变化，且 main 工作区干净。
- 最终 main 提交须通过 Linux race、Windows 和自动化 CI；实现 PR 的检查不能替代这一门禁。
- 评审来源与策略锁定、Go 1.27、直接 JSON v2、MIT 运行时及 Apache 生成通知、配对发布说明和示例。
- 门禁通过后发布不可变的实验版 v0.1.0。本记录中的发布、相同版本网页索引仍未执行。
- 遵循[发布清单](sts-v010-release-checklist.zh-CN.md)，取得真实索引证据后再关闭 #61、#57 和 milestone。
