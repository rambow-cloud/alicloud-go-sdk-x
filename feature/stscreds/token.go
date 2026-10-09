package stscreds

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
)

// TokenProvider supplies current federation material on each credential refresh.
// Implementations must support concurrent calls and honor cancellation. Return the
// original OIDC token or Base64 SAML assertion; do not decode or log the material.
type TokenProvider interface {
	// RetrieveToken returns an owned token string or an inspectable error.
	RetrieveToken(context.Context) (string, error)
}

// TokenProviderFunc adapts a concurrency-safe function into a TokenProvider.
// A nil function returns an error.
type TokenProviderFunc func(context.Context) (string, error)

// RetrieveToken calls f after checking cancellation.
func (f TokenProviderFunc) RetrieveToken(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if f == nil {
		return "", errors.New("stscreds: token provider required")
	}
	return f(ctx)
}

// FileTokenProvider reloads a token file on each call, supporting atomic rotation.
// Construct with NewFileTokenProvider. Concurrent reads are safe. File I/O itself
// cannot be interrupted; cancellation is checked before and after bounded reads.
// Writers should replace a regular file atomically. The SDK does not write it.
type FileTokenProvider struct{ filename string }

// NewFileTokenProvider records a nonempty token filename without reading it.
// Reads permit at most 1 MiB. Missing or invalid files return sanitized errors.
func NewFileTokenProvider(filename string) (*FileTokenProvider, error) {
	if strings.TrimSpace(filename) == "" {
		return nil, errors.New("stscreds: token filename required")
	}
	return &FileTokenProvider{filename: filename}, nil
}

// String returns a redacted representation without the filename or token.
func (*FileTokenProvider) String() string { return "FileTokenProvider(<redacted>)" }

// GoString returns a redacted representation for Go-syntax formatting.
func (p *FileTokenProvider) GoString() string { return p.String() }

// RetrieveToken reads current token bytes, trimming surrounding whitespace.
// Nonregular, empty and oversized files fail. Errors never contain file paths.
func (p *FileTokenProvider) RetrieveToken(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if p == nil || p.filename == "" {
		return "", errors.New("stscreds: unconfigured token file")
	}
	// Stat before Open avoids blocking on named pipes and devices. A replacement
	// during this check must still be a regular file; token writers are trusted.
	info, err := os.Stat(p.filename)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("stscreds: cannot read regular token file")
	}
	f, err := os.Open(p.filename)
	if err != nil {
		return "", errors.New("stscreds: cannot open token file")
	}
	defer f.Close()
	info, err = f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("stscreds: cannot read regular token file")
	}
	const limit = 1 << 20
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if e := ctx.Err(); e != nil {
		return "", e
	}
	if err != nil || len(b) > limit {
		return "", errors.New("stscreds: invalid token file size or read")
	}
	token := strings.TrimSpace(string(b))
	if token == "" {
		return "", errors.New("stscreds: empty token file")
	}
	return token, nil
}
