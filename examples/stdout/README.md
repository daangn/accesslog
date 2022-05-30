# stdout
This example is a simple use case with standard output. In this scenario, we print the access log to standard output.

1. Run application.
```
go run main.go
```

2. Do curl.
```
curl localhost:3000/ping
```

Finally, you can get logs like below.
```json
{
  "protocol": "http",
  "path": "/ping",
  "status": "200",
  "ua": "curl/7.64.1",
  "time": "2021-12-09T13:02:53.950724Z",
  "elapsed(ms)": 0.113,
  "data": "{\"foo\": \"bar\"}"
}
```
