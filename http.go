package accesslog

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	chi_middleware "github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

// DefaultHTTPLogger is default HTTP Logger.
var DefaultHTTPLogger = NewHTTPLogger(os.Stdout, &DefaultHTTPLogFormatter{})

// HTTPLogger is logger for HTTP access logging.
type HTTPLogger struct {
	l *zerolog.Logger
	f HTTPLogFormatter
}

// NewHTTPLogger returns a new HTTPLogger.
func NewHTTPLogger(w io.Writer, f HTTPLogFormatter) *HTTPLogger {
	l := zerolog.New(w)
	return &HTTPLogger{
		l: &l,
		f: f,
	}
}

// NewLogEntry returns a New LogEntry.
func (l *HTTPLogger) NewLogEntry(r *http.Request, ww middleware.WrapResponseWriter) LogEntry {
	return l.f.NewLogEntry(l.l, r, ww)
}

// HTTPLogFormatter is the interface for NewLogEntry method.
type HTTPLogFormatter interface {
	NewLogEntry(l *zerolog.Logger, r *http.Request, ww middleware.WrapResponseWriter) LogEntry
}

// DefaultHTTPLogFormatter is default HTTPLogFormatter.
type DefaultHTTPLogFormatter struct{}

// NewLogEntry returns a New LogEntry formatted in DefaultHTTPLogFormatter.
func (f *DefaultHTTPLogFormatter) NewLogEntry(l *zerolog.Logger, r *http.Request, ww middleware.WrapResponseWriter) LogEntry {
	return &DefaultHTTPLogEntry{
		l:   l,
		r:   r,
		ww:  ww,
		add: []func(e *zerolog.Event){},
	}
}

// DefaultHTTPLogEntry is the LogEntry formatted in DefaultHTTPLogFormatter.
type DefaultHTTPLogEntry struct {
	l   *zerolog.Logger
	r   *http.Request
	ww  chi_middleware.WrapResponseWriter
	add []func(e *zerolog.Event)
}

// Add adds function for adding fields to log event.
func (le *DefaultHTTPLogEntry) Add(f func(e *zerolog.Event)) {
	le.add = append(le.add, f)
}

// Write writes a log.
func (le *DefaultHTTPLogEntry) Write(t time.Time) {
	e := le.l.Log().
		Str("protocol", "http").
		Str("path", le.r.URL.Path).
		Str("status", strconv.Itoa(le.ww.Status())).
		Str("ua", le.r.UserAgent()).
		Str("time", t.UTC().Format(time.RFC3339Nano)).
		Dur("elapsed(ms)", time.Since(t))

	if val := le.r.Header.Get("authority"); val != "" {
		e.Str("authority", val)
	}
	if val := le.r.Header.Get("X-Envoy-External-Address"); val != "" {
		e.Str("X-Envoy-External-Address", val)
	}
	if val := le.r.Header.Get("X-Request-ID"); val != "" {
		e.Str("X-Request-ID", val)
	}
	if val := le.r.URL.RawQuery; val != "" {
		e.Str("qs", val)
	}

	for _, f := range le.add {
		f(e)
	}

	e.Send()
}
