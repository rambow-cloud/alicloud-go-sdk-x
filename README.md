# alicloud-go-sdk-x

[![Go Reference](https://pkg.go.dev/badge/github.com/rambow-cloud/alicloud-go-sdk-x.svg)](https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x)
[![CI](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/workflows/ci.yml/badge.svg)](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/workflows/ci.yml)

独立的阿里云 Go SDK 项目，目标是以 Go 原生 API、可预测请求行为和小依赖核心改善通用
OpenAPI 的使用体验。采用 **issue-driven development**，代码与 pkg.go.dev 文档一起交付。
本项目与阿里云官方 SDK 没有隶属关系。

**当前是初始化骨架，尚不能调用阿里云。** 已提供凭据 provider、可提取的服务错误契约、
离线示例和 CI 文档门禁；签名、HTTP 请求管线和 ECS 操作仍在路线图中。
API 在 v1.0.0 前可能变化；目前没有发布版本，文档 badge 是预期索引入口。

## 环境与约定

- Go **1.27+**，`go.mod` 最低版本为 1.27.0。
- JSON 直接使用标准库 **`encoding/json/v2`**，无需 GOEXPERIMENT。
- 运行时核心只有标准库依赖；后续阻塞 API 统一 `context.Context`。
- 原生 Go 注释、导出字段说明和可执行外部 Example 是完成条件。

## 离线使用示例

在本仓库运行 `go test ./...` 会验证包括下面用法在内的文档示例，无需云账号。
公开 module 路径为 `github.com/rambow-cloud/alicloud-go-sdk-x`；版本发布后再提供安装版本。

```go
package main

import (
    "context"
    "fmt"

    "github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
)

func main() {
    provider, err := credentials.NewStaticProvider(credentials.Credentials{
        AccessKeyID: "example-id", AccessKeySecret: "example-secret",
    })
    if err != nil {
        panic(err)
    }
    value, err := provider.Retrieve(context.Background())
    if err != nil {
        panic(err)
    }
    fmt.Println(value) // Credentials(<redacted>)
}
```

环境凭据使用 `credentials.EnvProvider{}`；读取 `ALIBABA_CLOUD_ACCESS_KEY_ID`、
`ALIBABA_CLOUD_ACCESS_KEY_SECRET` 和可选 `ALIBABA_CLOUD_SECURITY_TOKEN`。
目前只提供显式 provider，没有自动链、临时凭据缓存或刷新。
服务错误使用 `errors.As(err, &apiErr)` 提取 `*alicloud.APIError`。

## 调研与路线图

[调研报告](docs/research.md)记录 V1 生命周期，以及 V2 的 HTTP 配置、特定依赖组合、
生成注释和许可证问题报告，并说明证据边界。[设计](docs/design.md)据此定义改进方向。
OSS V2 已有 ctx-first 等良好设计；首版聚焦通用 OpenAPI 与 ECS 的一个只读操作。

| Issue | 交付 | 前置 |
| --- | --- | --- |
| [#1](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/1) | 标准库核心、凭据与错误契约 | 无，初始化交付 |
| [#2](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/2) | pkg.go.dev 文档、issue 流程与 CI | 无，初始化交付 |
| [#4](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/4) | 官方协议签名与编码向量 | 先固定协议 |
| [#3](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/3) | HTTP、context、endpoint、JSON v2 解码 | #4 |
| [#5](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/5) | ECS 首个只读操作 | #3、#4 |
| [#6](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/6) | 幂等性约束下的重试 | #3、#5 |
| [#9](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/9) | 临时凭据缓存与刷新 | 固定 STS/role 来源 |
| [#7](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/7) | 可取消的分页器 | #5 |
| [#8](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/8) | 可追溯模型生成与基准 | 先验证 #5 的 API 设计 |

Issue/label 结构参考 [rambow-cloud/powertools-lambda-go](https://github.com/rambow-cloud/powertools-lambda-go)。
类型、`module:*`、`priority:*` 与 `status:*` 标签的含义及自动化见
[维护约定](docs/issue-management.md)。GitHub issue 是执行状态的来源。

## 开发与文档

```sh
go run ./internal/cmd/doccheck
go vet ./...
go test ./...
node --test .github/scripts/issue-labels.test.cjs .github/scripts/issue-link.test.cjs
```

CI 在 Go 1.27 Linux/Windows 上检查格式、Go 文档与 JSON v2 导入策略、vet 和测试，
Linux 使用 race detector；自动分类脚本有独立测试。
doccheck 验证公开包概述、导出符号/字段注释和外部 Example 输出的存在性；
文字准确性与实际网页展示由评审和发布验收检查。

贡献前阅读 [CONTRIBUTING.md](CONTRIBUTING.md) 与 [AGENTS.md](AGENTS.md)。
提交新功能或缺陷时先建 issue；PR 完整交付使用 `Closes #N`，部分交付使用 `Refs #N`，
CI 会验证该 issue 确实存在。
发布及 pkg.go.dev 的浏览器验收步骤见 [releasing.md](docs/releasing.md)。

MIT License，见 [LICENSE](LICENSE)。
