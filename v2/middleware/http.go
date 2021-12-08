package middleware

import (
	"net/http"
	"time"

	chi_middleware "github.com/go-chi/chi/middleware"

	"github.com/daangn/accesslog/v2"
)

func AccessLog(logger *accesslog.HTTPLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chi_middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			le := logger.NewLogEntry(r, ww)

			t := time.Now().UTC()
			defer func() {
				le.Write(t)
			}()

			next.ServeHTTP(ww, RequestWithLogEntry(r, le))
		})
	}
}

// RequestWithLogEntry returns request that has a context with accesslog.HTTPLogEntry.
func RequestWithLogEntry(r *http.Request, le accesslog.LogEntry) *http.Request {
	r = r.WithContext(accesslog.SetLogEntry(r.Context(), le))

	return r
}
