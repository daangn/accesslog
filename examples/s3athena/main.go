package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"

	"github.com/daangn/accesslog"
	httpaccesslog "github.com/daangn/accesslog/http"
	"github.com/daangn/accesslog/logwriter"
)

func main() {
	// Create a new Fluent log writer. It implements io.Writer.
	w, err := logwriter.NewFluentLogWriter("alpha", "0.0.0.0", 24224)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	r := chi.NewRouter()
	r.Use(httpaccesslog.Middleware(
		accesslog.WithWriter(w),
		accesslog.WithHTTPLogFormatter(&httpaccesslog.DefaultHTTPLogFormatter{}),
	))
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		accesslog.GetLogEntry(r.Context()).Add(func(e *zerolog.Event) {
			e.Bytes("data", json.RawMessage(`{"foo": "bar"}`))
		})
		w.Write([]byte("pong"))
	})

	http.ListenAndServe(":3000", r)
}
