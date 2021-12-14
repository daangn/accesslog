package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"

	"github.com/daangn/accesslog"
	"github.com/daangn/accesslog/middleware"
)

func main() {
	w, err := accesslog.NewFluentLogWriter("alpha", "0.0.0.0", 24224)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	r := chi.NewRouter()
	r.Use(middleware.AccessLog(accesslog.NewHTTPLogger(w, &accesslog.DefaultHTTPLogFormatter{})))
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		accesslog.AddToLogEntry(r.Context(), func(e *zerolog.Event) {
			e.Bytes("data", json.RawMessage(`{"foo": "bar"}`))
		})
		w.Write([]byte("pong"))
	})

	http.ListenAndServe(":3000", r)
}
