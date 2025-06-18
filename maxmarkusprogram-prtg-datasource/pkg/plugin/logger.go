package plugin

import (
	"context"
	"fmt"
	"strings"
	"time"

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
func (l *prtgLogger) formatLogMessage(level, msg string, args ...any) string {
	timestamp := time.Now().Format("01-02|15:04:05")
	contextualArgs := log.ContextualAttributesFromContext(l.ctx)
	allArgs := l.sanitizeArgs(append(contextualArgs, args...))

	// Format key-value pairs
	kvPairs := ""
	for i := 0; i < len(allArgs)-1; i += 2 {
		kvPairs += fmt.Sprintf(" %v=%v", allArgs[i], allArgs[i+1])
	}

	// Always include logger name
	kvPairs = fmt.Sprintf(" logger=plugin.prtg-datasource%s", kvPairs)

	return fmt.Sprintf("%s[%s] %s%s", level, timestamp, msg, kvPairs)
}

func (l *prtgLogger) Debug(msg string, args ...any) {
	msg = l.capitalizeMessage(msg)
	formattedMsg := l.formatLogMessage("DEBUG", msg, args...)
	l.logger.Debug(formattedMsg)
}

func (l *prtgLogger) Info(msg string, args ...any) {
	msg = l.capitalizeMessage(msg)
	formattedMsg := l.formatLogMessage("INFO ", msg, args...)
	l.logger.Info(formattedMsg)
}

func (l *prtgLogger) Warn(msg string, args ...any) {
	msg = l.capitalizeMessage(msg)
	formattedMsg := l.formatLogMessage("WARN ", msg, args...)
	l.logger.Warn(formattedMsg)
}

func (l *prtgLogger) Error(msg string, args ...any) {
	msg = l.capitalizeMessage(msg)
	formattedMsg := l.formatLogMessage("ERROR", msg, args...)
	l.logger.Error(formattedMsg)
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
		// Sanitize key names to camelCase
		k = l.toCamelCase(k)
		args = append(args, k, v)
	}

	newCtx := log.WithContextualAttributes(l.ctx, args)
	return &prtgLogger{
		logger: l.logger,
		ctx:    newCtx,
	}
}

/* =================================== LOGGER HELPERS ======================================== */
func (l *prtgLogger) capitalizeMessage(msg string) string {
	if len(msg) == 0 {
		return msg
	}
	return strings.ToUpper(msg[:1]) + msg[1:]
}

func (l *prtgLogger) sanitizeArgs(args []any) []any {
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			// Convert keys to camelCase
			args[i] = l.toCamelCase(key)

			// Sanitize values if they're strings
			if val, ok := args[i+1].(string); ok {
				args[i+1] = l.SanitizeLogValue(val, 1000) // Max 1000 chars for string values
			}
		}
	}
	return args
}

func (l *prtgLogger) toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

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