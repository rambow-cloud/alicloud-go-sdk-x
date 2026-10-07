// Package otel provides opt-in OpenTelemetry operation and HTTP-attempt spans
// through middleware. It uses an injected TracerProvider and never changes global
// providers or exporters. Nil provider uses a local no-op provider. TraceContext
// propagation is enabled by default; baggage is not propagated unless explicitly
// selected. Raw URLs, query parameters, bodies, credentials and error messages
// are never recorded. Core packages do not import OpenTelemetry.
package otel
