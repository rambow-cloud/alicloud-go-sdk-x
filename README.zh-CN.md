# alicloud-go-sdk-x

[English](README.md)

- [RPC DSL 扩展](docs/dsl-rpc-expansion.zh-CN.md)：固定来源下生成 ECS 380/380、VPC 403/403、STS 4/4。生成及离线检查通过不表示所有操作均经过真实调用验收。

- [完整 DSL 真实调用证据](docs/product-live-validation.zh-CN.md): 完整 DSL ECS/VPC 的历史只读证据（#47）：镜像读取两页，实例和 VPC 返回空页，真实 waiter 检查跳过；#74/#75 产品验收仍单独进行。

- 首版路线：#60 代理 STS 验收 → ECS #74 → VPC #75 → #61 发布。独立人工体验另由可选 #76 跟踪，详见[开发路线](docs/sts-ecs-vpc-path.zh-CN.md)。

- 独立开发的阿里云 Go SDK，不是阿里云官方项目。

- 要求 Go 1.27，并直接使用 `encoding/json/v2`。v1 之前 API 可能调整。

- [Go API 文档](https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x) · [CI](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/workflows/ci.yml) · [开发路线](docs/development-path.zh-CN.md)

## 快速开始

- 本地使用前，执行一次 `aliyun configure --mode OAuth --profile oss-sftp` 完成登录。

- 调用 `config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile("oss-sftp"))`，再将配置传给生成客户端的 `NewFromConfig`。

- [默认配置指南](docs/default-configuration.zh-CN.md)提供完整程序，并说明来源优先级、支持的模式和 OAuth 续期。

- 应用优先使用可续期的 STS 角色凭据和 Cache。长期密钥必须显式启用；Config/Options 只接受凭据提供者，不接受裸密钥，也支持自定义来源。

- 默认使用 HTTPS，禁用重定向，不重试；操作超时为 30 秒，单个响应最多 8 MiB。核心只依赖标准库，OpenTelemetry 可按需启用。

- 端点遵循固定的官方端点表、地域规则及已审核私网组合。显式 BaseEndpoint 优先；能构造 URL 不代表网络可达。详见[端点规则](docs/endpoint-rules.zh-CN.md)和[支持范围](docs/support.zh-CN.md)。

## 支持范围

- [OSS 运行时](docs/oss-runtime.zh-CN.md)支持显式 OSS4、桶域名寻址及类型化 XML codec；生产 service/oss 生成和真实云验收仍待完成。
- [OSS 语义 IR](docs/oss-semantic-ir.zh-CN.md)发现完整来源中的 90 个操作，并降低了 16 个 XML 读取操作。这是前端进展，尚未交付公开 OSS 客户端。

- 完整官方 DSL 生成 STS 4、ECS 380、VPC 403 个操作，提供类型化模型、小型测试替身接口和离线 Example。FC 另提供 [ROA 离线预览](docs/fc-binary-generation.zh-CN.md)：生成全部 73 个操作，包括返回由调用者持有的响应流的二进制 InvokeFunction。生成、编译、消费者验收及真实调用分别记录。

- 已审核策略提供 ECS 14/VPC 15 个分页器、ECS 2/VPC 3 个 waiter。ECS 16/380、VPC 15/403 个操作有已审核策略；其余操作仍未审核。生成数量不代表真实调用已经验收。

- 客户端统一使用 `service/`，旧 `services/` 兼容桥已移除，见[迁移说明](docs/service-consolidation.zh-CN.md)。

- 生产流程为：固定官方 Darabonba DSL → 官方语义解析器 → IR（中间表示）→ 本项目 Go 后端和公共运行时。元数据只作补充，策略用于添加已审核的行为。

- 支持情况和限制见[产品指南](docs/products/ecs.zh-CN.md)、[能力策略](docs/capability-policy.zh-CN.md)、[文档生成](docs/product-documentation.zh-CN.md)和[来源决策](docs/darabonba-decisions.zh-CN.md)。

- 十一项基础能力和初始生成路线已验收，见[基础证据](docs/foundation-acceptance.zh-CN.md)与[集成记录](docs/generator-integration.zh-CN.md)。

- [产品验收](docs/product-acceptance.zh-CN.md)分别记录离线、真实调用、用户体验和发布证据。v0.1.0 按 STS 代理验收（#60）、ECS（#74）、VPC（#75）、发布（#61）的顺序推进；独立人工体验为可选 #76，基准测试 #20 单独跟踪。

## 离线 STS→ECS 示例

- 以下完整程序使用模拟 HTTP 和虚构凭据，不需要账号或网络。

- 显式配置的来源只用于签名 STS 请求，ECS 使用缓存的角色凭据。

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

