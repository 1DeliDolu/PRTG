package plugin

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Tracer struct {
	logger PrtgLogger
}

func NewTracer(logger PrtgLogger) *Tracer {
	return &Tracer{
		logger: logger,
	}
}

func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	t.logger.Debug("Starting span", "name", name)
	return tracing.DefaultTracer().Start(ctx, fmt.Sprintf("PRTG.%s", name))
}

func (t *Tracer) AddAttribute(span trace.Span, key string, value interface{}) {
	span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
}

/* =================================== TRACE HELPERS ======================================== */

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
	_, span := tracing.DefaultTracer().Start(ctx, fmt.Sprintf("prtg.api.%s", name),
		trace.WithAttributes(attribute.String("api.type", "prtg")),
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
		attribute.String("query.groupId", query.GroupId),
		attribute.String("query.device", query.Device),
		attribute.String("query.deviceId", query.DeviceId),
		attribute.String("query.sensor", query.Sensor),
		attribute.String("query.sensorId", query.SensorId),
	}

	if query.Channel != "" {
		attrs = append(attrs, attribute.String("query.channel", query.Channel))
	}

	if query.Property != "" {
		attrs = append(attrs, attribute.String("query.property", query.Property))
	}

	if query.FilterProperty != "" {
		attrs = append(attrs, attribute.String("query.filterProperty", query.FilterProperty))
	}

	span.SetAttributes(attrs...)
}
