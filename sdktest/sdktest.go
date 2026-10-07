package sdktest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RoundTripperFunc adapts a function to http.RoundTripper.
type RoundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip delegates to f, which must obey http.RoundTripper's concurrency contract.
func (f RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Step describes one offline HTTP exchange.
type Step struct {
	// StatusCode is the response status; zero means 200.
	StatusCode int
	// Header contains response headers and is copied when the response is built.
	Header http.Header
	// Body is the response body.
	Body string
	// Check inspects the request and can reject it before returning a response.
	Check func(*http.Request) error
	// Err is a transport error returned after Check succeeds.
	Err error
}

// ScriptedTransport consumes a fixed sequence of responses. Use NewTransport;
// its zero value is exhausted. Requests are not retained to avoid storing secrets.
type ScriptedTransport struct {
	mu    sync.Mutex
	steps []Step
	calls int
}

// NewTransport copies the script and header maps. Check functions remain shared.
func NewTransport(steps ...Step) *ScriptedTransport {
	copied := append([]Step(nil), steps...)
	for i := range copied {
		copied[i].Header = copied[i].Header.Clone()
	}
	return &ScriptedTransport{steps: copied}
}

// Calls returns the number of attempts, including exhausted calls.
func (t *ScriptedTransport) Calls() int { t.mu.Lock(); defer t.mu.Unlock(); return t.calls }

// RoundTrip consumes a step and returns an in-memory response. It never uses the network.
func (t *ScriptedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if err := r.Context().Err(); err != nil {
		return nil, err
	}
	t.mu.Lock()
	index := t.calls
	t.calls++
	if index >= len(t.steps) {
		t.mu.Unlock()
		return nil, errors.New("sdktest: script exhausted")
	}
	s := t.steps[index]
	t.mu.Unlock()
	if s.Check != nil {
		if err := s.Check(r); err != nil {
			return nil, err
		}
	}
	if s.Err != nil {
		return nil, s.Err
	}
	status := s.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{StatusCode: status, Header: s.Header.Clone(), Body: io.NopCloser(strings.NewReader(s.Body)), Request: r}, nil
}

// Clock is a concurrency-safe virtual clock. Use NewClock; the zero value starts at year one.
type Clock struct {
	mu  sync.Mutex
	now time.Time
}

// NewClock constructs a clock at now.
func NewClock(now time.Time) *Clock { return &Clock{now: now} }

// Now returns the current virtual time.
func (c *Clock) Now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.now }

// Advance increments virtual time. Negative durations panic to prevent time reversal.
func (c *Clock) Advance(d time.Duration) {
	if d < 0 {
		panic("sdktest: negative advance")
	}
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()
}

// Sleep checks cancellation and advances virtual time without blocking.
func (c *Clock) Sleep(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d < 0 {
		return errors.New("sdktest: negative sleep")
	}
	c.Advance(d)
	return nil
}
