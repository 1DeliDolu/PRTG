package plugin

import (
	"context"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

/* =================================== LOGGER INTERFACE ======================================== */
type PrtgLogger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	WithContext(ctx context.Context) PrtgLogger
	WithFields(fields map[string]any) PrtgLogger
	SanitizeLogValue(value string, maxLength int) string
}

/* =================================== LOGGER IMPLEMENTATION ======================================== */
type prtgLogger struct {
	logger log.Logger
	ctx    context.Context
}

func NewLogger() PrtgLogger {
	return &prtgLogger{
		logger: backend.Logger,
		ctx:    context.Background(),
	}
}

/* =================================== LOGGER METHODS ======================================== */
func (l *prtgLogger) Debug(msg string, args ...any) {
	// Debug logging disabled to reduce terminal output
}

func (l *prtgLogger) Info(msg string, args ...any) {
	// Info logging disabled to reduce terminal output
}

func (l *prtgLogger) Warn(msg string, args ...any) {
	// Warning logging disabled to reduce terminal output
}

func (l *prtgLogger) Error(msg string, args ...any) {
	// Keep error logging for important issues
	l.logger.Error(msg, args...)
}

func (l *prtgLogger) WithContext(ctx context.Context) PrtgLogger {
	if ctx == nil {
		ctx = context.Background()
	}
	return &prtgLogger{
		logger: l.logger,
		ctx:    ctx,
	}
}

func (l *prtgLogger) WithFields(fields map[string]any) PrtgLogger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	newCtx := log.WithContextualAttributes(l.ctx, args)
	return &prtgLogger{
		logger: l.logger,
		ctx:    newCtx,
	}
}

/* =================================== LOGGER HELPERS ======================================== */
func (l *prtgLogger) SanitizeLogValue(value string, maxLength int) string {
	// Remove any potentially sensitive information
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")

	// Truncate if too long
	if len(value) > maxLength {
		return value[:maxLength] + "..."
	}
	return value
}

/* =================================== DEFAULT LOGGER ======================================== */
var Logger = NewLogger()