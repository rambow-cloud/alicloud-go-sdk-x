# 初始化标准库核心、凭据与错误契约

### Affected areas

core, credentials

### 前置依赖

无；初始化交付。

状态：GitHub issue #1（执行状态以 GitHub 为准）。

## 问题与证据

V2 #177 报告特定 Tea/OpenAPI 版本组合无法编译；初始化先固定小核心和可测试的凭据、错误契约。

来源：[上游证据](https://github.com/aliyun/darabonba-openapi/issues/177)。具体边界见 [调研](../research.md)。

## 范围

根包 APIError；credentials.Provider、静态和环境 provider；Go module、LICENSE、仓库约定。不实现服务 API、自动凭据链或临时凭据刷新。

## 验收条件

- [ ] module 为 github.com/rambow-cloud/alicloud-go-sdk-x，Go 1.27，核心只有标准库依赖
- [ ] Retrieve(ctx) 保留取消错误；不完整凭据返回可 errors.Is 判断的错误
- [ ] 静态 provider 不修改凭据；环境 provider 使用约定的三个变量且不缓存
- [ ] APIError 可 errors.As 提取 Code/RequestID/HTTPStatusCode；默认格式不含 Message
- [ ] 完成凭据取消、缺失、环境、脱敏与错误包装的离线测试

## 文档要求

每项公开 API 同时交付 doc.go、导出符号/字段注释和可执行外部 Example。
更新 README 支持范围与相关设计/发布说明。文档完成与代码完成属于同一交付。

## 验证计划

按验收条件建立离线行为测试，运行文档检查、vet 和相关测试一次。
涉及并发时由 Linux CI 运行 race；上线或 pkg.go.dev 页面由浏览器确认。
