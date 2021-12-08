package http

import (
	"net/http"
	"time"

	chi_middleware "github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"

	"github.com/daangn/accesslog"
)

// Middleware returns middleware that will log incoming requests.
func Middleware(opts ...accesslog.Option) func(next http.Handler) http.Handler {
	logger := accesslog.NewLogger(opts...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chi_middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			entry := logger.HttpLogFormatter.NewLogEntry(r, ww)

			t := time.Now().UTC()
			defer func() {
				logger.Write(entry, t)
			}()

			next.ServeHTTP(ww, RequestWithLogEntry(r, entry))
		})
	}
}

// DefaultHTTPLogFormatter is the default HTTP log formatter.
type DefaultHTTPLogFormatter struct{}

// NewLogEntry creates a new LogEntry.
func (f *DefaultHTTPLogFormatter) NewLogEntry(r *http.Request, ww chi_middleware.WrapResponseWriter) accesslog.LogEntry {
	return &LogEntry{
		r:            r,
		ww:           ww,
		addExtraFunc: []func(e *zerolog.Event){},
	}
}

// LogEntry is the log entry for HTTP request.
type LogEntry struct {
	r            *http.Request
	ww           chi_middleware.WrapResponseWriter
	addExtraFunc []func(e *zerolog.Event)
}

func (le *LogEntry) Add(f func(e *zerolog.Event)) {
	le.addExtraFunc = append(le.addExtraFunc, f)
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (le *LogEntry) MarshalZerologObject(e *zerolog.Event) {
	e.Str("remoteAddr", le.r.RemoteAddr).
		Str("path", le.r.URL.Path).
		Str("method", le.r.Method).
		Int("status", le.ww.Status()).
		Str("ua", le.r.UserAgent())

	if val := le.r.Header.Get("authority"); val != "" {
		e.Str("authority", val)
	}
	if val := le.r.Header.Get("X-Forwarded-For"); val != "" {
		e.Str("X-Forwarded-For", val)
	}
	if val := le.r.Header.Get("X-Envoy-External-Address"); val != "" {
		e.Str("X-Envoy-External-Address", val)
	}
	if val := le.r.Header.Get("X-Request-ID"); val != "" {
		e.Str("X-Request-ID", val)
	}

	if le.r.URL.RawQuery != "" {
		e.Str("qs", le.r.URL.RawQuery)
	}

	for _, f := range le.addExtraFunc {
		f(e)
	}
}

// RequestWithLogEntry returns request that has a context with LogEntry.
func RequestWithLogEntry(r *http.Request, le accesslog.LogEntry) *http.Request {
	r = r.WithContext(accesslog.SetLogEntry(r.Context(), le))

	return r
}
