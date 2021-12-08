package main

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/daangn/accesslog"
	httpaccesslog "github.com/daangn/accesslog/http"
)

func main() {
	r := chi.NewRouter()
	r.Use(httpaccesslog.Middleware())
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		accesslog.SetLogData(r.Context(), []byte(`{"foo": "bar"}`))
		w.Write([]byte("pong"))
	})

	http.ListenAndServe(":3000", r)
}
