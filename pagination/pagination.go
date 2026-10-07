package pagination

import (
	"context"
	"errors"
)

// ErrNoMorePages indicates exhaustion without calling the fetcher.
var ErrNoMorePages = errors.New("pagination: no more pages")

// ErrRepeatedCursor is the former strict duplicate-continuation diagnostic.
// Deprecated: duplicate protection now returns the fetched page and stops successfully.
var ErrRepeatedCursor = errors.New("pagination: repeated cursor")

// Options controls continuation protection. New defaults to stopping duplicates.
type Options struct {
	// StopOnDuplicateCursor stops after returning a page with an already seen cursor.
	// Explicit false disables cycle protection and may allow unbounded traversal.
	StopOnDuplicateCursor bool
}

// Cursor represents a token, a one-based page number, or both as defined by an adapter.
type Cursor struct {
	// Token is an opaque continuation token; it must not be logged.
	Token string
	// PageNumber is a page index, or zero for token-only adapters.
	PageNumber int
}

// Page contains one typed result and an explicit continuation decision.
type Page[T any] struct {
	// Value is the fetched result.
	Value T
	// Next selects the following request when HasMore is true.
	Next Cursor
	// HasMore is the service-specific continuation decision, independent of item count.
	HasMore bool
}

// Fetcher retrieves a page without mutating caller-owned inputs.
type Fetcher[T any] func(context.Context, Cursor) (Page[T], error)

// Paginator tracks cursors for one consumer. Construct with New; do not copy or share it.
type Paginator[T any] struct {
	fetch   Fetcher[T]
	cursor  Cursor
	seen    map[Cursor]bool
	done    bool
	options Options
}

// New constructs a paginator at initial and rejects a nil fetcher.
func New[T any](initial Cursor, fetch Fetcher[T], optFns ...func(*Options)) (*Paginator[T], error) {
	if fetch == nil {
		return nil, errors.New("pagination: nil fetcher")
	}
	options := Options{StopOnDuplicateCursor: true}
	for _, f := range optFns {
		if f == nil {
			return nil, errors.New("pagination: nil option")
		}
		f(&options)
	}
	var seen map[Cursor]bool
	if options.StopOnDuplicateCursor {
		seen = map[Cursor]bool{initial: true}
	}
	return &Paginator[T]{fetch: fetch, cursor: initial, seen: seen, options: options}, nil
}

// HasMorePages reports whether another page can be requested; initially true.
func (p *Paginator[T]) HasMorePages() bool { return !p.done }

// NextPage retrieves a result, preserving cursor state on fetch or context failure.
// With duplicate protection, the repeated continuation's page is returned and
// iteration stops successfully. No fetched page is discarded by this protection.
func (p *Paginator[T]) NextPage(ctx context.Context) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	if p.done {
		return zero, ErrNoMorePages
	}
	page, err := p.fetch(ctx, p.cursor)
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}
	if err != nil {
		return zero, err
	}
	if page.HasMore {
		if p.options.StopOnDuplicateCursor && p.seen[page.Next] {
			p.done = true
			return page.Value, nil
		}
		if p.options.StopOnDuplicateCursor {
			p.seen[page.Next] = true
		}
		p.cursor = page.Next
	} else {
		p.done = true
	}
	return page.Value, nil
}
