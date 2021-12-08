package middleware

import (
	"net/http"
	"time"

	chi_middleware "github.com/go-chi/chi/middleware"

	"github.com/daangn/accesslog"
)

// AccessLog returns middleware that will log incoming requests.
func AccessLog(opts ...accesslog.Option) func(next http.Handler) http.Handler {
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

// RequestWithLogEntry returns request that has a context with DefaultHTTPLogEntry.
func RequestWithLogEntry(r *http.Request, le accesslog.LogEntry) *http.Request {
	r = r.WithContext(accesslog.SetLogEntry(r.Context(), le))

	return r
}
