/*
Package accesslog contains a library access logging.
*/
package accesslog

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// LogEntry is the interface for each log entry.
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

// AddToLogEntry add values to LogEntry that is from context.
// If LogEntry is nil, it doesn't do anything.
func AddToLogEntry(ctx context.Context, f func(e *zerolog.Event)) {
	if le := GetLogEntry(ctx); le != nil {
		le.Add(f)
	}
}
