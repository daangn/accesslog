# accesslog
accesslog provides access logs that capture detailed information abouut requests sent to your services. Each log contains information such as the time the request was created, the client's IP address, latencies, request paths, and server responses. You can use these access logs to analyze traffic patterns and troubleshoot issues.  

## Installation
```shell
GOPRIVATE=github.com/daangn go get -u github.com/daangn/accesslog 
```

## Getting started
Here's a basic usage of logging:

```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/daangn/accesslog"
	httpaccesslog "github.com/daangn/accesslog/http"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

func main() {
	r := chi.NewRouter()
	r.Use(httpaccesslog.Middleware())
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		accesslog.GetLogEntry(r.Context()).Add(func(e *zerolog.Event) {
			e.Bytes("data", json.RawMessage(`{"foo": "bar"}`))
		})
		w.Write([]byte("pong"))
	})

	http.ListenAndServe(":3000", r)
}

```

go run above code in your terminal, and then execute `curl localhost:3000/ping` in another terminal.
After, you can see some logs in your terminal like below.
```
{"remoteAddr":"[::1]:52151","path":"/ping","method":"GET","status":200,"ua":"curl/7.64.1","data":"{\"foo\": \"bar\"}","time":"2021-12-07T13:36:21.256115Z","dur(ms)":0.02387}
```

Check out the [examples](examples) for more!

## Log writer
In this library, the follwing log writers are available.

- stdout
- fluentd/fluent-bit

If you want one for yours, it's simple. Just implements the io.Writer.