package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestNewTracerProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  TracerConfig
		wantErr bool
	}{
		{
			name: "valid config with OTLP endpoint",
			config: TracerConfig{
				Enabled:        true,
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "test",
				OTLPEndpoint:   "localhost:4317",
				SamplingRate:   1.0,
			},
			wantErr: false,
		},
		{
			name: "disabled tracing",
			config: TracerConfig{
				Enabled:     false,
				ServiceName: "test-service",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp, err := NewTracerProvider(tt.config)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tp)

				// Cleanup
				if tp != nil {
					_ = tp.Shutdown(context.Background())
				}
			}
		})
	}
}

func TestTracerProviderShutdown(t *testing.T) {
	config := TracerConfig{
		Enabled:      false,
		ServiceName:  "test-service",
		Environment:  "test",
		SamplingRate: 1.0,
	}

	tp, err := NewTracerProvider(config)
	require.NoError(t, err)
	require.NotNil(t, tp)

	// Shutdown should not error
	err = tp.Shutdown(context.Background())
	assert.NoError(t, err)

	// Second shutdown should still not error
	err = tp.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestSetSpanAttributes(t *testing.T) {
	// Create in-memory tracer for testing
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	tests := []struct {
		name  string
		attrs []attribute.KeyValue
	}{
		{
			name: "string attributes",
			attrs: []attribute.KeyValue{
				attribute.String("key1", "value1"),
				attribute.String("key2", "value2"),
			},
		},
		{
			name: "int attributes",
			attrs: []attribute.KeyValue{
				attribute.Int("count", 42),
				attribute.Int("size", 1024),
			},
		},
		{
			name: "bool attributes",
			attrs: []attribute.KeyValue{
				attribute.Bool("success", true),
				attribute.Bool("cached", false),
			},
		},
		{
			name: "mixed attributes",
			attrs: []attribute.KeyValue{
				attribute.String("name", "test"),
				attribute.Int("count", 10),
				attribute.Bool("enabled", true),
			},
		},
		{
			name:  "empty attributes",
			attrs: []attribute.KeyValue{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetSpanAttributes(ctx, tt.attrs...)
			// Span should still be valid after setting attributes
			assert.True(t, span.IsRecording())
		})
	}
}

func TestSetSpanError(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "standard error",
			err:  assert.AnError,
		},
		{
			name: "nil error",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, span := tracer.Start(context.Background(), "test-span")
			defer span.End()

			SetSpanError(ctx, tt.err)
			assert.True(t, span.IsRecording())
		})
	}
}

func TestAddSpanEvent(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	tests := []struct {
		name  string
		event string
		attrs []attribute.KeyValue
	}{
		{
			name:  "simple event",
			event: "cache_hit",
			attrs: nil,
		},
		{
			name:  "event with attributes",
			event: "database_query",
			attrs: []attribute.KeyValue{
				attribute.String("query", "SELECT * FROM users"),
				attribute.Int("rows", 10),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddSpanEvent(ctx, tt.event, tt.attrs...)
			assert.True(t, span.IsRecording())
		})
	}
}

func TestTraceID(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")

	t.Run("with valid span", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		traceID := TraceID(ctx)
		assert.NotEmpty(t, traceID)
		assert.Len(t, traceID, 32) // 16 bytes hex encoded = 32 chars
	})

	t.Run("without span", func(t *testing.T) {
		ctx := context.Background()
		traceID := TraceID(ctx)
		assert.Empty(t, traceID)
	})
}

func TestSpanID(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")

	t.Run("with valid span", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		spanID := SpanID(ctx)
		assert.NotEmpty(t, spanID)
		assert.Len(t, spanID, 16) // 8 bytes hex encoded = 16 chars
	})

	t.Run("without span", func(t *testing.T) {
		ctx := context.Background()
		spanID := SpanID(ctx)
		assert.Empty(t, spanID)
	})
}

func TestInstrumentMCPRequest(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")

	tests := []struct {
		name   string
		method string
	}{
		{
			name:   "tools/list",
			method: "tools/list",
		},
		{
			name:   "tools/call",
			method: "tools/call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			ctx, span := InstrumentMCPRequest(ctx, tracer, tt.method)
			require.NotNil(t, span)
			defer span.End()

			// Verify trace and span IDs are set
			traceID := TraceID(ctx)
			spanID := SpanID(ctx)
			assert.NotEmpty(t, traceID)
			assert.NotEmpty(t, spanID)
			assert.True(t, span.IsRecording())
		})
	}
}

