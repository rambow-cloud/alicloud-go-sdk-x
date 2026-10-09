package externalcreds

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/rpcmodel"
)

// Options controls bounded external retrieval. Scalar options are copied.
type Options struct {
	// HTTPClient is shared and must honor context and prohibit credential redirects.
	// Nil uses a private client. Concrete http.Client values are copied with redirects disabled.
	HTTPClient alicloud.HTTPClient
	// Timeout bounds an entire retrieval; zero defaults to five seconds.
	Timeout time.Duration
	// MaxResponseBytes bounds response/process output; zero defaults to one MiB.
	MaxResponseBytes int64
}
type settings struct {
	client  alicloud.HTTPClient
	timeout time.Duration
	limit   int64
}

func configure(o Options, metadata bool) (settings, error) {
	if o.Timeout < 0 || o.MaxResponseBytes < 0 || (o.HTTPClient != nil && rpcmodel.IsNil(o.HTTPClient)) {
		return settings{}, errors.New("externalcreds: invalid retrieval options")
	}
	if o.Timeout == 0 {
		o.Timeout = 5 * time.Second
	}
	if o.MaxResponseBytes == 0 {
		o.MaxResponseBytes = 1 << 20
	}
	if o.MaxResponseBytes > 16<<20 {
		return settings{}, errors.New("externalcreds: excessive response bound")
	}
	client := o.HTTPClient
	if client == nil {
		tr := &http.Transport{}
		if base, ok := http.DefaultTransport.(*http.Transport); ok {
			tr = base.Clone()
		}
		if metadata {
			tr.Proxy = nil
		}
		client = &http.Client{Transport: tr, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	} else if c, ok := client.(*http.Client); ok {
		copy := *c
		copy.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		client = &copy
	}
	if metadata {
		if c, ok := client.(*http.Client); ok {
			copy := *c
			tr := copy.Transport
			if tr == nil {
				tr = http.DefaultTransport
			}
			if native, ok := tr.(*http.Transport); ok {
				owned := native.Clone()
				owned.Proxy = nil
				copy.Transport = owned
			}
			client = &copy
		}
	}
	return settings{client: client, timeout: o.Timeout, limit: o.MaxResponseBytes}, nil
}

type sourceError struct{ cause error }

func (e *sourceError) Error() string { return "externalcreds: external request failed" }
func (e *sourceError) Unwrap() error { return e.cause }
func (s settings) request(ctx context.Context, method, raw string, headers http.Header, body []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r, err := http.NewRequestWithContext(ctx, method, raw, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("externalcreds: invalid request")
	}
	for k, vs := range headers {
		for _, v := range vs {
			r.Header.Add(k, v)
		}
	}
	res, err := s.client.Do(r)
	if err != nil {
		if res != nil && res.Body != nil {
			res.Body.Close()
		}
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		return nil, &sourceError{cause: err}
	}
	if res == nil || res.Body == nil {
		return nil, errors.New("externalcreds: missing response")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("externalcreds: credential HTTP status rejected")
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, s.limit+1))
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if err != nil {
		return nil, &sourceError{cause: err}
	}
	if int64(len(b)) > s.limit {
		return nil, errors.New("externalcreds: response exceeds bound")
	}
	return b, nil
}

type wireCredentials struct {
	Code            string `json:"Code"`
	AccessKeyID     string `json:"AccessKeyId"`
	AccessKeySecret string `json:"AccessKeySecret"`
	SecurityToken   string `json:"SecurityToken"`
	Expiration      string `json:"Expiration"`
	ExpirationInt64 int64  `json:"ExpirationInt64"`
}

func (w wireCredentials) snapshot(source string) (credentials.Credentials, error) {
	if w.Code != "" && w.Code != "Success" {
		return credentials.Credentials{}, errors.New("externalcreds: credential issuance rejected")
	}
	if strings.TrimSpace(w.AccessKeyID) == "" || strings.TrimSpace(w.AccessKeySecret) == "" || strings.TrimSpace(w.SecurityToken) == "" {
		return credentials.Credentials{}, credentials.ErrMissingCredentials
	}
	var expires time.Time
	var err error
	if w.Expiration != "" {
		expires, err = time.Parse(time.RFC3339, w.Expiration)
		if err != nil {
			return credentials.Credentials{}, errors.New("externalcreds: invalid credential expiration")
		}
		if w.ExpirationInt64 != 0 && expires.Unix() != w.ExpirationInt64 {
			return credentials.Credentials{}, errors.New("externalcreds: inconsistent expiration")
		}
	} else if w.ExpirationInt64 > 0 {
		expires = time.Unix(w.ExpirationInt64, 0)
	}
	if expires.IsZero() || !time.Now().Before(expires) {
		return credentials.Credentials{}, credentials.ErrExpired
	}
	return credentials.Credentials{AccessKeyID: w.AccessKeyID, AccessKeySecret: w.AccessKeySecret, SecurityToken: w.SecurityToken, ExpiresAt: expires.UTC(), Source: source}, nil
}

// URIProvider retrieves temporary credentials from an explicitly configured HTTP(S) URI.
// Construct with NewURIProvider. Concurrent retrieval is safe; wrap in credentials.Cache.
// URLs may contain private query parameters and are never included in diagnostics.
type URIProvider struct {
	uri      string
	settings settings
}

// NewURIProvider validates an explicit URI and bounds without retrieving credentials.
// HTTP and HTTPS are accepted for compatibility with local/private credential brokers.
// Userinfo, fragments and empty hosts fail. Configure only a trusted broker URI.
func NewURIProvider(uri string, options Options) (*URIProvider, error) {
	u, err := url.Parse(uri)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" {
		return nil, errors.New("externalcreds: invalid credential URI")
	}
	s, err := configure(options, false)
	if err != nil {
		return nil, err
	}
	return &URIProvider{uri: uri, settings: s}, nil
}

// String returns a redacted representation without the credential URI.
func (*URIProvider) String() string { return "URIProvider(<redacted>)" }

// GoString returns a redacted representation for Go-syntax formatting.
func (p *URIProvider) GoString() string { return p.String() }

// Retrieve performs one bounded GET and validates complete future-expiring credentials.
// No redirect, cache or retry occurs. Cancellation remains inspectable.
func (p *URIProvider) Retrieve(ctx context.Context) (credentials.Credentials, error) {
	if err := ctx.Err(); err != nil {
		return credentials.Credentials{}, err
	}
	if p == nil || p.uri == "" {
		return credentials.Credentials{}, errors.New("externalcreds: unconfigured URI provider")
	}
	ctx, cancel := context.WithTimeout(ctx, p.settings.timeout)
	defer cancel()
	b, err := p.settings.request(ctx, http.MethodGet, p.uri, nil, nil)
	if err != nil {
		return credentials.Credentials{}, err
	}
	var w wireCredentials
	if json.Unmarshal(b, &w) != nil {
		return credentials.Credentials{}, errors.New("externalcreds: invalid credential JSON")
	}
	return w.snapshot("CredentialsURI")
}
