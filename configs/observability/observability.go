package observability

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/ares/dp-vc-webApp/configs/env"
)

var tracer trace.Tracer

// Init initializes OpenTelemetry tracing
func Init(config *env.Config) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var tp *sdktrace.TracerProvider

	// Only configure OTLP exporter if endpoint is provided
	if config.OTLPEndpoint != "" {
		// Create OTLP exporter
		client := otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(config.OTLPEndpoint),
		)

		exporter, err := otlptrace.New(ctx, client)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}

		// Create tracer provider with exporter
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
		)
	} else {
		// Create no-op tracer provider for local development
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
		)
	}

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer = tp.Tracer(config.ServiceName)

	return tp, nil
}

// Shutdown gracefully shuts down the tracer provider
func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider) {
	if tp != nil {
		_ = tp.Shutdown(ctx)
	}
}

// GetTracer returns the global tracer
func GetTracer() trace.Tracer {
	return tracer
}

// Middleware returns a Gin middleware for OpenTelemetry tracing
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if tracer == nil {
			c.Next()
			return
		}

		// Extract trace context from incoming request
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start a new span
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		if spanName == " " {
			spanName = c.Request.URL.Path
		}

		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPRoute(c.FullPath()),
				semconv.HTTPURL(c.Request.URL.String()),
			),
		)
		defer span.End()

		// Set context and continue
		c.Request = c.Request.WithContext(ctx)
		c.Next()

		// Record status code
		status := c.Writer.Status()
		span.SetAttributes(
			semconv.HTTPStatusCode(status),
			attribute.Int64("http.request.body.size", c.Request.ContentLength),
		)

		if status >= 400 {
			span.RecordError(fmt.Errorf("HTTP %d", status))
		}
	}
}

// RecordError records an error in the current span
func RecordError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	span := oteltrace.SpanFromContext(c.Request.Context())
	if !span.SpanContext().IsValid() {
		return
	}

	span.RecordError(err)
}

// AddAttribute adds an attribute to the current span
func AddAttribute(c *gin.Context, key string, value interface{}) {
	span := oteltrace.SpanFromContext(c.Request.Context())
	if !span.SpanContext().IsValid() {
		return
	}

	switch v := value.(type) {
	case string:
		span.SetAttributes(attribute.String(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	default:
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
	}
}

// StartSpan starts a new manual span for custom operations
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name)
}

// TraceHTTPRequest traces an HTTP client request
func TraceHTTPRequest(ctx context.Context, req *http.Request) ([]func(*sdktrace.TracerProvider), error) {
	return nil, nil
}
