package endpoint

import (
	"context"
	"errors"
	"maps"
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
	// Network selects a reviewed endpoint network; empty or public uses public rules.
	// Other combinations require an exact rule or an explicit BaseEndpoint.
	Network string
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
	// Network selects an exact network; empty is equivalent to public.
	Network string
	// Service is the product identifier.
	Service string
	// Region is the exact region identifier.
	Region string
	// URL is the HTTPS origin.
	URL string
}

// Rules is an immutable resolver; the zero value supports explicit overrides only.
type Rules struct {
	origins    map[[3]string]string
	deployment map[string]deploymentRule
}
type deploymentRule struct{ productCode, kind string }

var defaultEndpointRules []Rule
var defaultDeploymentRules map[string]deploymentRule

// NewRules copies validated rules and rejects duplicate or empty keys.
func NewRules(rules []Rule) (*Rules, error) {
	r := &Rules{origins: map[[3]string]string{}}
	for _, rule := range rules {
		if rule.Service == "" || rule.Region == "" {
			return nil, ErrUnsupported
		}
		if err := Validate(rule.URL); err != nil {
			return nil, err
		}
		network := rule.Network
		if network == "" {
			network = "public"
		}
		if !validLabel(network) {
			return nil, ErrUnsupported
		}
		key := [3]string{rule.Service, rule.Region, network}
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
		if r == nil {
			return Endpoint{}, ErrUnsupported
		}
		network := p.Network
		if network == "" {
			network = "public"
		}
		raw = r.origins[[3]string{p.Service, p.Region, network}]
		if raw == "" && network == "public" {
			if rule, ok := r.deployment[p.Service]; ok {
				switch rule.kind {
				case "regional":
					if validLabel(p.Region) {
						raw = "https://" + rule.productCode + "." + p.Region + ".aliyuncs.com"
					}
				case "global":
					raw = "https://" + rule.productCode + ".aliyuncs.com"
				}
			}
		}
	}
	if raw == "" {
		return Endpoint{}, ErrUnsupported
	}
	if err := Validate(raw); err != nil {
		return Endpoint{}, err
	}
	return Endpoint{URL: strings.TrimSuffix(raw, "/")}, nil
}

// DefaultResolver returns fresh rules generated from pinned official product DSL.
// Public mappings precede regional/global construction rules. A syntactically valid
// region does not prove service availability. Nonpublic combinations require a
// reviewed exact rule; explicit BaseEndpoint always takes precedence. No network I/O occurs.
func DefaultResolver() *Rules {
	r, _ := NewRules(defaultEndpointRules)
	r.deployment = maps.Clone(defaultDeploymentRules)
	return r
}

func validLabel(value string) bool {
	if len(value) == 0 || len(value) > 63 || value[0] == '-' || value[len(value)-1] == '-' {
		return false
	}
	for _, c := range value {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}
