package accesslog

import (
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

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
	headers      map[string]string
	withClientIP bool
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
	if le.isIgnored() {
		return
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

	if whs := le.cfg.headers; len(whs) != 0 {
		for h, a := range whs {
			if val := le.r.Header.Get(h); val != "" {
				n := h
				if a != "" {
					n = a
				}
				e.Str(n, val)
			}
		}
	}

	if le.cfg.withClientIP {
		if ip := clientIP(le.r.Header); ip != "" {
			e.Str("client-ip", ip)
		} else if ip, _, err := net.SplitHostPort(strings.TrimSpace(le.r.RemoteAddr)); err == nil {
			e.Str("client-ip", ip)
		}
	}

	for _, f := range le.add {
		f(e)
	}

	e.Send()
}

// isIgnored check whether a request path should be ignored
func (le *DefaultHTTPLogEntry) isIgnored() bool {
	if ips := le.cfg.ignoredPaths; len(ips) != 0 {
		for _, ignorePath := range ips[le.r.Method] {
			p := le.r.URL.Path
			if p[0] != '/' {
				p = "/" + p
			}
			if m, _ := path.Match(ignorePath, p); m {
				return true
			}
		}
	}
	return false
}
