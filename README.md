# alicloud-go-sdk-x

[中文](README.zh-CN.md)

- Delivery: #60 agent STS acceptance -> ECS #74 -> VPC #75 -> #61 publication. Independent human UX is optional #76; see [the route](docs/sts-ecs-vpc-path.md).

- An independent Alibaba Cloud SDK for Go.
- This is not an official Alibaba Cloud project.

- Requires Go 1.27 and `encoding/json/v2`.
- APIs may change before v1.

- [Go reference](https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x) · [CI](https://github.com/rambow-cloud/alicloud-go-sdk-x/actions/workflows/ci.yml) · [Development path](docs/development-path.md)

## Start here

- For local use, run `aliyun configure --mode OAuth --profile oss-sftp` once.

- Load with `config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile("oss-sftp"))`.
- Pass the result to a generated `NewFromConfig`.

- See [default configuration](docs/default-configuration.md) for a complete program, source order, supported modes and OAuth renewal.

- Prefer renewable STS role credentials with Cache.
- Long-lived keys require explicit provider opt-in; Config/Options accept providers, not raw keys.
- Custom providers work.

- Defaults: HTTPS, redirects disabled, no retry, 30-second operation timeout, 8 MiB response limit.
- The core uses only the standard library; OpenTelemetry is optional.

- Default endpoint rules cover five reviewed public ECS/STS/VPC regions.
- See [support](docs/support.md).

## Coverage

- The full-DSL generator emits 583 actions in `service/ecs`, `service/vpc` and `service/sts`, with complete models, small mock interfaces and offline Examples.

- Reviewed policies add four paginators, one ECS waiter and nine operation policies.
- Generation counts do not prove live acceptance.

- `services/` keeps the older five-operation reference bridge.
- See [migration](docs/batch-go-emission.md) before changing imports.

- The production path is pinned official Darabonba DSL → official semantic parser → IR → our Go backend and shared runtime.
- Metadata is optional enrichment; policies add reviewed behavior.

- [Product guides](docs/products/ecs.md), [capability policies](docs/capability-policy.md), [documentation generation](docs/product-documentation.md) and [source decisions](docs/darabonba-decisions.md) record support and limits.

- The eleven foundation capabilities and initial generator route are accepted; see [foundation evidence](docs/foundation-acceptance.md) and [integration](docs/generator-integration.md).

- [Product acceptance](docs/product-acceptance.md) separates offline, live, user and release evidence. v0.1.0 follows STS agent acceptance (#60), ECS (#74), VPC (#75), then publication (#61). Independent human UX is optional follow-up #76. Benchmarks #20 are separate.

## Offline STS-to-ECS example

- This complete program uses scripted HTTP and synthetic credentials.
- It needs no account or network.

- The explicit source signs only STS.
- ECS uses cached role credentials.

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

- Expected output: two `cn-hangzhou` lines and `requests: 3` (one AssumeRole and two ECS requests).
- The 2099 expiry is synthetic.

- For real calls, use an authorized source/role and a reviewed endpoint, and remove the scripted HTTP client.
- StaticProvider does not renew copied tokens.

## Guides

- Runtime: [HTTP](docs/runtime.md), [middleware](docs/middleware.md), [endpoints](docs/endpoints.md), [errors](docs/errors.md).

- Credentials: [providers](docs/credentials.md), [cache](docs/credential-cache.md), [STS](docs/sts.md).

- Behavior: [VPC](docs/vpc.md), [retry](docs/retry.md), [pagination](docs/pagination.md), [waiters](docs/waiters.md), [testing](docs/testing.md), [telemetry](docs/telemetry.md).

- Project: [design](docs/design.md), [research](docs/research.md), [issue management](docs/issue-management.md), [release steps](docs/releasing.md).

## Contribute and verify

- Read [AGENTS.md](AGENTS.md) and [CONTRIBUTING.md](CONTRIBUTING.md).
- Use issues and English-only issue/PR text.

- Write separate, linked English and Chinese guides.
- Keep Go docs in English.
- See [writing rules](docs/documentation-style.md).

- Run the checks below once per meaningful change.
- Linux CI adds race detection; Windows CI checks portability.

```powershell
go run ./internal/cmd/sdkgen check
go run ./internal/cmd/sdkgen product-check
go run ./internal/cmd/doccheck
node .github/scripts/check-doc-language.cjs
go vet ./...
go test ./...
node --test .github/scripts/*.test.cjs
```

- Doccheck covers Go docs, JSON v2 and standard-library core imports.
- Language checks cover file pairs and links; reviewers check meaning.
- Local success does not prove pkg.go.dev indexing.

- Original runtime/tooling use MIT ([LICENSE](LICENSE)); upstream-derived generated packages retain Apache terms and notices.
- No release tag is created by this documentation change.
