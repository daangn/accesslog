package reqlog

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	chi_middleware "github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

// Middleware returns middleware that will log incoming requests.
func Middleware(opts ...Option) func(next http.Handler) http.Handler {
	logger := NewLogger(opts...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chi_middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			entry := logger.HttpLogFormatter.NewLogEntry(r, ww, logger.UserGetter)

			t1 := time.Now()
			defer func() {
				logger.Write(entry, time.Since(t1))
			}()

			next.ServeHTTP(ww, RequestWithLogEntry(r, entry))
		})
	}
}

// DefaultHTTPLogFormatter is the default HTTP log formatter.
type DefaultHTTPLogFormatter struct{}

// NewLogEntry creates a new LogEntry.
func (f *DefaultHTTPLogFormatter) NewLogEntry(r *http.Request, ww chi_middleware.WrapResponseWriter, userGetter UserGetter) LogEntry {
	return &HTTPLogEntry{
		r:          r,
		ww:         ww,
		userGetter: userGetter,
	}
}

// HTTPLogEntry is the log entry for HTTP request.
type HTTPLogEntry struct {
	r          *http.Request
	ww         chi_middleware.WrapResponseWriter
	userGetter UserGetter
	data       json.RawMessage
}

// SetData sets data field of the HTTPLogEntry.
func (le *HTTPLogEntry) SetData(data json.RawMessage) {
	le.data = data
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (le *HTTPLogEntry) MarshalZerologObject(e *zerolog.Event) {
	e.Str("addr", le.r.Host).
		Str("path", le.r.URL.Path).
		Str("method", le.r.Method).
		Int("status", le.ww.Status()).
		Str("ua", le.r.UserAgent())

	if le.r.URL.RawQuery != "" {
		e.Str("qs", le.r.URL.RawQuery)
	}

	if le.userGetter != nil {
		userID := le.userGetter.GetUserID(le.r.Context())
		if userID != 0 {
			e.Int64("user", userID)
		}
	}

	if len(le.data) != 0 && json.Valid(le.data) {
		e.Str("data", string(le.data))
	}
}

// RequestWithLogEntry returns request that has a context with LogEntry.
func RequestWithLogEntry(r *http.Request, entry LogEntry) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), LogEntryCtxKey, entry))

	return r
}
