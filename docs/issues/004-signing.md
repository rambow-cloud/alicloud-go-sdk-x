# 按官方协议实现签名与编码测试向量

### Affected areas

signing

### 前置依赖

先确定首个产品的官方协议、签名版本及固定测试向量，再开始实现。

状态：GitHub issue #4（执行状态以 GitHub 为准）。

## 问题与证据

新的独立 SDK 需要协议正确性证据，不能从旧 SDK 使用体验问题推导签名实现。

来源：[上游证据](https://github.com/aliyun/alibaba-cloud-sdk-go)。具体边界见 [调研](../research.md)。

## 范围

先明确首个产品的签名协议和 RPC/ROA 范围，在 internal/signing 实现；不混合协议。

## 验收条件

- [ ] issue 补充官方协议 URL、日期和首版支持范围后才实现
- [ ] 官方向量或独立人工推导 golden fixtures 验证 canonical request、string-to-sign、签名
- [ ] 覆盖 Unicode、空值、保留字符、重复参数、JSON/form、STS token
- [ ] 可注入时钟/nonce 进行确定性测试；不记录 key、签名或 raw body
- [ ] 与 transport 集成并验证每次请求独立签名，禁止覆盖调用者参数

## 文档要求

每项公开 API 同时交付 doc.go、导出符号/字段注释和可执行外部 Example。
更新 README 支持范围与相关设计/发布说明。文档完成与代码完成属于同一交付。

## 验证计划

按验收条件建立离线行为测试，运行文档检查、vet 和相关测试一次。
涉及并发时由 Linux CI 运行 race；上线或 pkg.go.dev 页面由浏览器确认。