func main() {
	ctx := context.Background()
	// Long-lived keys are an explicit bootstrap choice, used only by STS here.
	source, err := credentials.NewStaticProvider(credentials.Credentials{
		AccessKeyID: "synthetic-source", AccessKeySecret: "synthetic-secret",
	})
	if err != nil {
		panic(err)
	}
	checkRole := func(r *http.Request) error {
		if !strings.Contains(r.Header.Get("Authorization"), "Credential=synthetic-role,") || r.Header.Get("X-Acs-Security-Token") != "synthetic-token" {
			return errors.New("expected role credentials")
		}
		return nil
	}
	transport := sdktest.NewTransport(
		sdktest.Step{Body: `{"Credentials":{"AccessKeyId":"synthetic-role","AccessKeySecret":"synthetic-secret","SecurityToken":"synthetic-token","Expiration":"2099-01-01T00:00:00Z"}}`},
		sdktest.Step{Body: `{"Regions":{"Region":[{"RegionId":"cn-hangzhou"}]}}`, Check: checkRole},
		sdktest.Step{Body: `{"Regions":{"Region":[{"RegionId":"cn-hangzhou"}]}}`, Check: checkRole},
	)
	httpClient := &http.Client{Transport: transport}
	api, err := sts.NewFromConfig(alicloud.Config{
		Region: "cn-hangzhou", CredentialsProvider: source, HTTPClient: httpClient,
	})
	if err != nil {
		panic(err)
	}
	roleARN, session := "acs:ram::123456789012:role/example", "application"
	provider, err := stscreds.NewAssumeRoleProviderFromClient(api, sts.AssumeRoleInput{
		RoleARN: &roleARN, RoleSessionName: &session,
	})
	if err != nil {
		panic(err)
	}
	cache, err := credentials.NewCache(provider, credentials.CacheOptions{})
	if err != nil {
		panic(err)
	}
	client, err := ecs.NewFromConfig(alicloud.Config{
		Region: "cn-hangzhou", CredentialsProvider: cache, HTTPClient: httpClient,
	})
	if err != nil {
		panic(err)
	}
	for range 2 {
		output, err := client.DescribeRegions(ctx, nil)
		if err != nil {
			panic(err)
		}
		fmt.Println(*output.Regions.Region[0].RegionID)
	}
	fmt.Println("requests:", transport.Calls())
}
```

- 预期输出：两行 `cn-hangzhou` 和一行 `requests: 3`，分别来自一次 AssumeRole 和两次 ECS 请求。2099 年的到期时间是测试值。

- 改为真实调用时，使用已授权的来源、角色及核实的端点，并移除模拟 HTTP 客户端。StaticProvider 不会自动续期复制进来的令牌。

## 使用指南

- 运行时：[HTTP](docs/runtime.zh-CN.md)、[中间件](docs/middleware.zh-CN.md)、[端点](docs/endpoints.zh-CN.md)、[错误](docs/errors.zh-CN.md)。

- 凭据：[提供者](docs/credentials.zh-CN.md)、[缓存](docs/credential-cache.zh-CN.md)、[STS](docs/sts.zh-CN.md)。

- 行为：[VPC](docs/vpc.zh-CN.md)、[重试](docs/retry.zh-CN.md)、[分页](docs/pagination.zh-CN.md)、[状态等待](docs/waiters.zh-CN.md)、[测试](docs/testing.zh-CN.md)、[链路追踪](docs/telemetry.zh-CN.md)。

- 项目：[设计](docs/design.zh-CN.md)、[调研](docs/research.zh-CN.md)、[issue 管理](docs/issue-management.zh-CN.md)、[发布流程](docs/releasing.zh-CN.md)。

## 参与开发与验证

- 先阅读 [AGENTS.md](AGENTS.zh-CN.md) 和[贡献指南](CONTRIBUTING.zh-CN.md)。通过 issue 跟踪工作，issue/PR 只写英文。

- 中英文指南分别保存并互相链接，Go 注释使用英文；详见[写作规范](docs/documentation-style.zh-CN.md)。

- 有意义的修改后，执行下面的检查一次。Linux CI 增加竞态检查，Windows CI 验证可移植性。

```powershell
go run ./internal/cmd/sdkgen product-check
go run ./internal/cmd/doccheck
node .github/scripts/check-doc-language.cjs
go vet ./...
go test ./...
node --test .github/scripts/*.test.cjs
```

- doccheck 检查 Go 文档、JSON v2 和核心标准库依赖；语言检查验证文件配对和链接，语义由评审确认。本地检查通过不代表 pkg.go.dev 已完成索引。

- 原创运行时和工具采用 MIT（[LICENSE](LICENSE)）；上游派生的生成包保留 Apache 条款和来源通知。本次文档修改不创建发布标签。
