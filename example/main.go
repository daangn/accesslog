package main

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/daangn/accesslog"
)

func main() {
	// Create a new Fluent log writer. It implements io.Writer.
	w, err := accesslog.NewFluentLogWriter("alpha", "0.0.0.0", 24224)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	r := chi.NewRouter()
	r.Use(accesslog.Middleware(accesslog.WithWriter(w)))
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		accesslog.SetLogData(r.Context(), []byte(`{"foo": "bar"}`))
		w.Write([]byte("pong"))
	})

	http.ListenAndServe(":3000", r)
}
