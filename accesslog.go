package accesslog

import (
	"context"
	"encoding/json"
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

	UserGetter       UserGetter
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
		UserGetter:       cfg.userGetter,
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
	NewLogEntry(r *http.Request, ww chi_middleware.WrapResponseWriter, userGetter UserGetter) LogEntry
}

// LogEntry is the interface for each log entry. It embeds zerolog.LogObjectMarshaler.
type LogEntry interface {
	zerolog.LogObjectMarshaler

	SetData(data json.RawMessage)
}

// LogEntryCtxKey is the context key for LogEntry.
var LogEntryCtxKey = struct{}{}

// SetLogData sets log data in LogEntry. the LogEntry is in context value.
func SetLogData(ctx context.Context, data json.RawMessage) context.Context {
	if len(data) == 0 {
		return ctx
	}

	if entry, ok := ctx.Value(LogEntryCtxKey).(LogEntry); ok {
		entry.SetData(data)
	}

	return ctx
}
