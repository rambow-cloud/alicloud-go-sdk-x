package credentials

import (
	"context"
	"errors"
	"time"
)

// ProviderFunc adapts a concurrent-safe, context-aware retrieval function.
type ProviderFunc func(context.Context) (Credentials, error)

// Retrieve invokes f. A nil function must not be used.
func (f ProviderFunc) Retrieve(ctx context.Context) (Credentials, error) { return f(ctx) }

// Chain tries providers in explicit order; it never discovers files, processes
// or metadata implicitly. Its zero value returns ErrNotFound.
type Chain struct{ providers []Provider }

// NewChain copies providers and rejects nil and typed-nil sources. Sources remain shared and
// must obey Provider's concurrency and cancellation contract.
func NewChain(providers ...Provider) (*Chain, error) {
	for _, p := range providers {
		if isNilProvider(p) {
			return nil, errors.New("credentials: nil provider")
		}
	}
	return &Chain{providers: append([]Provider(nil), providers...)}, nil
}

// Retrieve returns the first valid snapshot, skips only ErrNotFound, and stops
// on partial configuration, invalid credentials, cancellation or retrieval failure.
func (c *Chain) Retrieve(ctx context.Context) (Credentials, error) {
	for _, p := range c.providers {
		if err := ctx.Err(); err != nil {
			return Credentials{}, err
		}
		v, err := p.Retrieve(ctx)
		if contextErr := ctx.Err(); contextErr != nil {
			return Credentials{}, contextErr
		}
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return Credentials{}, err
		}
		if err = validate(v); err != nil {
			return Credentials{}, err
		}
		if !v.ExpiresAt.IsZero() && !time.Now().Before(v.ExpiresAt) {
			return Credentials{}, ErrExpired
		}
		return v, nil
	}
	if err := ctx.Err(); err != nil {
		return Credentials{}, err
	}
	return Credentials{}, ErrNotFound
}
