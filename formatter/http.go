package formatter

import (
	"net/http"

	chi_middleware "github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"

	"github.com/daangn/accesslog"
)

// DefaultHTTPLogFormatter is the default HTTP log formatter.
type DefaultHTTPLogFormatter struct{}

// NewLogEntry creates a new DefaultHTTPLogEntry.
func (f *DefaultHTTPLogFormatter) NewLogEntry(r *http.Request, ww chi_middleware.WrapResponseWriter) accesslog.LogEntry {
	return &DefaultHTTPLogEntry{
		r:            r,
		ww:           ww,
		addExtraFunc: []func(e *zerolog.Event){},
	}
}

// DefaultHTTPLogEntry is the log entry for HTTP request.
type DefaultHTTPLogEntry struct {
	r            *http.Request
	ww           chi_middleware.WrapResponseWriter
	addExtraFunc []func(e *zerolog.Event)
}

// Add adds Extra functions that add log fields.
func (le *DefaultHTTPLogEntry) Add(f func(e *zerolog.Event)) {
	le.addExtraFunc = append(le.addExtraFunc, f)
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (le *DefaultHTTPLogEntry) MarshalZerologObject(e *zerolog.Event) {
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
