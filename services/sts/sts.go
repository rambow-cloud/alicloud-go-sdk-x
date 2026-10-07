package sts

import (
	"context"
	"encoding/json/v2"
	"errors"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Options aliases per-call runtime options.
type Options = alicloud.CallOptions

// Client is a concurrency-safe STS reference client. Construct with New.
type Client struct{ runtime *alicloud.Client }

// New validates and copies runtime configuration using the source credentials provider.
func New(config alicloud.Config) (*Client, error) {
	c, err := alicloud.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Client{runtime: c}, nil
}

// AssumeRoleInput describes token issuance; RoleARN and RoleSessionName are required.
type AssumeRoleInput struct {
	// RoleARN is the RAM role ARN, for example acs:ram::123456789012:role/example.
	RoleARN string
	// RoleSessionName is an audit session name of 2..64 ASCII characters.
	RoleSessionName string
	// DurationSeconds is the token lifetime; zero omits it (service default 3600).
	DurationSeconds int64
	// Policy is an optional restrictive JSON policy, limited to 2048 bytes.
	Policy string
	// ExternalID is an optional external identity used to prevent confused-deputy access.
	ExternalID string
	// SourceIdentity is an optional identity retained through role chaining.
	SourceIdentity string
}

var sessionPattern = regexp.MustCompile(`^[A-Za-z0-9.@_-]{2,64}$`)
var externalPattern = regexp.MustCompile(`^[A-Za-z0-9_+=,.@:/-]+$`)
var sourcePattern = regexp.MustCompile(`^[A-Za-z0-9_+=,.@-]{2,64}$`)

// Validate checks required fields, documented session syntax and policy JSON;
// the role's maximum session duration and authorization remain service decisions.
func (in AssumeRoleInput) Validate() error {
	if !strings.HasPrefix(in.RoleARN, "acs:ram::") || !strings.Contains(in.RoleARN, ":role/") || !sessionPattern.MatchString(in.RoleSessionName) {
		return errors.New("sts: invalid role ARN or session name")
	}
	if in.DurationSeconds != 0 && in.DurationSeconds < 900 {
		return errors.New("sts: duration must be at least 900 seconds")
	}
	if len(in.Policy) > 2048 {
		return errors.New("sts: policy exceeds 2048 bytes")
	}
	if in.Policy != "" {
		var policy map[string]any
		if err := json.Unmarshal([]byte(in.Policy), &policy); err != nil || policy == nil {
			return errors.New("sts: invalid policy JSON object")
		}
	}
	if in.ExternalID != "" && (len(in.ExternalID) < 2 || len(in.ExternalID) > 1224 || !externalPattern.MatchString(in.ExternalID)) {
		return errors.New("sts: invalid external ID")
	}
	if in.SourceIdentity != "" {
		if !sourcePattern.MatchString(in.SourceIdentity) || strings.HasPrefix(in.SourceIdentity, "acs:") || strings.HasPrefix(in.SourceIdentity, "aliyun:") || strings.HasPrefix(in.SourceIdentity, "alibabacloud:") {
			return errors.New("sts: invalid source identity")
		}
	}
	return nil
}

// RoleCredentials is a temporary credential response. Default formatting redacts it.
type RoleCredentials struct {
	// AccessKeyID is the temporary access key identifier.
	AccessKeyID string `json:"AccessKeyId"`
	// AccessKeySecret is the temporary signing secret and must never be logged.
	AccessKeySecret string `json:"AccessKeySecret"`
	// SecurityToken is the STS security token; no fixed length is assumed.
	SecurityToken string `json:"SecurityToken"`
	// ExpiresAt is the parsed RFC3339 expiration from the service.
	ExpiresAt time.Time `json:"Expiration"`
}

// String returns a redacted representation.
func (RoleCredentials) String() string { return "RoleCredentials(<redacted>)" }

// GoString returns a redacted representation for %#v formatting.
func (c RoleCredentials) GoString() string { return c.String() }

// AssumedRoleUser identifies the resulting audit principal.
type AssumedRoleUser struct {
	// ARN identifies the assumed role session.
	ARN string `json:"Arn"`
	// AssumedRoleID is the session principal identifier.
	AssumedRoleID string `json:"AssumedRoleId"`
}

// AssumeRoleOutput contains temporary credentials and identity metadata.
type AssumeRoleOutput struct {
	// Credentials contains sensitive temporary values.
	Credentials RoleCredentials `json:"Credentials"`
	// AssumedRoleUser identifies the assumed session.
	AssumedRoleUser AssumedRoleUser `json:"AssumedRoleUser"`
	// SourceIdentity is the service-returned chained identity.
	SourceIdentity string `json:"SourceIdentity"`
	// Metadata contains transport information and is excluded from wire decoding.
	Metadata alicloud.Metadata `json:"-"`
}

// AssumeRoleAPI is the minimal interface used by credential helpers and fakes.
type AssumeRoleAPI interface {
	// AssumeRole issues temporary role credentials using the client's source identity.
	AssumeRole(context.Context, *AssumeRoleInput, ...func(*Options)) (*AssumeRoleOutput, error)
}

// AssumeRole issues temporary credentials; nil or invalid input fails before transport.
// It is conservatively non-idempotent and Standard does not retry it.
func (c *Client) AssumeRole(ctx context.Context, input *AssumeRoleInput, opts ...func(*Options)) (*AssumeRoleOutput, error) {
	fail := func(err error) (*AssumeRoleOutput, error) {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, &alicloud.OperationError{Service: "sts", Operation: "AssumeRole", Err: err}
	}
	if input == nil {
		return fail(errors.New("sts: input required"))
	}
	in := *input
	if err := in.Validate(); err != nil {
		return fail(err)
	}
	q := url.Values{"RoleArn": {in.RoleARN}, "RoleSessionName": {in.RoleSessionName}}
	if in.DurationSeconds != 0 {
		q.Set("DurationSeconds", strconv.FormatInt(in.DurationSeconds, 10))
	}
	if in.Policy != "" {
		q.Set("Policy", in.Policy)
	}
	if in.ExternalID != "" {
		q.Set("ExternalId", in.ExternalID)
	}
	if in.SourceIdentity != "" {
		q.Set("SourceIdentity", in.SourceIdentity)
	}
	out := &AssumeRoleOutput{}
	meta, err := c.runtime.Invoke(ctx, alicloud.Operation{Service: "sts", Name: "AssumeRole", Version: "2015-04-01"}, alicloud.Request{Method: http.MethodPost, Query: q, Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}}}, out, opts...)
	if err != nil {
		return nil, err
	}
	out.Metadata = meta
	return out, nil
}
