package accesslog

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	chi_middleware "github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

// DefaultHTTPLogger is default HTTP Logger.
var DefaultHTTPLogger = NewHTTPLogger(os.Stdout, NewDefaultHTTPLogFormatter())

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

type httpConfig struct {
	ignoredPaths map[string][]string
	Headers      map[string]struct{}
}

type httpOption func(cfg *httpConfig)

// WithIgnoredPaths specifies methods and paths to be captured by the logger.
// This only works when using chi.Router.
func WithIgnoredPaths(ips map[string][]string) httpOption {
	return func(cfg *httpConfig) {
		cfg.ignoredPaths = ips
	}
}

// WithHeaders specifies headers to be captured by the logger.
func WithHeaders(hs ...string) httpOption {
	whs := make(map[string]struct{}, len(hs))
	for _, e := range hs {
		whs[e] = struct{}{}
	}
	return func(cfg *httpConfig) {
		cfg.Headers = whs
	}
}

// DefaultHTTPLogFormatter is default HTTPLogFormatter.
type DefaultHTTPLogFormatter struct {
	cfg *httpConfig
}

// NewDefaultHTTPLogFormatter returns a new DefaultHTTPLogFormatter.
func NewDefaultHTTPLogFormatter(opts ...httpOption) *DefaultHTTPLogFormatter {
	cfg := new(httpConfig)
	for _, fn := range opts {
		fn(cfg)
	}

	return &DefaultHTTPLogFormatter{cfg: cfg}
}

// NewLogEntry returns a New LogEntry formatted in DefaultHTTPLogFormatter.
func (f *DefaultHTTPLogFormatter) NewLogEntry(l *zerolog.Logger, r *http.Request, ww middleware.WrapResponseWriter) LogEntry {
	return &DefaultHTTPLogEntry{
		cfg: f.cfg,
		l:   l,
		r:   r,
		ww:  ww,
		add: []func(e *zerolog.Event){},
	}
}

// DefaultHTTPLogEntry is the LogEntry formatted in DefaultHTTPLogFormatter.
type DefaultHTTPLogEntry struct {
	cfg *httpConfig
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
	if ips := le.cfg.ignoredPaths; len(ips) != 0 {
		rctx := chi.RouteContext(le.r.Context())
		for m, ps := range ips {
			for _, p := range ps {
				if rctx.Routes.Match(rctx, m, p) {
					return
				}
			}
		}
	}

	e := le.l.Log().
		Str("protocol", "http").
		Str("path", le.r.URL.Path).
		Str("status", strconv.Itoa(le.ww.Status())).
		Str("ua", le.r.UserAgent()).
		Str("time", t.UTC().Format(time.RFC3339Nano)).
		Dur("elapsed(ms)", time.Since(t))

	if val := le.r.URL.RawQuery; val != "" {
		e.Str("qs", val)
	}

	if whs := le.cfg.Headers; len(whs) != 0 {
		for h := range whs {
			if val := le.r.Header.Get(h); val != "" {
				e.Str(h, val)
			}
		}
	}

	for _, f := range le.add {
		f(e)
	}

	e.Send()
}
