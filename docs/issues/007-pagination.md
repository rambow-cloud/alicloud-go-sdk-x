# 实现可取消且防重复 token 的分页器

### Affected areas

ecs

### 前置依赖

依赖 #5 的类型化操作与官方分页模型。

状态：GitHub issue #7（执行状态以 GitHub 为准）。

## 问题与证据

V1 #669 报告部分模型缺少 NextToken，应用需要可维护的分页支持。

来源：[上游证据](https://github.com/aliyun/alibaba-cloud-sdk-go/issues/669)。具体边界见 [调研](../research.md)。

## 范围

随 ECS 已实现操作提供官方支持的分页方式；不假定每个操作均有 NextToken。

## 验收条件

- [ ] 按 schema 区分 token 与 page-number 分页
- [ ] ctx 取消停止请求；重复 token 返回明确错误，避免无限循环
- [ ] 不修改原始输入；单页、多页、空页、错误和重复 token 均有离线测试
- [ ] 文档说明分页器并发/状态契约，外部 Example 展示遍历
- [ ] 请求次数、内存行为和服务端 page-size 限制写入说明

## 文档要求

每项公开 API 同时交付 doc.go、导出符号/字段注释和可执行外部 Example。
更新 README 支持范围与相关设计/发布说明。文档完成与代码完成属于同一交付。

## 验证计划

按验收条件建立离线行为测试，运行文档检查、vet 和相关测试一次。
涉及并发时由 Linux CI 运行 race；上线或 pkg.go.dev 页面由浏览器确认。
