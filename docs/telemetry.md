# OpenTelemetry

[中文](telemetry.zh-CN.md)

- Opt in by assigning `otel.NewMiddleware(otel.Options{TracerProvider: provider})` from `telemetry/otel` to Config.Middleware.
- Inject an application-owned provider; no global provider/exporter is configured or read.
- Nil uses a private no-op provider.
- Each operation creates one logical span and each attempt a child span; defaults record service/action/region, attempt count, HTTP method/status and service request ID.
- Errors use type-only attributes and a fixed status description; no raw exception events, URLs, query values, bodies, authorization or credentials are recorded.
- W3C TraceContext is injected into attempts; baggage requires an explicit propagator.
- Place operation middleware first to include other middleware work.
- The application owns exporter/provider shutdown and sampling.
- Core import graphs remain standard-library-only; this optional adapter adds OpenTelemetry 1.47 dependencies.
- In-memory exporter tests verify retry hierarchy, propagation and redaction.
- See [OpenTelemetry Go instrumentation](https://opentelemetry.io/docs/languages/go/instrumentation/).
