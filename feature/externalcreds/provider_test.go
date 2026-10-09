package externalcreds_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/externalcreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(r *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}
}

const issued = `{"Code":"Success","AccessKeyId":"synthetic-issued","AccessKeySecret":"synthetic-secret","SecurityToken":"synthetic-token","Expiration":"2099-01-01T00:00:00Z"}`

func TestURIValidationBoundsRedirectsAndErrors(t *testing.T) {
	for _, uri := range []string{"", "file:///private", "https://user:private@broker.invalid", "https://broker.invalid/#private"} {
		if _, err := externalcreds.NewURIProvider(uri, externalcreds.Options{}); err == nil {
			t.Fatal("invalid URI accepted")
		}
	}
	for _, tc := range []struct {
		name, body string
		status     int
		limit      int64
		want       error
	}{
		{"valid", issued, 200, 0, nil},
		{"missing-token", `{"AccessKeyId":"id","AccessKeySecret":"secret","Expiration":"2099-01-01T00:00:00Z"}`, 200, 0, credentials.ErrMissingCredentials},
		{"expired", strings.ReplaceAll(issued, "2099", "2000"), 200, 0, credentials.ErrExpired},
		{"missing-expiration", `{"AccessKeyId":"id","AccessKeySecret":"secret","SecurityToken":"token"}`, 200, 0, credentials.ErrExpired},
		{"inconsistent", strings.ReplaceAll(issued, `"Expiration":`, `"ExpirationInt64":1,"Expiration":`), 200, 0, errors.New("invalid")},
		{"duplicate", `{"AccessKeyId":"private","AccessKeyId":"two"}`, 200, 0, errors.New("invalid")},
		{"oversized", issued, 200, 8, errors.New("invalid")},
		{"redirect", issued, 302, 0, errors.New("invalid")},
		{"denied", `private-body`, 403, 0, errors.New("invalid")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != "GET" || r.URL.RawQuery != "token=private-query" {
					t.Fatal("URI wire mismatch")
				}
				res := response(r, tc.status, tc.body)
				res.Header.Set("Location", "https://leak.invalid/")
				return res, nil
			})}
			p, err := externalcreds.NewURIProvider("https://broker.invalid/?token=private-query", externalcreds.Options{HTTPClient: client, MaxResponseBytes: tc.limit})
			if err != nil || calls != 0 {
				t.Fatal("constructor performed HTTP")
			}
			c, err := p.Retrieve(context.Background())
			if tc.want == nil {
				if err != nil || c.AccessKeyID != "synthetic-issued" || c.Source != "CredentialsURI" {
					t.Fatal("valid broker response rejected")
				}
			} else if err == nil {
				t.Fatal("invalid broker response accepted")
			} else if tc.want == credentials.ErrExpired || tc.want == credentials.ErrMissingCredentials {
				if !errors.Is(err, tc.want) {
					t.Fatal("credential error identity lost")
				}
			}
			if calls != 1 || strings.Contains(fmt.Sprintf("%v %#v", p, err), "private") {
				t.Fatal("redirect, retry or diagnostic leak")
			}
		})
	}
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })}
	p, err := externalcreds.NewURIProvider("https://broker.invalid", externalcreds.Options{HTTPClient: client, Timeout: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Retrieve(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("retrieval timeout lost")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = p.Retrieve(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation lost")
	}
}

func TestIMDSv2SequenceAndNoDowngrade(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprint(explicit), func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Host != "100.100.100.200" {
					t.Fatal("metadata origin changed")
				}
				switch r.URL.Path {
				case "/latest/api/token":
					if r.Method != "PUT" || r.Header.Get("X-aliyun-ecs-metadata-token-ttl-seconds") != "21600" {
						t.Fatal("IMDSv2 token wire mismatch")
					}
					return response(r, 200, "synthetic-session"), nil
				case "/latest/meta-data/ram/security-credentials/":
					if explicit || r.Header.Get("X-aliyun-ecs-metadata-token") != "synthetic-session" {
						t.Fatal("role discovery mismatch")
					}
					return response(r, 200, "test-role\n"), nil
				case "/latest/meta-data/ram/security-credentials/test-role":
					if r.Header.Get("X-aliyun-ecs-metadata-token") != "synthetic-session" {
						t.Fatal("metadata token missing")
					}
					return response(r, 200, issued), nil
				default:
					t.Fatal("unexpected metadata path")
					return nil, errors.New("path")
				}
			})}
			role := ""
			want := 3
			if explicit {
				role = "test-role"
				want = 2
			}
			p, err := externalcreds.NewECSMetadataProvider(role, externalcreds.Options{HTTPClient: client})
			if err != nil || calls != 0 {
				t.Fatal("metadata construction was not lazy")
			}
			c, err := p.Retrieve(context.Background())
			if err != nil || c.Source != "ECS.IMDSv2" || calls != want {
				t.Fatal("IMDSv2 sequence failed")
			}
		})
	}
	for _, tc := range []struct {
		status int
		body   string
	}{{403, "denied"}, {200, ""}, {200, "token\nother"}} {
		calls := 0
		client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) { calls++; return response(r, tc.status, tc.body), nil })}
		p, _ := externalcreds.NewECSMetadataProvider("test-role", externalcreds.Options{HTTPClient: client})
		if _, err := p.Retrieve(context.Background()); err == nil || calls != 1 {
			t.Fatal("IMDSv1 downgrade or invalid token accepted")
		}
	}
	for _, role := range []string{"../other", "two\nroles", "role?query", ".."} {
		if _, err := externalcreds.NewECSMetadataProvider(role, externalcreds.Options{}); err == nil {
			t.Fatal("invalid role accepted")
		}
	}
}

