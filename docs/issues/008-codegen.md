# 建立可追溯产品模型生成与依赖基准

### Affected areas

tools, ecs

### 前置依赖

先验证 #5 的 API 设计并确定 schema 许可。

状态：GitHub issue #8（执行状态以 GitHub 为准）。

## 问题与证据

需要降低生成代码使用成本并及时覆盖服务模型更新，相关效果须可测量。

来源：[上游证据](https://github.com/alibabacloud-go/ecs-20140526/blob/master/client/client.go)。具体边界见 [调研](../research.md)。

## 范围

固定 schema 输入、生成器版本、Go 原生注释模板，另建可复现官方 SDK 对照基准。

## 验收条件

- [ ] 生成器与运行时依赖分离，schema 来源/日期/license 可追溯
- [ ] 生成结果确定，二次生成无 diff，含 doc comments 和操作示例框架
- [ ] 避免 map[string]interface{} 成为公开产品 API；保留必要 optional
- [ ] 比较固定 Go/OS/架构/SDK 版本下依赖数量、编译时间、二进制大小
- [ ] 提供 schema 差异评审与兼容性规则，不手改生成产物

## 文档要求

每项公开 API 同时交付 doc.go、导出符号/字段注释和可执行外部 Example。
更新 README 支持范围与相关设计/发布说明。文档完成与代码完成属于同一交付。

## 验证计划

按验收条件建立离线行为测试，运行文档检查、vet 和相关测试一次。
涉及并发时由 Linux CI 运行 race；上线或 pkg.go.dev 页面由浏览器确认。
