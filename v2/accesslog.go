package accesslog

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type LogEntry interface {
	Write(t time.Time)
	Add(func(e *zerolog.Event))
}

// LogEntryCtxKey is the context key for LogEntry.
var LogEntryCtxKey = struct{}{}

// GetLogEntry gets LogEntry in context.
func GetLogEntry(ctx context.Context) LogEntry {
	if le, ok := ctx.Value(LogEntryCtxKey).(LogEntry); ok {
		return le
	}
	return nil
}

// SetLogEntry sets LogEntry in context.
func SetLogEntry(ctx context.Context, le LogEntry) context.Context {
	return context.WithValue(ctx, LogEntryCtxKey, le)
}
