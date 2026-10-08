package alicloud

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/endpoint"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/signing"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
)

// HTTPClient is the transport seam. Implementations must be concurrency safe,
// honor request context, and never follow redirects with signed credentials.
type HTTPClient interface {
	// Do sends a request and returns a response whose body the SDK will close.
	Do(*http.Request) (*http.Response, error)
}

// Config configures a client. NewClient copies fields and middleware registrations;
// providers, resolvers, transports, retry policies and hooks remain shared.
type Config struct {
	// Region is the default operation region; individual operations may override it.
	Region string
	// CredentialsProvider is required and must honor the Provider contract.
	// Prefer config.LoadDefaultConfig with native temporary/Profile/OAuth sources
	// or a cached STS role provider. Long-lived sources require explicit opt-in;
	// nil and typed-nil providers are rejected.
	// Construction never retrieves credentials or discovers a fallback source.
	// AnonymousProvider is an explicit marker for reviewed anonymous operations;
	// it cannot supply credentials for signed operations.
	CredentialsProvider credentials.Provider
	// HTTPClient is optional; nil uses a private http.Client with redirects disabled.
	HTTPClient HTTPClient
	// EndpointResolver is optional; nil uses endpoint.DefaultResolver.
	EndpointResolver endpoint.Resolver
	// BaseEndpoint is an optional explicit HTTPS origin.
	BaseEndpoint string
	// Retryer is optional; nil means no retries.
	Retryer retry.Retryer
	// Middleware registers shared interceptors in execution order.
	Middleware []middleware.Registration
	// Timeout bounds the whole operation; zero defaults to thirty seconds.
	Timeout time.Duration
	// MaxResponseBytes bounds each response; zero defaults to eight MiB.
	MaxResponseBytes int64
	// Sleep is the backoff seam; nil uses retry.Wait. It must honor context cancellation.
	Sleep func(context.Context, time.Duration) error
}

// CallOptions overrides client defaults for one invocation. Middleware is appended
// to client registrations; callbacks must not retain this options value.
type CallOptions struct {
	// Config replaces the base client configuration for this call when non-nil.
	// It is copied and validated; Middleware below appends to its registrations.
	// Providers and other extension objects remain shared and concurrency safe.
	Config *Config
	// Region overrides the operation's region and its RegionId query parameter.
	Region string
	// BaseEndpoint overrides endpoint resolution with an explicit HTTPS origin.
	BaseEndpoint string
	// Timeout overrides the total timeout when nonzero.
	Timeout time.Duration
	// Middleware appends operation-specific interceptors.
	Middleware []middleware.Registration
}

// AuthenticationMode identifies the reviewed request authentication protocol.
// The zero value uses ACS3 signing; unsupported values fail before transport.
type AuthenticationMode uint8

const (
	// AuthenticationACS3 signs each request with explicit source credentials.
	AuthenticationACS3 AuthenticationMode = iota
	// AuthenticationAnonymousRPC uses unsigned legacy RPC query framing over HTTPS.
	// It never retrieves a provider or adds source authentication fields.
	AuthenticationAnonymousRPC
)

// Operation describes a reviewed API operation; clients provide concrete models.
type Operation struct {
	// Service identifies the product, such as ecs or sts.
	Service string
	// Name is the x-acs-action value.
	Name string
	// Version is the x-acs-version value.
	Version string
	// Idempotent explicitly permits retry when the policy also permits it.
	Idempotent bool
	// Authentication selects the reviewed protocol; zero means signed ACS3.
	// Generated clients set anonymous RPC only from approved official DSL evidence.
	Authentication AuthenticationMode
}

// Request supplies pre-encoded RPC/ROA wire data. Invoke copies maps and bytes.
type Request struct {
	// Method is the uppercase HTTP method; empty defaults to POST.
	Method string
	// Path is the decoded absolute resource path; empty defaults to /.
	Path string
	// Region overrides Config.Region before CallOptions.Region is applied.
	Region string
	// Query contains already-flattened service parameters.
	Query url.Values
	// Header contains additional headers; signing headers are managed by the runtime.
	Header http.Header
	// Body contains exact wire bytes and is limited to eight MiB, including hook changes.
	Body []byte
}