func TestInstrumentIndexerOperation(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")

	tests := []struct {
		name      string
		operation string
		path      string
	}{
		{
			name:      "successful indexing",
			operation: "index",
			path:      "/path/to/repo",
		},
		{
			name:      "search operation",
			operation: "search",
			path:      "/path/to/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			ctx, span := InstrumentIndexerOperation(ctx, tracer, tt.operation, tt.path)
			require.NotNil(t, span)
			defer span.End()

			traceID := TraceID(ctx)
			spanID := SpanID(ctx)
			assert.NotEmpty(t, traceID)
			assert.NotEmpty(t, spanID)
			assert.True(t, span.IsRecording())
		})
	}
}

func TestInstrumentEmbedding(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")

	tests := []struct {
		name       string
		provider   string
		textLength int
	}{
		{
			name:       "openai provider",
			provider:   "openai",
			textLength: 1000,
		},
		{
			name:       "local provider",
			provider:   "local",
			textLength: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			ctx, span := InstrumentEmbedding(ctx, tracer, tt.provider, tt.textLength)
			require.NotNil(t, span)
			defer span.End()

			traceID := TraceID(ctx)
			spanID := SpanID(ctx)
			assert.NotEmpty(t, traceID)
			assert.NotEmpty(t, spanID)
			assert.True(t, span.IsRecording())
		})
	}
}

func TestInstrumentVectorSearch(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")

	tests := []struct {
		name       string
		searchType string
		limit      int
	}{
		{
			name:       "semantic search",
			searchType: "semantic",
			limit:      10,
		},
		{
			name:       "hybrid search",
			searchType: "hybrid",
			limit:      20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			ctx, span := InstrumentVectorSearch(ctx, tracer, tt.searchType, tt.limit)
			require.NotNil(t, span)
			defer span.End()

			traceID := TraceID(ctx)
			spanID := SpanID(ctx)
			assert.NotEmpty(t, traceID)
			assert.NotEmpty(t, spanID)
			assert.True(t, span.IsRecording())
		})
	}
}

func TestSpanFromContext(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")

	t.Run("context with span", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		retrievedSpan := SpanFromContext(ctx)
		assert.NotNil(t, retrievedSpan)
		assert.True(t, retrievedSpan.IsRecording())
	})

	t.Run("context without span", func(t *testing.T) {
		ctx := context.Background()
		span := SpanFromContext(ctx)
		assert.NotNil(t, span) // Returns no-op span, not nil
		assert.False(t, span.IsRecording())
	})
}

func TestSamplingConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		samplingRate float64
	}{
		{"always sample", 1.0},
		{"never sample", 0.0},
		{"fifty percent", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := TracerConfig{
				Enabled:      false, // Use false to avoid connecting to OTLP endpoint
				ServiceName:  "test-service",
				Environment:  "test",
				SamplingRate: tt.samplingRate,
			}

			tp, err := NewTracerProvider(config)
			require.NoError(t, err)
			require.NotNil(t, tp)

			defer tp.Shutdown(context.Background())
		})
	}
}

func TestTracerProviderTracer(t *testing.T) {
	config := TracerConfig{
		Enabled:     true,
		ServiceName: "test-service",
	}

	tp, err := NewTracerProvider(config)
	require.NoError(t, err)
	require.NotNil(t, tp)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer()
	require.NotNil(t, tracer)

	// Use the tracer to create a span
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	assert.NotEmpty(t, TraceID(ctx))
}

func TestTracerProviderStartSpan(t *testing.T) {
	config := TracerConfig{
		Enabled:     true,
		ServiceName: "test-service",
	}

	tp, err := NewTracerProvider(config)
	require.NoError(t, err)
	require.NotNil(t, tp)
	defer tp.Shutdown(context.Background())

	ctx, span := tp.StartSpan(context.Background(), "test-operation")
	require.NotNil(t, span)
	defer span.End()

	assert.NotEmpty(t, TraceID(ctx))
	assert.NotEmpty(t, SpanID(ctx))
}
