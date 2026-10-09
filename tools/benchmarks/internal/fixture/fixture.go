// Package fixture supplies the same account-free identity response to both SDKs.
package fixture

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

// Transport returns one synthetic identity response per correctly framed call.
// It performs no network I/O. A separate transport is used by each serial workload.
type Transport struct {
	// Calls counts successful serial fixture calls. It is not safe for concurrent access.
	Calls int
}

// Call adapts the official SDK's transport seam without opening a connection.
func (t *Transport) Call(r *http.Request, _ *http.Transport) (*http.Response, error) {
	return t.RoundTrip(r)
}

// RoundTrip validates the native action and consumes the request before replying.
func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	action := ""
	for k, v := range r.Header {
		if strings.EqualFold(k, "X-Acs-Action") && len(v) > 0 {
			action = v[0]
		}
	}
	if action == "" {
		action = r.URL.Query().Get("Action")
	}
	if action != "GetCallerIdentity" || r.URL.Host != "example.invalid" {
		return nil, errors.New("fixture: identity framing mismatch")
	}
	if r.Body != nil {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			return nil, err
		}
	}
	t.Calls++
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"RequestId":"fixture-request","AccountId":"fixture-account","IdentityType":"RAMUser"}`)), Request: r}, nil
}