// Client is a concurrency-safe immutable runtime. Construct with NewClient;
// the zero value must not be used. It supports ACS3 JSON OpenAPI operations only.
type Client struct {
	config Config
	stack  *middleware.Stack
}

// NewClient validates configuration. Supplied *http.Client values are shallow
// copied and redirects are disabled without modifying the caller's client.
func NewClient(config Config) (*Client, error) {
	provider := reflect.ValueOf(config.CredentialsProvider)
	missingProvider := !provider.IsValid()
	if provider.IsValid() {
		switch provider.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			missingProvider = provider.IsNil()
		}
	}
	if missingProvider {
		return nil, errors.New("alicloud: credentials provider required")
	}
	if config.Timeout < 0 || config.MaxResponseBytes < 0 || config.MaxResponseBytes == int64(1<<63-1) {
		return nil, errors.New("alicloud: invalid limits")
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxResponseBytes == 0 {
		config.MaxResponseBytes = 8 << 20
	}
	if config.BaseEndpoint != "" {
		if err := endpoint.Validate(config.BaseEndpoint); err != nil {
			return nil, err
		}
	}
	if config.EndpointResolver == nil {
		config.EndpointResolver = endpoint.DefaultResolver()
	}
	if config.Retryer == nil {
		config.Retryer = retry.NoRetry{}
	}
	if config.Retryer.MaxAttempts() < 1 || config.Retryer.MaxAttempts() > 100 {
		return nil, errors.New("alicloud: retry attempt limit must be in [1,100]")
	}
	if config.Sleep == nil {
		config.Sleep = retry.Wait
	}
	var httpClient http.Client
	if config.HTTPClient == nil {
		config.HTTPClient = &httpClient
	}
	if supplied, ok := config.HTTPClient.(*http.Client); ok {
		if supplied == nil {
			return nil, errors.New("alicloud: nil HTTP client")
		}
		copyClient := *supplied
		copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		config.HTTPClient = &copyClient
	}
	config.Middleware = append([]middleware.Registration(nil), config.Middleware...)
	stack, err := middleware.NewStack(config.Middleware)
	if err != nil {
		return nil, err
	}
	return &Client{config: config, stack: stack}, nil
}

// Region returns the configured default region without exposing mutable configuration.
func (c *Client) Region() string { return c.config.Region }

// Config returns a configuration snapshot with copied middleware registrations.
// Extension objects are shared; mutating the snapshot never changes this client.
func (c *Client) Config() Config {
	config := c.config
	config.Middleware = append([]middleware.Registration(nil), config.Middleware...)
	return config
}

// Codec binds an owned model input and output to reviewed wire serialization.
// Functions must honor context, use JSON v2 and avoid retaining model pointers.
type Codec struct {
	// Encode converts the owned typed input into a new request after Initialize.
	Encode func(context.Context, any) (Request, error)
	// Decode assigns a fresh typed output after successful response decoding.
	// On error the temporary output is discarded before publication.
	Decode func(context.Context, []byte, any) error
}

type modelBinding struct {
	input any
	codec Codec
}

// InvokeModel executes a typed operation with Initialize and Serialize hooks.
// Input must be a private owned snapshot; generated clients copy caller models.
// Output must be a non-nil pointer. Hooks may replace models with the exact same
// pointer type; they must not retain them. Publication is atomic on success.
func (c *Client) InvokeModel(ctx context.Context, op Operation, input any, request Request, output any, codec Codec, optFns ...func(*CallOptions)) (Metadata, error) {
	return c.invoke(ctx, op, request, output, &modelBinding{input: input, codec: codec}, optFns...)
}

// Invoke sends copied wire data and decodes into a non-nil pointer. Output is
// assigned only after successful JSON v2 decoding; unknown fields are ignored,
// duplicate names and invalid UTF-8 fail. All operation failures wrap OperationError.
// Callers must not concurrently mutate request inputs or share output pointers.
func (c *Client) Invoke(ctx context.Context, op Operation, input Request, output any, optFns ...func(*CallOptions)) (meta Metadata, err error) {
	return c.invoke(ctx, op, input, output, nil, optFns...)
}

