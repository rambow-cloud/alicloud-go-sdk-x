package retry

import (
	"context"
	"errors"
	"math/rand/v2"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Attempt describes a failed HTTP attempt without sensitive request data.
type Attempt struct {
	// Number is the one-based attempt number.
	Number int
	// StatusCode is the response status or zero for transport errors.
	StatusCode int
	// Code is the service error code, if available.
	Code string
	// Err is the underlying failure.
	Err error
	// Idempotent indicates the operation can be replayed without duplicate effects.
	Idempotent bool
	// Replayable indicates the body can be reconstructed exactly.
	Replayable bool
	// RetryAfter is the parsed server delay; policy caps it to MaxDelay.
	RetryAfter time.Duration
}

// Retryer controls retry decisions and must be concurrency safe.
type Retryer interface {
	// MaxAttempts returns a positive total attempt limit, including the first.
	MaxAttempts() int
	// ShouldRetry decides once after a failure and may reserve shared budget.
	ShouldRetry(Attempt) bool
	// Delay returns a nonnegative backoff after ShouldRetry succeeds.
	Delay(Attempt) time.Duration
	// RecordSuccess refunds budget after a successful operation.
	RecordSuccess()
}

// NoRetry permits exactly one attempt. Its zero value is ready to use.
type NoRetry struct{}

// MaxAttempts returns one.
func (NoRetry) MaxAttempts() int { return 1 }

// ShouldRetry always returns false.
func (NoRetry) ShouldRetry(Attempt) bool { return false }

// Delay returns zero.
func (NoRetry) Delay(Attempt) time.Duration { return 0 }

// RecordSuccess has no effect.
func (NoRetry) RecordSuccess() {}

// Options configures Standard. Zero fields select documented defaults.
type Options struct {
	// MaxAttempts defaults to three and must not exceed 100.
	MaxAttempts int
	// BaseDelay defaults to 200 milliseconds.
	BaseDelay time.Duration
	// MaxDelay defaults to twenty seconds and bounds all delays, including Retry-After.
	MaxDelay time.Duration
	// Budget defaults to twenty retry tokens per Standard instance.
	Budget int
	// Jitter samples a duration in [0, ceiling]; nil uses math/rand/v2.
	Jitter func(ceiling time.Duration) time.Duration
}

// Standard is a mutex-protected retry policy. Construct with NewStandard.
// Sharing one instance shares its budget; create one per client for isolated budgets.
type Standard struct {
	options Options
	mu      sync.Mutex
	tokens  int
}

// NewStandard validates options and constructs a bounded retry budget.
func NewStandard(o Options) (*Standard, error) {
	if o.MaxAttempts < 0 || o.MaxAttempts > 100 || o.BaseDelay < 0 || o.MaxDelay < 0 || o.Budget < 0 {
		return nil, errors.New("retry: invalid options")
	}
	if o.MaxAttempts == 0 {
		o.MaxAttempts = 3
	}
	if o.BaseDelay == 0 {
		o.BaseDelay = 200 * time.Millisecond
	}
	if o.MaxDelay == 0 {
		o.MaxDelay = 20 * time.Second
	}
	if o.Budget == 0 {
		o.Budget = 20
	}
	if o.BaseDelay > o.MaxDelay {
		return nil, errors.New("retry: base delay exceeds maximum")
	}
	if o.Jitter == nil {
		o.Jitter = func(d time.Duration) time.Duration { return time.Duration(rand.Int64N(int64(d))) }
	}
	return &Standard{options: o, tokens: o.Budget}, nil
}

// MaxAttempts returns the configured total attempt limit.
func (s *Standard) MaxAttempts() int { return s.options.MaxAttempts }

// ShouldRetry reserves one token only for replayable idempotent operations and
// transient transport, throttling or selected server errors. Cancellation is never retried.
func (s *Standard) ShouldRetry(a Attempt) bool {
	if a.Number < 1 || a.Number >= s.options.MaxAttempts || !a.Idempotent || !a.Replayable || a.Err == nil || errors.Is(a.Err, context.Canceled) || errors.Is(a.Err, context.DeadlineExceeded) {
		return false
	}
	transient := a.StatusCode == 429 || a.StatusCode == 500 || a.StatusCode == 502 || a.StatusCode == 503 || a.StatusCode == 504 || a.Code == "Throttling" || strings.HasPrefix(a.Code, "Throttling.")
	var network net.Error
	if a.StatusCode == 0 && errors.As(a.Err, &network) {
		transient = true
	}
	if !transient {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokens == 0 {
		return false
	}
	s.tokens--
	return true
}

// Delay calculates capped exponential full jitter and honors a larger server
// delay up to MaxDelay. Invalid custom jitter output is clamped to the range.
func (s *Standard) Delay(a Attempt) time.Duration {
	ceiling := s.options.BaseDelay
	for n := 1; n < a.Number && ceiling < s.options.MaxDelay; n++ {
		if ceiling > s.options.MaxDelay/2 {
			ceiling = s.options.MaxDelay
		} else {
			ceiling *= 2
		}
	}
	d := s.options.Jitter(ceiling)
	if d < 0 {
		d = 0
	}
	if d > ceiling {
		d = ceiling
	}
	if a.RetryAfter > d {
		d = a.RetryAfter
	}
	if d > s.options.MaxDelay {
		d = s.options.MaxDelay
	}
	return d
}

// RecordSuccess refunds one token without exceeding the initial budget.
func (s *Standard) RecordSuccess() {
	s.mu.Lock()
	if s.tokens < s.options.Budget {
		s.tokens++
	}
	s.mu.Unlock()
}

// Wait sleeps until d elapses or context cancellation. Negative durations fail.
func Wait(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d < 0 {
		return errors.New("retry: negative delay")
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ctx.Err()
	}
}

// ParseRetryAfter accepts nonnegative delta seconds or an HTTP date. Invalid or
// past values return zero. Extremely large delta seconds saturate at maximum duration.
func ParseRetryAfter(raw string, now time.Time) time.Duration {
	if n, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64); err == nil {
		if n > uint64((1<<63-1)/int64(time.Second)) {
			return time.Duration(1<<63 - 1)
		}
		return time.Duration(n) * time.Second
	}
	if date, err := http.ParseTime(raw); err == nil && date.After(now) {
		return date.Sub(now)
	}
	return 0
}
