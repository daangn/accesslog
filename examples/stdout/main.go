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
	r := chi.NewRouter()
	r.Use(middleware.AccessLog(accesslog.DefaultHTTPLogger))
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		accesslog.GetLogEntry(r.Context()).Add(func(e *zerolog.Event) {
			e.Bytes("data", json.RawMessage(`{"foo": "bar"}`))
		})
		w.Write([]byte("pong"))
	})

	http.ListenAndServe(":3000", r)
}
