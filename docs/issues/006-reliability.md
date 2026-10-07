# 实现幂等性约束下的有界重试

### Affected areas

transport

### 前置依赖

依赖 #3、#5 的管线和 operation 幂等性元数据。

状态：GitHub issue #6（执行状态以 GitHub 为准）。

## 问题与证据

重试会影响写入安全和长时应用稳定性，须先定义行为。临时凭据刷新单独由 issue #9 交付。

来源：[上游证据](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/DEVGUIDE-CN.md)。具体边界见 [调研](../research.md)。

## 范围

实现有界重试策略；默认不自动重试写入。不包含凭据刷新。

## 验收条件

- [ ] MaxAttempts 指总尝试次数；区分可重试 HTTP 状态、服务错误和网络错误
- [ ] 默认重试关闭；非幂等写入不重试；尊重 Retry-After 和总 context deadline
- [ ] 退避带 jitter 和上限，可取消且测试不依赖真实长等待
- [ ] 并发/race 测试与取消测试；凭据或授权头不进入默认日志

## 文档要求

每项公开 API 同时交付 doc.go、导出符号/字段注释和可执行外部 Example。
更新 README 支持范围与相关设计/发布说明。文档完成与代码完成属于同一交付。

## 验证计划

按验收条件建立离线行为测试，运行文档检查、vet 和相关测试一次。
涉及并发时由 Linux CI 运行 race；上线或 pkg.go.dev 页面由浏览器确认。
