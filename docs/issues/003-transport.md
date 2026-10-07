# 实现 ctx-first HTTP 请求管线与 endpoint 注入

### Affected areas

transport, core

### 前置依赖

依赖 #4 的首版签名协议。

状态：GitHub issue #3（执行状态以 GitHub 为准）。

## 问题与证据

V2 #72 和 V1 #612 反映 HTTP 客户端与 endpoint 自定义需求。

来源：[上游证据](https://github.com/aliyun/alibabacloud-go-sdk/issues/72)。具体边界见 [调研](../research.md)。

## 范围

标准 net/http 注入、endpoint resolver、deadline、响应限制、可提取 APIError；默认不重试。

## 验收条件

- [ ] 凭据、resolver 和 HTTP 传播 context，errors.Is 可识别取消/超时
- [ ] time.Duration 配置；默认总期限固定并测试用户更短 deadline
- [ ] 不修改调用者请求、HTTP client 或共享配置；race 验证并发契约
- [ ] HTTPS 校验及跨 host 重定向保护；自定义 endpoint 不猜测地域域名
- [ ] 限制读取响应体、关闭 body，解析成功/错误/非 JSON 错误响应并保留 request ID
- [ ] 直接使用 encoding/json/v2；验证重复字段、非法 UTF-8 与未知字段行为

## 文档要求

每项公开 API 同时交付 doc.go、导出符号/字段注释和可执行外部 Example。
更新 README 支持范围与相关设计/发布说明。文档完成与代码完成属于同一交付。

## 验证计划

按验收条件建立离线行为测试，运行文档检查、vet 和相关测试一次。
涉及并发时由 Linux CI 运行 race；上线或 pkg.go.dev 页面由浏览器确认。
