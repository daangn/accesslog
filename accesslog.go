package accesslog

import (
	"context"
	"net/http"
	"time"

	chi_middleware "github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
}

// Logger is the request logger.
type Logger struct {
	zerolog.Logger

	HttpLogFormatter HTTPLogFormatter
}

// NewLogger creates a new logger.
func NewLogger(opts ...Option) *Logger {
	cfg := new(config)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}

	return &Logger{
		Logger:           zerolog.New(cfg.writer),
		HttpLogFormatter: cfg.httpLogFormatter,
	}
}

// Write writes a log.
func (l *Logger) Write(le LogEntry, t time.Time) {
	l.Log().
		EmbedObject(le).
		Time("time", t).
		Dur("dur(ms)", time.Since(t)).
		Send()
}

// HTTPLogFormatter is the interface for the NewLogEntry method.
type HTTPLogFormatter interface {
	NewLogEntry(r *http.Request, ww chi_middleware.WrapResponseWriter) LogEntry
}

// LogEntryCtxKey is the context key for LogEntry.
var LogEntryCtxKey = struct{}{}

// LogEntry is the interface for each log entry. It embeds zerolog.LogObjectMarshaler.
type LogEntry interface {
	zerolog.LogObjectMarshaler

	Add(func(e *zerolog.Event))
}

// GetLogEntry gets LogEntry in context.
func GetLogEntry(ctx context.Context) LogEntry {
	if le, ok := ctx.Value(LogEntryCtxKey).(LogEntry); ok {
		return le
	}
	return nil
}

func SetLogEntry(ctx context.Context, le LogEntry) context.Context {
	return context.WithValue(ctx, LogEntryCtxKey, le)
}
