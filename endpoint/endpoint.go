package endpoint

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// ErrUnsupported indicates a service/region absent from the rule set.
var ErrUnsupported = errors.New("endpoint: unsupported service or region")

// ErrInvalid indicates an endpoint that is not an absolute HTTPS origin.
var ErrInvalid = errors.New("endpoint: require HTTPS origin without credentials, path, query or fragment")

// Parameters identifies an endpoint request.
type Parameters struct {
	// Service is a lowercase product identifier, such as ecs or sts.
	Service string
	// Region is the Alibaba Cloud region identifier.
	Region string
	// BaseEndpoint overrides a rule with an explicit HTTPS origin when nonempty.
	BaseEndpoint string
}

// Endpoint describes the resolved origin.
type Endpoint struct {
	// URL is an absolute HTTPS origin without user information, query or fragment.
	URL string
}

// Resolver resolves an endpoint and must be safe for concurrent use.
type Resolver interface {
	// ResolveEndpoint selects a valid origin or returns an error.
	ResolveEndpoint(context.Context, Parameters) (Endpoint, error)
}

// ResolverFunc adapts a concurrent-safe function to Resolver.
type ResolverFunc func(context.Context, Parameters) (Endpoint, error)

// ResolveEndpoint invokes f. A nil function must not be used.
func (f ResolverFunc) ResolveEndpoint(ctx context.Context, p Parameters) (Endpoint, error) {
	return f(ctx, p)
}

// Validate checks that raw is a usable HTTPS origin. A trailing slash is allowed.
func Validate(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || (u.Path != "" && u.Path != "/") || strings.Contains(u.Host, "\\") {
		return ErrInvalid
	}
	return nil
}

// Rule maps one service and region to an origin.
type Rule struct {
	// Service is the product identifier.
	Service string
	// Region is the exact region identifier.
	Region string
	// URL is the HTTPS origin.
	URL string
}

// Rules is an immutable resolver; the zero value supports explicit overrides only.
type Rules struct{ origins map[[2]string]string }

// NewRules copies validated rules and rejects duplicate or empty keys.
func NewRules(rules []Rule) (*Rules, error) {
	r := &Rules{origins: map[[2]string]string{}}
	for _, rule := range rules {
		if rule.Service == "" || rule.Region == "" {
			return nil, ErrUnsupported
		}
		if err := Validate(rule.URL); err != nil {
			return nil, err
		}
		key := [2]string{rule.Service, rule.Region}
		if _, ok := r.origins[key]; ok {
			return nil, errors.New("endpoint: duplicate rule")
		}
		r.origins[key] = strings.TrimSuffix(rule.URL, "/")
	}
	return r, nil
}

// ResolveEndpoint prefers BaseEndpoint, then an exact rule. It never makes network calls.
func (r *Rules) ResolveEndpoint(ctx context.Context, p Parameters) (Endpoint, error) {
	if err := ctx.Err(); err != nil {
		return Endpoint{}, err
	}
	raw := p.BaseEndpoint
	if raw == "" {
		raw = r.origins[[2]string{p.Service, p.Region}]
	}
	if raw == "" {
		return Endpoint{}, ErrUnsupported
	}
	if err := Validate(raw); err != nil {
		return Endpoint{}, err
	}
	return Endpoint{URL: strings.TrimSuffix(raw, "/")}, nil
}

// DefaultResolver returns fresh ECS/STS public rules for cn-hangzhou, cn-shanghai,
// cn-beijing, cn-shenzhen and ap-southeast-1. VPC and other partitions require custom rules.
func DefaultResolver() *Rules {
	var rules []Rule
	for _, service := range []string{"ecs", "sts"} {
		for _, region := range []string{"cn-hangzhou", "cn-shanghai", "cn-beijing", "cn-shenzhen", "ap-southeast-1"} {
			rules = append(rules, Rule{Service: service, Region: region, URL: "https://" + service + "." + region + ".aliyuncs.com"})
		}
	}
	r, _ := NewRules(rules)
	return r
}
