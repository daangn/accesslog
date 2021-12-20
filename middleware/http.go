package middleware

import (
	"net/http"
	"time"

	chi_middleware "github.com/go-chi/chi/v5/middleware"

	"github.com/daangn/accesslog"
)

// AccessLog returns middleware that will log incoming requests.
func AccessLog(logger *accesslog.HTTPLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chi_middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			entry := logger.NewLogEntry(r, ww)

			t := time.Now().UTC()
			defer func() {
				entry.Write(t)
			}()

			next.ServeHTTP(ww, RequestWithLogEntry(r, entry))
		})
	}
}

// RequestWithLogEntry returns request that has a context with accesslog.HTTPLogEntry.
func RequestWithLogEntry(r *http.Request, le accesslog.LogEntry) *http.Request {
	r = r.WithContext(accesslog.SetLogEntry(r.Context(), le))

	return r
}
