package alicloud

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"sync"
)

// StreamingOutput receives a low-level ResponseBodyStream result. Its zero value
// has no body. After successful Invoke, the caller must close Body.
type StreamingOutput struct {
	// Body is an owned bounded response stream. Read has one owner; Close may
	// run concurrently with Read. The operation context and timeout remain active.
	Body io.ReadCloser
	// Headers is an owned copy of the response headers.
	Headers http.Header
	// StatusCode is the successful HTTP status code.
	StatusCode int
}

type responseStream struct {
	body      io.ReadCloser
	ctx       context.Context
	op        Operation
	meta      Metadata
	limit     int64
	read      int64 // Read has one owner.
	mu        sync.Mutex
	ended     bool
	cause     error
	cancel    context.CancelFunc
	stop      func() bool
	closeOnce sync.Once
	closeErr  error
}

func newResponseStream(ctx context.Context, body io.ReadCloser, limit int64, op Operation, meta Metadata) *responseStream {
	s := &responseStream{ctx: ctx, body: body, limit: limit, op: op, meta: meta}
	s.mu.Lock()
	s.stop = context.AfterFunc(ctx, func() { s.finish(ctx.Err()) })
	s.mu.Unlock()
	return s
}

func (s *responseStream) wrap(err error) error {
	if err == nil || err == io.EOF {
		return err
	}
	return &OperationError{Service: s.op.Service, Operation: s.op.Name, Metadata: s.meta, Err: err}
}

func (s *responseStream) closeBody() {
	s.closeOnce.Do(func() { s.closeErr = s.body.Close() })
}

func (s *responseStream) finish(cause error) {
	s.mu.Lock()
	if !s.ended {
		s.ended, s.cause = true, cause
	}
	cancel, stop := s.cancel, s.stop
	s.stop = nil
	s.mu.Unlock()
	if stop != nil {
		stop()
	}
	s.closeBody()
	if cancel != nil {
		cancel()
	}
}

// adopt transfers operation cancellation only after successful publication.
// Closing a failed attempt must not cancel an otherwise retryable operation.
func (s *responseStream) adopt(cancel context.CancelFunc) {
	s.mu.Lock()
	s.cancel = cancel
	ended := s.ended
	s.mu.Unlock()
	if ended {
		cancel()
	}
}

func (s *responseStream) terminal() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ended, s.cause
}

func (s *responseStream) Read(p []byte) (int, error) {
	if s.ctx.Err() != nil {
		s.finish(s.ctx.Err())
	}
	if ended, cause := s.terminal(); ended {
		return 0, s.wrap(cause)
	}
	if len(p) == 0 {
		return 0, nil
	}
	remaining := s.limit - s.read
	var probe [1]byte
	buffer := p
	if remaining == 0 {
		buffer = probe[:]
	} else if int64(len(buffer)) > remaining {
		buffer = buffer[:remaining]
	}
	n, err := s.body.Read(buffer)
	if remaining == 0 && n > 0 {
		n, err = 0, ErrResponseTooLarge
	}
	s.read += int64(n)
	if s.ctx.Err() != nil {
		err = s.ctx.Err()
	}
	if err != nil {
		s.finish(err)
	}
	if ended, cause := s.terminal(); ended {
		err = cause
	}
	return n, s.wrap(err)
}

func (s *responseStream) Close() error {
	s.finish(io.EOF)
	return s.wrap(s.closeErr)
}

var readCloserType = reflect.TypeFor[io.ReadCloser]()

func hasStreamField(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() && field.Type == readCloserType {
			return true
		}
	}
	return false
}

func ownsStream(output reflect.Value, stream *responseStream) bool {
	if stream == nil || !output.IsValid() || output.Kind() != reflect.Pointer || output.IsNil() {
		return false
	}
	v := output.Elem()
	if v.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if field.IsExported() && field.Type == readCloserType && !v.Field(i).IsNil() {
			if owned, ok := v.Field(i).Interface().(*responseStream); ok && owned == stream {
				return true
			}
		}
	}
	return false
}
