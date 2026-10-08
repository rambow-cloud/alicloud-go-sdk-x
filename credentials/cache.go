package credentials

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CacheOptions controls refresh behavior. Function fields must be concurrency safe.
type CacheOptions struct {
	// ExpiryWindow starts early background refresh; zero defaults to one minute.
	ExpiryWindow time.Duration
	// RefreshTimeout bounds shared source calls; zero defaults to ten seconds.
	RefreshTimeout time.Duration
	// Now supplies the clock; nil uses time.Now.
	Now func() time.Time
}
type refresh struct {
	done       chan struct{}
	err        error
	generation uint64
}

// Cache coalesces source calls and never returns known expired credentials.
// Construct with NewCache. A canceled caller does not cancel a shared refresh.
// Sources must honor context deadlines; concurrent use is safe.
type Cache struct {
	mu         sync.Mutex
	source     Provider
	options    CacheOptions
	value      Credentials
	flight     *refresh
	generation uint64
}

// NewCache validates options and wraps a shared provider. A zero expiration
// snapshot is cached until Invalidate; a near-expiry valid snapshot is returned
// immediately while one refresh runs in the background.
// Nil and typed-nil sources are rejected without retrieval.
func NewCache(source Provider, options CacheOptions) (*Cache, error) {
	if isNilProvider(source) {
		return nil, errors.New("credentials: nil cache source")
	}
	if options.ExpiryWindow < 0 || options.RefreshTimeout < 0 {
		return nil, errors.New("credentials: invalid cache durations")
	}
	if options.ExpiryWindow == 0 {
		options.ExpiryWindow = time.Minute
	}
	if options.RefreshTimeout == 0 {
		options.RefreshTimeout = 10 * time.Second
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Cache{source: source, options: options}, nil
}

// Invalidate clears the cached value and prevents an older in-flight refresh
// from publishing. An existing source call finishes before another starts.
func (c *Cache) Invalidate() { c.mu.Lock(); c.generation++; c.value = Credentials{}; c.mu.Unlock() }

// Retrieve returns a valid snapshot or waits for one bounded shared refresh.
// Cancellation only ends this caller's wait. Failed refreshes never publish values.
func (c *Cache) Retrieve(ctx context.Context) (Credentials, error) {
	for {
		if err := ctx.Err(); err != nil {
			return Credentials{}, err
		}
		c.mu.Lock()
		now := c.options.Now()
		v := c.value
		valid := validate(v) == nil && (v.ExpiresAt.IsZero() || now.Before(v.ExpiresAt))
		if valid {
			if !v.ExpiresAt.IsZero() && !now.Add(c.options.ExpiryWindow).Before(v.ExpiresAt) && c.flight == nil {
				c.startLocked(ctx)
			}
			c.mu.Unlock()
			return v, nil
		}
		if c.flight == nil {
			c.startLocked(ctx)
		}
		f := c.flight
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return Credentials{}, ctx.Err()
		case <-f.done:
		}
		if err := ctx.Err(); err != nil {
			return Credentials{}, err
		}
		c.mu.Lock()
		current := f.generation == c.generation
		c.mu.Unlock()
		if current && f.err != nil {
			return Credentials{}, f.err
		}
	}
}
func (c *Cache) startLocked(ctx context.Context) {
	f := &refresh{done: make(chan struct{}), generation: c.generation}
	c.flight = f
	refreshCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), c.options.RefreshTimeout)
	go func() {
		defer cancel()
		v, err := c.source.Retrieve(refreshCtx)
		if refreshCtx.Err() != nil {
			err = refreshCtx.Err()
		}
		if err == nil {
			err = validate(v)
		}
		c.mu.Lock()
		if err == nil && !v.ExpiresAt.IsZero() && !c.options.Now().Before(v.ExpiresAt) {
			err = ErrExpired
		}
		if err == nil && f.generation == c.generation {
			c.value = v
		}
		f.err = err
		c.flight = nil
		close(f.done)
		c.mu.Unlock()
	}()
}
