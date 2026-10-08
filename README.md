# alicloud-go-sdk-x

[English](#english) | [中文](#中文)

## English

[Go Reference](https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x) ·
[CI](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/workflows/ci.yml) ·
[Development path](docs/development-path.md)

[Product acceptance](docs/product-acceptance.md) defines candidate Beta scope,
developer tasks and separate foundation/live/UX/release evidence; compilation and
generation counts alone are not a product acceptance result.

The initial generator route and review fix are integrated into `main`; see the paired
[review and integration record](docs/generator-integration.md) for commits, CI evidence
and the accepted scope.

An independent Alibaba Cloud SDK for Go. Requires Go 1.27 and encoding/json/v2.
Development is issue-driven, runtime-first, with English-primary Go docs and paired
English/Chinese guides. It is not an official SDK. APIs may change before v1.

The full-DSL [batch backend](docs/batch-go-emission.md) now emits 583 supported
operations in `service/ecs`, `service/vpc` and `service/sts`, with complete native
models, operation interfaces and offline Examples. Existing `services/` imports
remain the five-operation reference bridge described below. The [sparse policy backend](docs/capability-policy.md)
adds four native paginators, an ECS waiter and now nine reviewed operation policies (#37/#59).
Unsupported actions and separate acceptance
levels are recorded in [product guides](docs/products/ecs.md).

[Documentation automation](docs/product-documentation.md) extracts licensed official
descriptions into English Go comments and generates paired usage/source indexes with
per-symbol coverage. Missing source descriptions and Chinese semantic translations
are reported explicitly. Generated packages include Apache terms and source notices;
original runtime/tooling retain MIT. Executable Examples remain offline.

The eleven shared foundation capabilities were accepted with handwritten references and are now exercised by generated
ECS/STS/VPC clients. See the [supported-operation matrix](docs/support.md) and
[acceptance mapping](docs/foundation-acceptance.md). The `services/` bridge covers the documented
reference operations; no full-product or live-cloud acceptance is claimed. The first
[generator profile](docs/generator.md), extended under #24/#25, generates five operations, models, codecs, mock
interfaces, paginator/waiter adapters, Go docs, offline Examples and bilingual guides.
Benchmarks #20 remain separate. No version tag is published by this task.

The production frontend now uses pinned official Darabonba product DSL and the official
semantic parser, joined with public metadata and our reviewed policies. See
[migration and tools](docs/darabonba-migration.md) and [DSL/metadata decisions](docs/darabonba-decisions.md).

Foundation: shared middleware, endpoints, structured errors, bounded retry, credential
providers/cache/chain, STS helper, unified pagination/waiters, mock interfaces, testing
helpers and opt-in OpenTelemetry. Generator implementation follows foundation acceptance.
See docs/design.md, docs/research.md, docs/issue-management.md and docs/releasing.md.

Defaults: HTTPS, disabled redirects, no retries, a 30-second total operation deadline
and eight-MiB response limit. Core imports use only the standard library; telemetry is
optional. Default ECS/STS/VPC endpoint rules cover five reviewed public regions. APIs are early v0.

Prefer renewable STS role credentials with credentials.Cache for application clients.
Use `config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile("oss-sftp"))`
for native CLI Profile/OAuth configuration, cached refresh and STS exchange. Pass the
result to a generated NewFromConfig. Initial login uses `aliyun configure --mode OAuth`.
See [default configuration](docs/default-configuration.md) for precedence and supported
modes. Long-lived AK/SK require explicit provider opt-in; direct Config/Options never
accept bare keys or discover a fallback. Custom providers remain supported.

This complete STS-to-ECS example runs offline. Its explicitly constructed long-lived
source is used only by STS; ECS uses cached role credentials:

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

Output: two `cn-hangzhou` lines and `requests: 3` (one AssumeRole plus two ECS
requests sharing cached role credentials). All values and the 2099 expiry are synthetic.
Real calls require an explicitly configured authorized source and role, and a reviewed
endpoint; remove the scripted HTTP client. StaticProvider does not renew a copied token. Guides: [runtime](docs/runtime.md),
[credentials](docs/credentials.md), [cache](docs/credential-cache.md), [STS](docs/sts.md),
[VPC](docs/vpc.md), [retry](docs/retry.md), [pagination](docs/pagination.md), [waiters](docs/waiters.md),
[middleware](docs/middleware.md), [endpoints](docs/endpoints.md), [errors](docs/errors.md),
[testing](docs/testing.md), [telemetry](docs/telemetry.md).

Run go run ./internal/cmd/sdkgen check, go run ./internal/cmd/doccheck, node .github/scripts/check-doc-language.cjs,
go vet ./..., go test ./..., and node --test .github/scripts/\*.test.cjs.
Linux CI runs race detection; Windows CI verifies portability. Public packages provide
offline external Examples. Documentation checks cover structure; reviewers check semantics.
The checker also enforces JSON v2 and standard-library core dependencies.
Local documentation checks do not imply pkg.go.dev indexing.
Read AGENTS.md and CONTRIBUTING.md before contributing. MIT: see LICENSE.

## 中文

本地使用 `config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile("oss-sftp"))`
加载原生 CLI Profile/OAuth、缓存刷新与 STS 交换，返回值直接传给生成 NewFromConfig。
初次登录执行 `aliyun configure --mode OAuth`；优先级与模式见[默认配置](docs/default-configuration.md)。
长期 AK/SK 仍需显式启用 provider，直接 Config/Options 不接受裸密钥、不自行回退发现。

初始生成路线和评审修复已集成 `main`；双语[评审与集成记录](docs/generator-integration.md)
提供合并提交、CI 证据及已验收范围。

完整 DSL [批量后端](docs/batch-go-emission.md) 已在 `service/ecs`、`service/vpc`、
`service/sts` 输出 583 个支持操作、完整原生模型、操作小接口和离线 Example。
原 `services/` 导入仍为下文五操作参考桥；[稀疏策略后端](docs/capability-policy.md)
提供四个原生分页器、ECS waiter，现有九项已审核操作策略（#37/#59），
不支持操作与不同验收层次见[产品指南](docs/products/ecs.md)。

[文档自动化](docs/product-documentation.md) 将授权官方说明转换为英文 Go 注释，
生成对应双语使用/来源索引和逐符号覆盖；缺失说明与中文语义翻译明确报告。
生成包附 Apache 条款/来源通知，原创 runtime/工具保留 MIT，Example 仍离线运行。

生产前端使用固定官方 Darabonba 产品 DSL 与官方语义 parser，结合公共元数据及本项目
审核策略；见[迁移/工具](docs/darabonba-migration.md)及[DSL/元数据决策](docs/darabonba-decisions.md)。

[Go 文档](https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x) ·
[CI](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/workflows/ci.yml) ·
[开发路径](docs/development-path.md)

[产品验收](docs/product-acceptance.md)明确候选 Beta、开发者任务和独立的基础/真实/
体验/发布证据，编译及生成数量本身不表示产品验收通过。

独立阿里云 Go SDK，要求 Go 1.27 和直接使用 JSON v2。采用 issue 驱动、runtime 优先，
Go 注释英文为主，使用指南中英文对应。本项目非官方 SDK，v1 前 API 可能变化。

十一项共享基础能力已由手写参考验收，现由生成的 ECS/STS/VPC 客户端继续验证。参见[支持矩阵](docs/support.md)和
[验收映射](docs/foundation-acceptance.md)。原 `services/` 桥覆盖记载的参考操作，不宣称全产品或真实云验收。
[生成器](docs/generator.md) 经 #24/#25 扩展后生成五个操作、模型、编码、小 mock 接口、分页/waiter 适配器、Go 注释、
离线 Example 和双语指南；基准 #20 继续独立。
本次任务不发布版本标签。基础包括共享 middleware、endpoint、结构化错误、有界重试、凭据 provider/cache/chain、
STS helper、统一分页/waiter、mock 接口、测试辅助和可选 OpenTelemetry；基础验收后才建设 generator。
设计、调研、issue 维护和发布步骤见英文章节所列文档。

默认 HTTPS、禁用重定向、不重试、操作总期限 30 秒、每响应八 MiB。核心导入仅标准库，
telemetry 可选。默认 ECS/STS/VPC 端点覆盖五个核实的公网地域。API 属于早期 v0。

应用客户端优先采用可刷新的 STS role provider 与 credentials.Cache。长期 AK/SK
必须显式注册 StaticProvider/EnvProvider；Config/Options 不接受裸密钥、不发现回退来源，仍支持
自定义 provider，见[凭据契约](docs/credentials.md)。

以下完整 STS→ECS 示例无需网络；显式构造的长期来源仅用于 STS，ECS 使用缓存的角色凭据：

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

输出两行 `cn-hangzhou` 和 `requests: 3`（一次 AssumeRole 与两次复用角色缓存的 ECS 请求）。
所有值及 2099 过期时间均为合成数据；真实调用需要显式配置授权来源、角色与核实端点，并移除
脚本 HTTP 客户端。StaticProvider 不会刷新复制的 token。
使用指南：[运行时](docs/runtime.md)、[凭据](docs/credentials.md)、[缓存](docs/credential-cache.md)、
[STS](docs/sts.md)、[VPC](docs/vpc.md)、[重试](docs/retry.md)、[分页](docs/pagination.md)、[waiter](docs/waiters.md)、
[middleware](docs/middleware.md)、[端点](docs/endpoints.md)、[错误](docs/errors.md)、
[测试](docs/testing.md)、[telemetry](docs/telemetry.md)。

运行英文章节的文档、双语、vet、Go 与自动化测试命令；Linux CI 使用 race，Windows 验证可移植性。
公共包提供离线外部 Examples；检查工具验证结构，评审检查语义。贡献前阅读约束与贡献指南，MIT 许可证见 LICENSE。
检查器也约束 JSON v2 和核心标准库依赖。本地文档检查不代表 pkg.go.dev 已索引。

## First-release status / 首版状态

### English

All four pinned STS actions now generate. [v0.1.0 notes](docs/releases/v0.1.0.md) and the [release checklist](docs/sts-v010-release-checklist.md) are prepared; independent docs-only acceptance and publication/indexing remain pending. The [account-free kit](examples/stsacceptance/README.md) supplies the pinned official-v2 comparison. Do not interpret the documentation link as a published tag or broader Beta.

### 中文

四个固定 STS 操作均已生成，[v0.1.0 说明](docs/releases/v0.1.0.md)及[发布清单](docs/sts-v010-release-checklist.md)已准备；独立文档验收及发布/索引仍待完成。[无账号验收包](examples/stsacceptance/README.md)提供固定官方 v2 对比，文档链接不表示标签已发布或更广 Beta。
