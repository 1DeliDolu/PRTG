package plugin

import (
	"context"
	"fmt"
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
