package otel

import (
	"context"
	"fmt"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Options supplies application-owned telemetry components, which must be concurrency safe.
type Options struct {
	// TracerProvider owns sampling and export; nil selects a private no-op provider.
	TracerProvider trace.TracerProvider
	// Propagator injects context into each attempt; nil uses W3C TraceContext without baggage.
	Propagator propagation.TextMapPropagator
}

// NewMiddleware returns Initialize and Finalize registrations. Place operation
// instrumentation first to include other middleware work in its span. The caller
// owns provider/exporter shutdown. No global telemetry state is read or written.
func NewMiddleware(options Options) []middleware.Registration {
	provider := options.TracerProvider
	if provider == nil {
		provider = noop.NewTracerProvider()
	}
	propagator := options.Propagator
	if propagator == nil {
		propagator = propagation.TraceContext{}
	}
	tracer := provider.Tracer("github.com/rambow-cloud/alicloud-go-sdk-x/telemetry/otel")
	operation := middleware.Func("otel.operation", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		ctx, span := tracer.Start(ctx, e.Service+"."+e.Operation, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attribute.String("rpc.system", "alicloud"), attribute.String("rpc.service", e.Service), attribute.String("rpc.method", e.Operation), attribute.String("cloud.region", e.Region)))
		defer span.End()
		err := next(ctx, e)
		span.SetAttributes(attribute.Int("alicloud.attempts", e.Attempt))
		annotate(span, e, err)
		return err
	})
	attempt := middleware.Func("otel.attempt", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		ctx, span := tracer.Start(ctx, e.Service+"."+e.Operation+".attempt", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attribute.Int("alicloud.attempt", e.Attempt)))
		defer span.End()
		if e.Request != nil {
			span.SetAttributes(attribute.String("http.request.method", e.Request.Method))
			propagator.Inject(ctx, propagation.HeaderCarrier(e.Request.Header))
		}
		err := next(ctx, e)
		annotate(span, e, err)
		return err
	})
	return []middleware.Registration{{Stage: middleware.Initialize, Middleware: operation}, {Stage: middleware.Finalize, Middleware: attempt}}
}
func annotate(span trace.Span, e *middleware.Exchange, err error) {
	if e.Response != nil {
		span.SetAttributes(attribute.Int("http.response.status_code", e.Response.StatusCode))
	}
	if e.RequestID != "" {
		span.SetAttributes(attribute.String("alicloud.request_id", e.RequestID))
	}
	if err != nil {
		span.SetStatus(codes.Error, "operation failed")
		span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", err)))
	}
}
