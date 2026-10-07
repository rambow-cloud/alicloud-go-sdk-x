package pagination

import (
	"context"
	"errors"
)

// ErrNoMorePages indicates exhaustion without calling the fetcher.
var ErrNoMorePages = errors.New("pagination: no more pages")

// ErrRepeatedCursor indicates a duplicate or cyclic continuation cursor.
var ErrRepeatedCursor = errors.New("pagination: repeated cursor")

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
	fetch  Fetcher[T]
	cursor Cursor
	seen   map[Cursor]bool
	done   bool
}

// New constructs a paginator at initial and rejects a nil fetcher.
func New[T any](initial Cursor, fetch Fetcher[T]) (*Paginator[T], error) {
	if fetch == nil {
		return nil, errors.New("pagination: nil fetcher")
	}
	return &Paginator[T]{fetch: fetch, cursor: initial, seen: map[Cursor]bool{initial: true}}, nil
}

// HasMorePages reports whether another page can be requested; initially true.
func (p *Paginator[T]) HasMorePages() bool { return !p.done }

// NextPage retrieves a result, preserving cursor state on fetch or context failure.
// A repeated continuation returns ErrRepeatedCursor and permanently stops iteration.
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
		if p.seen[page.Next] {
			p.done = true
			return zero, ErrRepeatedCursor
		}
		p.seen[page.Next] = true
		p.cursor = page.Next
	} else {
		p.done = true
	}
	return page.Value, nil
}