func TestCloudSSOPortalWireRotationAndAmbiguity(t *testing.T) {
	var tokens atomic.Int32
	calls := 0
	token := stscreds.TokenProviderFunc(func(context.Context) (string, error) { return fmt.Sprintf("synthetic-login-%d", tokens.Add(1)), nil })
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls++
		body, _ := io.ReadAll(r.Body)
		if r.Method != "POST" || r.URL.String() != "https://portal.invalid/cloud-credentials" || r.Header.Get("Authorization") != fmt.Sprintf("Bearer synthetic-login-%d", calls) || string(body) != `{"AccountId":"account","AccessConfigurationId":"configuration"}` {
			t.Fatal("CloudSSO portal wire mismatch")
		}
		return response(r, 200, `{"CloudCredential":`+issued+`}`), nil
	})}
	p, err := externalcreds.NewCloudSSOProvider(externalcreds.CloudSSOOptions{SignInURL: "https://portal.invalid/login", AccountID: "account", AccessConfigurationID: "configuration", AccessToken: token, Retrieval: externalcreds.Options{HTTPClient: client}})
	if err != nil || calls != 0 || tokens.Load() != 0 {
		t.Fatal("CloudSSO construction retrieved tokens")
	}
	for range 2 {
		c, err := p.Retrieve(context.Background())
		if err != nil || c.Source != "CloudSSO" || c.AccessKeyID != "synthetic-issued" {
			t.Fatalf("portal exchange failed: %v", err)
		}
	}
	for _, body := range []string{`{"Code":"Denied","CloudCredential":` + issued + `}`, `{"AccessKeyId":"private","CloudCredential":` + issued + `}`, `{"CloudCredential":null}`, `{"CloudCredential":{"Expiration":"2000-01-01T00:00:00Z"}}`} {
		invalidClient := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) { return response(r, 200, body), nil })}
		invalid, err := externalcreds.NewCloudSSOProvider(externalcreds.CloudSSOOptions{SignInURL: "https://portal.invalid", AccountID: "account", AccessConfigurationID: "configuration", AccessToken: token, Retrieval: externalcreds.Options{HTTPClient: invalidClient}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := invalid.Retrieve(context.Background()); err == nil {
			t.Fatal("invalid portal response accepted")
		}
	}
}

