package waiter

import (
	"context"
	"errors"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"time"
)

// ErrTimeout identifies exhaustion of the waiter's own total duration.
var ErrTimeout = errors.New("waiter: maximum wait exceeded")

// ErrFailure identifies an acceptor's terminal failure.
var ErrFailure = errors.New("waiter: terminal state")

// TimeoutError preserves waiter expiry, context.DeadlineExceeded and the last fetch failure.
type TimeoutError struct {
	// LastError is the last retryable fetch failure, or nil.
	LastError error
}

// Error returns a safe diagnostic without the underlying error text.
func (e *TimeoutError) Error() string { return ErrTimeout.Error() }

// Unwrap returns inspectable timeout and last-error causes.
func (e *TimeoutError) Unwrap() []error {
	causes := []error{ErrTimeout, context.DeadlineExceeded}
	if e.LastError != nil {
		causes = append(causes, e.LastError)
	}
	return causes
}

// FailureError preserves terminal rejection and an optional fetch cause.
type FailureError struct {
	// Err is the rejected fetch failure, or nil for a terminal state.
	Err error
}

// Error returns a diagnostic without including fetch error text.
func (e *FailureError) Error() string { return ErrFailure.Error() }

// Unwrap returns inspectable rejection and fetch causes.
func (e *FailureError) Unwrap() []error {
	if e.Err != nil {
		return []error{ErrFailure, e.Err}
	}
	return []error{ErrFailure}
}

// Decision is an acceptor's explicit polling decision.
type Decision uint8

const (
	// Retry requests another poll after backoff.
	Retry Decision = iota
	// Success completes the waiter with the current result.
	Success
	// Failure terminates with ErrFailure.
	Failure
)

// Fetcher retrieves a typed observation and must honor the supplied context.
type Fetcher[T any] func(context.Context) (T, error)

// Acceptor decides how to handle a result and its fetch error; it must not block.
type Acceptor[T any] func(T, error) Decision

// Options controls polling. Zero durations select one-second minimum and five-second maximum.
type Options struct {
	// MinDelay is the initial positive poll delay.
	MinDelay time.Duration
	// MaxDelay caps exponential poll delay.
	MaxDelay time.Duration
	// Now is a concurrency-safe clock; nil uses time.Now.
	Now func() time.Time
	// Sleep is cancellation-aware; nil uses retry.Wait.
	Sleep func(context.Context, time.Duration) error
}

// Waiter is an immutable polling engine. Construct with New.
type Waiter[T any] struct {
	fetch   Fetcher[T]
	accept  Acceptor[T]
	options Options
}

// New validates functions and delays. Callbacks are retained and must obey their contracts.
func New[T any](fetch Fetcher[T], accept Acceptor[T], options Options) (*Waiter[T], error) {
	if fetch == nil || accept == nil {
		return nil, errors.New("waiter: fetcher and acceptor required")
	}
	if options.MinDelay < 0 || options.MaxDelay < 0 {
		return nil, errors.New("waiter: negative delay")
	}
	if options.MinDelay == 0 {
		options.MinDelay = time.Second
	}
	if options.MaxDelay == 0 {
		options.MaxDelay = 5 * time.Second
	}
	if options.MinDelay > options.MaxDelay {
		return nil, errors.New("waiter: minimum exceeds maximum")
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.Sleep == nil {
		options.Sleep = retry.Wait
	}
	return &Waiter[T]{fetch: fetch, accept: accept, options: options}, nil
}

// Wait polls within a positive maxWait. Caller cancellation/deadline errors pass
// through; the engine's own expiry returns TimeoutError. Failed waits return zero T.
func (w *Waiter[T]) Wait(ctx context.Context, maxWait time.Duration) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	if maxWait <= 0 {
		return zero, errors.New("waiter: positive maximum wait required")
	}
	pollCtx, cancel := context.WithTimeout(ctx, maxWait)
	defer cancel()
	deadline := w.options.Now().Add(maxWait)
	delay := w.options.MinDelay
	var lastError error
	expired := func() bool { return pollCtx.Err() != nil || !w.options.Now().Before(deadline) }
	for {
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		if expired() {
			return zero, &TimeoutError{LastError: lastError}
		}
		value, err := w.fetch(pollCtx)
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		if expired() {
			return zero, &TimeoutError{LastError: err}
		}
		decision := w.accept(value, err)
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		if expired() {
			return zero, &TimeoutError{LastError: err}
		}
		switch decision {
		case Success:
			if err != nil {
				return zero, &FailureError{Err: err}
			}
			return value, nil
		case Failure:
			return zero, &FailureError{Err: err}
		case Retry:
			lastError = err
		default:
			return zero, errors.New("waiter: invalid acceptor decision")
		}
		remaining := deadline.Sub(w.options.Now())
		if remaining <= 0 {
			return zero, &TimeoutError{LastError: lastError}
		}
		sleep := min(delay, remaining)
		if err = w.options.Sleep(pollCtx, sleep); err != nil {
			if ctx.Err() != nil {
				return zero, ctx.Err()
			}
			if expired() {
				return zero, &TimeoutError{LastError: lastError}
			}
			return zero, err
		}
		if delay > w.options.MaxDelay/2 {
			delay = w.options.MaxDelay
		} else {
			delay = min(delay*2, w.options.MaxDelay)
		}
	}
}
