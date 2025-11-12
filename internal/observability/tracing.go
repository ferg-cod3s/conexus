package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerConfig configures OpenTelemetry tracing.
type TracerConfig struct {
	// ServiceName is the name of the service (defaults to "conexus")
	ServiceName string
	// ServiceVersion is the version of the service
	ServiceVersion string
	// Environment is the deployment environment (dev, staging, prod)
	Environment string
	// OTLPEndpoint is the OpenTelemetry collector endpoint
	OTLPEndpoint string
	// SamplingRate is the trace sampling rate (0.0 to 1.0)
	SamplingRate float64
	// Enabled enables tracing (can be disabled for development)
	Enabled bool
}

// DefaultTracerConfig returns a default tracer configuration.
func DefaultTracerConfig() TracerConfig {
	return TracerConfig{
		ServiceName:    "conexus",
		ServiceVersion: "0.1.0",
		Environment:    "development",
		OTLPEndpoint:   "localhost:4317",
		SamplingRate:   1.0,
		Enabled:        false, // Disabled by default
	}
}

// TracerProvider wraps the OpenTelemetry tracer provider.
type TracerProvider struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
}

// NewTracerProvider creates a new OpenTelemetry tracer provider.
func NewTracerProvider(cfg TracerConfig) (*TracerProvider, error) {
	if !cfg.Enabled {
		// Return a no-op tracer provider
		return &TracerProvider{
			provider: nil,
			tracer:   otel.Tracer(cfg.ServiceName),
		}, nil
	}

	// Create OTLP exporter
	ctx := context.Background()
	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlptracegrpc.WithInsecure(), // Use TLS in production
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRate)),
	)

	// Set global tracer provider and propagator
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &TracerProvider{
		provider: provider,
		tracer:   provider.Tracer(cfg.ServiceName),
	}, nil
}

// Tracer returns the OpenTelemetry tracer.
func (tp *TracerProvider) Tracer() trace.Tracer {
	return tp.tracer
}

// Shutdown shuts down the tracer provider.
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.provider != nil {
		return tp.provider.Shutdown(ctx)
	}
	return nil
}

// StartSpan starts a new span with the given name and options.
func (tp *TracerProvider) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tp.tracer.Start(ctx, name, opts...)
}

// SpanFromContext returns the current span from the context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// SetSpanAttributes sets attributes on the current span.
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	trace.SpanFromContext(ctx).SetAttributes(attrs...)
}

// SetSpanError records an error on the current span.
func SetSpanError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// AddSpanEvent adds an event to the current span.
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	trace.SpanFromContext(ctx).AddEvent(name, trace.WithAttributes(attrs...))
}

// TraceID returns the trace ID from the context.
func TraceID(ctx context.Context) string {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// SpanID returns the span ID from the context.
func SpanID(ctx context.Context) string {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}
	return ""
}

// InstrumentMCPRequest instruments an MCP request with tracing.
func InstrumentMCPRequest(ctx context.Context, tracer trace.Tracer, method string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("mcp.%s", method),
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("mcp.method", method),
		),
	)
	return ctx, span
}

// InstrumentIndexerOperation instruments an indexer operation with tracing.
func InstrumentIndexerOperation(ctx context.Context, tracer trace.Tracer, operation, path string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("indexer.%s", operation),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("indexer.operation", operation),
			attribute.String("indexer.path", path),
		),
	)
	return ctx, span
}

// InstrumentEmbedding instruments an embedding request with tracing.
func InstrumentEmbedding(ctx context.Context, tracer trace.Tracer, provider string, textLength int) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("embedding.%s", provider),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("embedding.provider", provider),
			attribute.Int("embedding.text_length", textLength),
		),
	)
	return ctx, span
}

// InstrumentVectorSearch instruments a vector search with tracing.
func InstrumentVectorSearch(ctx context.Context, tracer trace.Tracer, searchType string, limit int) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("search.%s", searchType),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("search.type", searchType),
			attribute.Int("search.limit", limit),
		),
	)
	return ctx, span
}
