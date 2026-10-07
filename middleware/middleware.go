package middleware

import (
	"context"
	"fmt"
	"net/http"
)

// Stage identifies an execution boundary.
type Stage uint8

const (
	// Initialize surrounds the entire operation, including retries.
	Initialize Stage = iota
	// Serialize converts an owned model input to wire data once per operation.
	Serialize
	// Build surrounds request construction once per operation.
	Build
	// Finalize surrounds signing and transport once per attempt.
	Finalize
	// Deserialize surrounds response decoding once per attempt.
	Deserialize
)

// Exchange is mutable state owned by a single operation.
type Exchange struct {
	// Input is the owned typed input for model operations, or nil for wire Invoke.
	// Hooks may modify it before serialization and must not retain it after Handle.
	Input any
	// Output is the current attempt's typed result after decoding. A successful
	// short circuit must supply the operation's exact non-nil output pointer type.
	// Hooks may modify it after next; no result is published on pipeline failure.
	Output any
	// Service is the service identifier.
	Service string
	// Operation is the action name.
	Operation string
	// Region is the selected region.
	Region string
	// Attempt is the one-based HTTP attempt, or zero before sending.
	Attempt int
	// Request is nil before Build; modifications before next are signed.
	Request *http.Request
	// Response is nil until transport returns a response.
	Response *http.Response
	// RequestID is the latest decoded request identifier.
	RequestID string
}

// Handler continues a pipeline with the supplied context and exchange.
type Handler func(context.Context, *Exchange) error

// Middleware decorates a handler. Implementations must be concurrency safe.
type Middleware interface {
	// ID returns a stable, nonempty identifier unique within its stage.
	ID() string
	// Handle invokes next at most once and preserves cancellation errors.
	Handle(context.Context, *Exchange, Handler) error
}

type function struct {
	id string
	fn func(context.Context, *Exchange, Handler) error
}

func (f function) ID() string { return f.id }
func (f function) Handle(ctx context.Context, e *Exchange, next Handler) error {
	return f.fn(ctx, e, next)
}

// Func adapts a function to Middleware. A nil function is rejected by NewStack.
func Func(id string, fn func(context.Context, *Exchange, Handler) error) Middleware {
	return function{id, fn}
}

// Registration associates middleware with a stage.
type Registration struct {
	// Stage selects the boundary.
	Stage Stage
	// Middleware implements the interceptor.
	Middleware Middleware
}

// Stack is an immutable set of registrations. Its zero value passes through.
type Stack struct{ stages [5][]Middleware }

// NewStack copies registrations and rejects unknown stages, empty or duplicate IDs,
// and nil interceptors. Earlier registrations execute outside later ones.
func NewStack(registrations []Registration) (*Stack, error) {
	s := &Stack{}
	seen := [5]map[string]bool{}
	for _, r := range registrations {
		if r.Stage > Deserialize || r.Middleware == nil {
			return nil, fmt.Errorf("middleware: invalid registration")
		}
		if f, ok := r.Middleware.(function); ok && f.fn == nil {
			return nil, fmt.Errorf("middleware: nil function")
		}
		id := r.Middleware.ID()
		if id == "" {
			return nil, fmt.Errorf("middleware: empty ID")
		}
		if seen[r.Stage] == nil {
			seen[r.Stage] = map[string]bool{}
		}
		if seen[r.Stage][id] {
			return nil, fmt.Errorf("middleware: duplicate ID %q", id)
		}
		seen[r.Stage][id] = true
		s.stages[r.Stage] = append(s.stages[r.Stage], r.Middleware)
	}
	return s, nil
}

// Run executes one stage. Context errors short-circuit before middleware runs.
// An unknown stage returns an error. The exchange and final handler must be non-nil.
func (s *Stack) Run(ctx context.Context, stage Stage, e *Exchange, final Handler) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if stage > Deserialize || e == nil || final == nil {
		return fmt.Errorf("middleware: invalid invocation")
	}
	next := final
	for i := len(s.stages[stage]) - 1; i >= 0; i-- {
		m, inner := s.stages[stage][i], next
		next = func(ctx context.Context, e *Exchange) error { return m.Handle(ctx, e, inner) }
	}
	return next(ctx, e)
}
