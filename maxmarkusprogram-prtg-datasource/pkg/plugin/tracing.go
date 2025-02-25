package plugin

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Tracer struct {
	tracer trace.Tracer
	logger PrtgLogger
}

func NewTracer(logger PrtgLogger) *Tracer {
	return &Tracer{
		tracer: otel.Tracer("prtg-plugin"),
		logger: logger,
	}
}

func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	t.logger.Debug("Starting span", "name", name)
	ctx, span := t.tracer.Start(ctx, fmt.Sprintf("PRTG.%s", name))
	return ctx, span
}

func (t *Tracer) AddAttribute(span trace.Span, key string, value interface{}) {
	span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
}

/* =================================== TRACE HELPERS ======================================== */

// startTrace creates a new span and adds common attributes
func startTrace(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return tracing.DefaultTracer().Start(
		ctx,
		name,
		trace.WithAttributes(attrs...),
	)
}

// addAPIAttributes adds common API request attributes to a span
func addAPIAttributes(span trace.Span, method, endpoint string, params map[string]string) {
	attrs := []attribute.KeyValue{
		attribute.String("api.method", method),
		attribute.String("api.endpoint", endpoint),
	}

	for k, v := range params {
		attrs = append(attrs, attribute.String("api.param."+k, v))
	}

	span.SetAttributes(attrs...)
}

// recordError adds error details to a span
func recordError(span trace.Span, err error, message string) {
	span.RecordError(err)
	span.SetAttributes(
		attribute.String("error.message", err.Error()),
		attribute.String("error.type", fmt.Sprintf("%T", err)),
	)
	span.SetStatus(codes.Error, message)
}

/* =================================== API TRACING ======================================== */

// wrapAPICall wraps an API call with tracing
func wrapAPICall(ctx context.Context, name string, method string, params map[string]string, fn func() error) error {
	_, span := startTrace(ctx, fmt.Sprintf("prtg.api.%s", name),
		attribute.String("api.type", "prtg"),
	)
	defer span.End()

	addAPIAttributes(span, method, name, params)

	startTime := time.Now()
	err := fn()
	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Float64("duration_ms", float64(duration.Milliseconds())),
	)

	if err != nil {
		recordError(span, err, "API call failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

/* =================================== QUERY TRACING ======================================== */




// addQueryAttributes adds query-specific attributes to a span
func addQueryAttributes(span trace.Span, query queryModel) {
	attrs := []attribute.KeyValue{
		attribute.String("query.group", query.Group),
		attribute.String("query.device", query.Device),
		attribute.String("query.sensor", query.Sensor),
	}

	if len(query.Channels) > 0 {
		attrs = append(attrs, attribute.StringSlice("query.channels", query.Channels))
	}

	if query.Channel != "" {
		attrs = append(attrs, attribute.String("query.channel", query.Channel))
	}

	span.SetAttributes(attrs...)
}
