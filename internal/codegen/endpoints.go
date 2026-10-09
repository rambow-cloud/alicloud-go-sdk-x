package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/rambow-cloud/alicloud-go-sdk-x/endpoint"
)

type productEndpoints struct {
	Profile        string                    `json:"profile"`
	ProductCode    string                    `json:"productCode"`
	Rule           string                    `json:"rule"`
	Source         productSource             `json:"source"`
	ResolverSource productSource             `json:"resolverSource"`
	Overrides      []productEndpointOverride `json:"overrides"`
}
type productEndpointOverride struct {
	Region string        `json:"region"`
	URL    string        `json:"url"`
	Source productSource `json:"source"`
}
type endpointPolicy struct {
	Region   string   `json:"region"`
	Network  string   `json:"network"`
	URL      string   `json:"url"`
	Evidence []string `json:"evidence"`
}

var endpointLabel = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

func renderEndpoints(products []productIR) (map[string][]byte, error) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s// SPDX-License-Identifier: Apache-2.0\n// Copyright (c) 2009-present, Alibaba Cloud All rights reserved.\n// Modified: pinned DSL endpoint maps converted to Go; reviewed network rules added.\n// See LICENSE and NOTICE for source attribution.\npackage endpoint\nfunc init(){defaultEndpointRules=[]Rule{\n", productGenerated)
	var deployment bytes.Buffer
	deployment.WriteString("defaultDeploymentRules=map[string]deploymentRule{\n")
	seenProducts := map[string]bool{}
	for _, p := range products {
		ep := p.Endpoints
		if seenProducts[p.Product] || ep.Profile != "map-before-regional-v1" || !packageName.MatchString(ep.ProductCode) || (ep.Rule != "regional" && ep.Rule != "global") || ep.Source.File != "products/"+p.Product+"/main.tea" || ep.Source.Line < 1 || ep.ResolverSource.File != ep.Source.File || ep.ResolverSource.Line < 1 {
			return nil, errors.New("product: unsupported endpoint projection")
		}
		seenProducts[p.Product] = true
		fmt.Fprintf(&deployment, "%q:{productCode:%q,kind:%q},\n", p.Product, ep.ProductCode, ep.Rule)
		seen := map[string]bool{}
		overrides := slices.Clone(ep.Overrides)
		slices.SortFunc(overrides, func(a, b productEndpointOverride) int { return strings.Compare(a.Region, b.Region) })
		for _, rule := range overrides {
			if !endpointLabel.MatchString(rule.Region) || seen[rule.Region] || rule.Source.File != ep.Source.File || rule.Source.Line < 1 || endpoint.Validate(rule.URL) != nil {
				return nil, errors.New("product: invalid endpoint override")
			}
			seen[rule.Region] = true
			fmt.Fprintf(&b, "// Source: %s:%d.\n{Service:%q,Region:%q,URL:%q},\n", rule.Source.File, rule.Source.Line, p.Product, rule.Region, rule.URL)
		}
		if p.Policy != nil {
			networks := slices.Clone(p.Policy.Endpoints)
			slices.SortFunc(networks, func(a, b endpointPolicy) int {
				if v := strings.Compare(a.Network, b.Network); v != 0 {
					return v
				}
				return strings.Compare(a.Region, b.Region)
			})
			seen := map[string]bool{}
			for _, rule := range networks {
				key := rule.Network + "/" + rule.Region
				if rule.Network != "vpc" || !endpointLabel.MatchString(rule.Region) || seen[key] || endpoint.Validate(rule.URL) != nil || len(rule.Evidence) == 0 {
					return nil, errors.New("policy: invalid network endpoint")
				}
				seen[key] = true
				// This reviewed profile permits exact vpc origins only, never arbitrary suffixes.
				if rule.URL != "https://"+ep.ProductCode+"-vpc."+rule.Region+".aliyuncs.com" {
					return nil, errors.New("policy: network endpoint differs from reviewed profile")
				}
				for _, e := range rule.Evidence {
					if !strings.HasPrefix(e, "https://www.alibabacloud.com/help/") {
						return nil, errors.New("policy: network endpoint evidence must be official documentation")
					}
				}
				fmt.Fprintf(&b, "// Reviewed source: %s.\n{Service:%q,Region:%q,Network:%q,URL:%q},\n", rule.Evidence[0], p.Product, rule.Region, rule.Network, rule.URL)
			}
		}
	}
	b.WriteString("}\n")
	deployment.WriteString("}\n")
	b.Write(deployment.Bytes())
	b.WriteString("}\n")
	data, err := productFormat(&b)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{"endpoint/rules.gen.go": data, "endpoint/LICENSE": append([]byte(productGenerated+"\n"), productApacheLicense...), "endpoint/NOTICE": []byte(productGenerated + "\nCopyright (c) 2009-present, Alibaba Cloud All rights reserved.\nEndpoint tables derive from pinned official Alibaba Cloud product DSL in sources/darabonba.\nApache-2.0 applies to those tables; see LICENSE. Reviewed network rules cite official documentation.\nOriginal resolver, tooling and handwritten tests retain the project MIT license.\n\n阿里云上游版权如上。端点表来自 sources/darabonba 固定官方 DSL，按 Apache-2.0 许可；完整条款见 LICENSE。\n私网规则引用官方文档。原创解析器、工具和手写测试保留项目 MIT 许可。\n")}, nil
}

func (r *productRenderer) appendEndpointGuide(b *bytes.Buffer, chinese bool) {
	if r.p.Endpoints.Profile == "" {
		return
	}
	count := 0
	if r.p.Policy != nil {
		count = len(r.p.Policy.Endpoints)
	}
	if chinese {
		fmt.Fprintf(b, "## 端点\n\n- 公开端点使用官方 %s 规则和 %d 个准确映射；构造出的区域地址不代表服务已部署或可访问。\n- 已审核私网组合：%d 个；设置 Config.Network 为 vpc，不支持的组合直接报错。\n- BaseEndpoint 优先级最高；明确差异与限制见[端点规则](../endpoint-rules.zh-CN.md)。\n\n", r.p.Endpoints.Rule, len(r.p.Endpoints.Overrides), count)
	} else {
		fmt.Fprintf(b, "## Endpoints\n\n- Public endpoints use the official %s rule and %d exact mappings. A constructed regional address does not prove deployment or availability.\n- Reviewed private combinations: %d. Set Config.Network to vpc; unreviewed combinations fail.\n- BaseEndpoint takes precedence. See [endpoint rules](../endpoint-rules.md) for behavior and limits.\n\n", r.p.Endpoints.Rule, len(r.p.Endpoints.Overrides), count)
	}
}