func (c *Client) invoke(ctx context.Context, op Operation, input Request, output any, binding *modelBinding, optFns ...func(*CallOptions)) (meta Metadata, err error) {
	defer func() {
		if err != nil {
			err = &OperationError{Service: op.Service, Operation: op.Name, Metadata: meta, Err: err}
		}
	}()
	if err = ctx.Err(); err != nil {
		return
	}
	if op.Service == "" || op.Name == "" || op.Version == "" {
		err = errors.New("alicloud: incomplete operation")
		return
	}
	if op.Authentication != AuthenticationACS3 && op.Authentication != AuthenticationAnonymousRPC {
		err = errors.New("alicloud: unsupported authentication protocol")
		return
	}
	target := reflect.ValueOf(output)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		err = errors.New("alicloud: output must be a non-nil pointer")
		return
	}
	var decoded reflect.Value
	if binding != nil && (binding.codec.Encode == nil || binding.codec.Decode == nil || binding.input == nil) {
		err = errors.New("alicloud: incomplete model codec")
		return
	}
	if binding != nil && (reflect.ValueOf(binding.input).Kind() != reflect.Pointer || reflect.ValueOf(binding.input).IsNil()) {
		err = errors.New("alicloud: model input must be a non-nil pointer")
		return
	}
	config := c.config
	region := input.Region
	if region == "" {
		region = c.config.Region
	}
	options := CallOptions{Region: region, BaseEndpoint: c.config.BaseEndpoint, Timeout: c.config.Timeout}
	for _, f := range optFns {
		if f == nil {
			err = errors.New("alicloud: nil call option")
			return
		}
		f(&options)
	}
	if options.Config != nil {
		configured, configErr := NewClient(*options.Config)
		if configErr != nil {
			err = configErr
			return
		}
		config = configured.config
		if options.Region == region && input.Region == "" {
			options.Region = config.Region
		}
		if options.BaseEndpoint == c.config.BaseEndpoint {
			options.BaseEndpoint = config.BaseEndpoint
		}
		if options.Timeout == c.config.Timeout {
			options.Timeout = config.Timeout
		}
	}
	if options.Timeout == 0 {
		options.Timeout = config.Timeout
	}
	if options.Timeout < 0 {
		err = errors.New("alicloud: timeout must be positive")
		return
	}
	callCtx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()
	stack := c.stack
	if options.Config != nil {
		stack, err = middleware.NewStack(config.Middleware)
		if err != nil {
			return
		}
	}
	if len(options.Middleware) > 0 {
		regs := append(append([]middleware.Registration(nil), config.Middleware...), options.Middleware...)
		stack, err = middleware.NewStack(regs)
		if err != nil {
			return
		}
	}
	e := &middleware.Exchange{Service: op.Service, Operation: op.Name, Region: options.Region}
	if binding != nil {
		e.Input = binding.input
	}
	body := append([]byte(nil), input.Body...)
	query := cloneQuery(input.Query)
	headers := input.Header.Clone()
	if len(body) > 8<<20 {
		err = errors.New("alicloud: request exceeds eight MiB")
		return
	}
	err = stack.Run(callCtx, middleware.Initialize, e, func(ctx context.Context, e *middleware.Exchange) error {
		serialized := binding == nil
		serializeErr := stack.Run(ctx, middleware.Serialize, e, func(ctx context.Context, e *middleware.Exchange) error {
			if binding == nil {
				return nil
			}
			if reflect.TypeOf(e.Input) != reflect.TypeOf(binding.input) || reflect.ValueOf(e.Input).IsNil() {
				return errors.New("alicloud: invalid middleware input type")
			}
			encoded, encodeErr := binding.codec.Encode(ctx, e.Input)
			if encodeErr != nil {
				return encodeErr
			}
			input = encoded
			query = cloneQuery(encoded.Query)
			headers = encoded.Header.Clone()
			body = append([]byte(nil), encoded.Body...)
			serialized = true
			if len(body) > 8<<20 {
				return errors.New("alicloud: request exceeds eight MiB")
			}
			if encoded.Region != "" && options.Region == region && e.Region == region {
				e.Region = encoded.Region
			}
			return nil
		})
		if serializeErr != nil {
			return serializeErr
		}
		if !serialized {
			if e.Output == nil {
				return ErrIncompleteOperation
			}
			return nil
		}
		if _, ok := query["RegionId"]; ok {
			if e.Region == "" {
				return errors.New("alicloud: region required for RegionId")
			}
			query.Set("RegionId", e.Region)
		}
		resolved, resolveErr := config.EndpointResolver.ResolveEndpoint(ctx, endpoint.Parameters{Service: op.Service, Region: e.Region, BaseEndpoint: options.BaseEndpoint})
		if resolveErr != nil {
			return resolveErr
		}
		if validateErr := endpoint.Validate(resolved.URL); validateErr != nil {
			return validateErr
		}
		u, _ := url.Parse(resolved.URL)
		path := input.Path
		if path == "" {
			path = "/"
		}
		if !strings.HasPrefix(path, "/") {
			return errors.New("alicloud: path must be absolute")
		}
		u.Path = path
		u.RawPath = ""
		u.RawQuery = signing.CanonicalQuery(query)
		method := input.Method
		if method == "" {
			method = http.MethodPost
		}
		if method != strings.ToUpper(method) {
			return errors.New("alicloud: method must be uppercase")
		}
		r, newErr := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
		if newErr != nil {
			return errors.New("alicloud: invalid request")
		}
		r.Header = headers.Clone()
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("Accept", "application/json")
		e.Request = r
		return stack.Run(ctx, middleware.Build, e, func(ctx context.Context, e *middleware.Exchange) error {
			if e.Request == nil {
				return errors.New("alicloud: middleware removed request")
			}
			base := e.Request.Clone(ctx)
			payload, readErr := requestBytes(e.Request)
			if readErr != nil {
				return readErr
			}
			attempts := config.Retryer.MaxAttempts()
			if attempts < 1 || attempts > 100 {
				return errors.New("alicloud: invalid retry limit")
			}
			for number := 1; number <= attempts; number++ {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				e.Attempt = number
				decoded = reflect.Value{}
				e.Output = nil
				meta.Attempts = number
				e.Response = nil
				e.RequestID = ""
				meta.RequestID = ""
				meta.HTTPStatusCode = 0
				e.Request = base.Clone(ctx)
				setBody(e.Request, payload)
				attemptErr := stack.Run(ctx, middleware.Finalize, e, func(ctx context.Context, e *middleware.Exchange) error {
					if e.Request == nil || e.Request.URL == nil || e.Request.URL.Scheme != "https" || e.Request.URL.Hostname() == "" || e.Request.URL.User != nil {
						return endpoint.ErrInvalid
					}
					e.Request = e.Request.WithContext(ctx)
					actual, readErr := requestBytes(e.Request)
					if readErr != nil {
						return readErr
					}
					setBody(e.Request, actual)
					var value credentials.Credentials
					if op.Authentication == AuthenticationACS3 {
						var credErr error
						value, credErr = config.CredentialsProvider.Retrieve(ctx)
						if credErr != nil {
							return credErr
						}
						if strings.TrimSpace(value.AccessKeyID) == "" || strings.TrimSpace(value.AccessKeySecret) == "" {
							return credentials.ErrMissingCredentials
						}
						if !value.ExpiresAt.IsZero() && !time.Now().Before(value.ExpiresAt) {
							return credentials.ErrExpired
						}
					}
					var nonce [16]byte
					if _, randErr := rand.Read(nonce[:]); randErr != nil {
						return errors.New("alicloud: nonce generation failed")
					}
					var signErr error
					if op.Authentication == AuthenticationAnonymousRPC {
						if len(actual) != 0 {
							return errors.New("alicloud: anonymous RPC body unsupported")
						}
						signErr = signing.PrepareAnonymousRPC(e.Request, op.Name, op.Version, time.Now(), hex.EncodeToString(nonce[:]))
					} else {
						signErr = signing.Sign(e.Request, actual, value, op.Name, op.Version, time.Now(), hex.EncodeToString(nonce[:]))
					}
					if signErr != nil {
						return signErr
					}
					response, sendErr := config.HTTPClient.Do(e.Request)
					if response != nil && response.Body != nil {
						defer response.Body.Close()
					}
					if sendErr != nil {
						return sendErr
					}
					if response == nil || response.Body == nil {
						return errors.New("alicloud: transport returned incomplete response")
					}
					e.Response = response
					meta.HTTPStatusCode = response.StatusCode
					return stack.Run(ctx, middleware.Deserialize, e, func(ctx context.Context, e *middleware.Exchange) error {
						if ctx.Err() != nil {
							return ctx.Err()
						}
						data, readErr := io.ReadAll(io.LimitReader(response.Body, config.MaxResponseBytes+1))
						if readErr != nil {
							return &retry.ResponseReadError{Err: readErr}
						}
						if int64(len(data)) > config.MaxResponseBytes {
							return ErrResponseTooLarge
						}
						envelope := struct {
							Code      string `json:"Code"`
							Message   string `json:"Message"`
							RequestID string `json:"RequestId"`
						}{}
						decodeErr := json.Unmarshal(data, &envelope)
						meta.RequestID = response.Header.Get("X-Acs-Request-Id")
						if envelope.RequestID != "" {
							meta.RequestID = envelope.RequestID
						}
						e.RequestID = meta.RequestID
						if response.StatusCode < 200 || response.StatusCode >= 300 {
							if decodeErr != nil {
								envelope.Code = "InvalidErrorResponse"
								envelope.Message = ""
							}
							return &APIError{Code: envelope.Code, Message: envelope.Message, RequestID: meta.RequestID, HTTPStatusCode: response.StatusCode}
						}
						if decodeErr != nil {
							return decodeErr
						}
						temporary := reflect.New(target.Elem().Type())
						if binding == nil {
							decodeErr = json.Unmarshal(data, temporary.Interface())
						} else {
							decodeErr = binding.codec.Decode(ctx, data, temporary.Interface())
						}
						if decodeErr != nil {
							return decodeErr
						}
						decoded = temporary.Elem()
						e.Output = temporary.Interface()
						return nil
					})
				})
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if attemptErr == nil {
					result := reflect.ValueOf(e.Output)
					if !result.IsValid() || result.Type() != target.Type() || result.IsNil() {
						return ErrIncompleteOperation
					}
					decoded = result.Elem()
					config.Retryer.RecordSuccess()
					return nil
				}
				a := retry.Attempt{Number: number, Err: attemptErr, StatusCode: meta.HTTPStatusCode, Idempotent: op.Idempotent, Replayable: true}
				var api *APIError
				if errors.As(attemptErr, &api) {
					a.StatusCode = api.HTTPStatusCode
					a.Code = api.Code
				}
				if e.Response != nil {
					a.RetryAfter = retry.ParseRetryAfter(e.Response.Header.Get("Retry-After"), time.Now())
				}
				if number == attempts || !config.Retryer.ShouldRetry(a) {
					return attemptErr
				}
				delay := config.Retryer.Delay(a)
				if delay < 0 {
					return errors.New("alicloud: negative retry delay")
				}
				if sleepErr := config.Sleep(ctx, delay); sleepErr != nil {
					return sleepErr
				}
			}
			return errors.New("alicloud: exhausted attempts")
		})
	})
	if callCtx.Err() != nil {
		err = callCtx.Err()
	}
	if err == nil {
		result := reflect.ValueOf(e.Output)
		if !result.IsValid() || result.Type() != target.Type() || result.IsNil() {
			err = ErrIncompleteOperation
		} else {
			decoded = result.Elem()
		}
	}
	if err == nil && decoded.IsValid() {
		target.Elem().Set(decoded)
	}
	return
}
func cloneQuery(input url.Values) url.Values {
	copyQuery := make(url.Values, len(input))
	for k, v := range input {
		copyQuery[k] = append([]string(nil), v...)
	}
	return copyQuery
}
func requestBytes(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, (8<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(body) > 8<<20 {
		return nil, errors.New("alicloud: request exceeds eight MiB")
	}
	return body, nil
}
func setBody(r *http.Request, body []byte) {
	r.ContentLength = int64(len(body))
	if len(body) == 0 {
		r.Body = http.NoBody
		r.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
}