func TestProcessHelper(t *testing.T) {
	if os.Getenv("SDK_TEST_CREDENTIAL_PROCESS") != "1" {
		return
	}
	switch os.Args[len(os.Args)-1] {
	case "valid":
		fmt.Print(`{"mode":"StsToken","access_key_id":"synthetic-issued","access_key_secret":"synthetic-secret","sts_token":"synthetic-token","sts_expiration":4070908800}`)
	case "ak":
		fmt.Print(`{"mode":"AK","access_key_id":"synthetic-issued","access_key_secret":"synthetic-secret"}`)
	case "oversized":
		fmt.Print(strings.Repeat("private", 500))
	case "sleep":
		time.Sleep(time.Minute)
	case "recursive":
		fmt.Print(`{"mode":"External","process_command":"private-command"}`)
	default:
		fmt.Fprintln(os.Stderr, "private-stderr")
		os.Exit(1)
	}
	os.Exit(0)
}

func TestProcessBoundedResultAndOptIn(t *testing.T) {
	t.Setenv("SDK_TEST_CREDENTIAL_PROCESS", "1")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		allow   bool
		limit   int64
		timeout time.Duration
		ok      bool
		want    error
	}{
		{"valid", false, 0, 0, true, nil}, {"ak", false, 0, 0, false, nil}, {"ak", true, 0, 0, true, nil}, {"recursive", false, 0, 0, false, nil}, {"oversized", false, 32, 0, false, nil}, {"error", false, 0, 0, false, nil}, {"sleep", false, 0, 100 * time.Millisecond, false, context.DeadlineExceeded},
	} {
		t.Run(fmt.Sprintf("%s-%t", tc.name, tc.allow), func(t *testing.T) {
			argv := []string{exe, "-test.run=^TestProcessHelper$", "--", tc.name}
			p, err := externalcreds.NewProcessProvider(argv, externalcreds.ProcessOptions{AllowLongLived: tc.allow, Retrieval: externalcreds.Options{Timeout: tc.timeout, MaxResponseBytes: tc.limit}})
			if err != nil {
				t.Fatal(err)
			}
			argv[len(argv)-1] = "mutation"
			c, err := p.Retrieve(context.Background())
			if tc.ok {
				if err != nil || c.AccessKeyID != "synthetic-issued" {
					t.Fatal("owned command/result failed")
				}
			} else if err == nil {
				t.Fatal("invalid process result accepted")
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatal("process timeout identity lost")
			}
			if strings.Contains(fmt.Sprintf("%v %#v", p, err), "private") {
				t.Fatal("process diagnostic leak")
			}
		})
	}
}

func TestCommandGrammarAndNilProviders(t *testing.T) {
	args, err := externalcreds.ParseCommand(`"C:\Program Files\broker.exe" '' "a b" '$HOME' "quoted\"value"`)
	if err != nil || len(args) != 5 || args[0] != `C:\Program Files\broker.exe` || args[1] != "" || args[2] != "a b" || args[3] != "$HOME" || args[4] != `quoted"value` {
		t.Fatal("literal argv grammar failed")
	}
	for _, command := range []string{"", `'' arg`, `"unterminated`} {
		if _, err := externalcreds.ParseCommand(command); err == nil {
			t.Fatal("invalid command accepted")
		}
	}
	ctx := context.Background()
	for _, p := range []credentials.Provider{(*externalcreds.URIProvider)(nil), (*externalcreds.ProcessProvider)(nil), (*externalcreds.CloudSSOProvider)(nil), (*externalcreds.ECSMetadataProvider)(nil)} {
		if _, err := p.Retrieve(ctx); err == nil {
			t.Fatal("nil provider accepted")
		}
	}
}
