# grpc
This example is a simple use case with standard output. In this scenario, we print access logs to standard output.

1. Run application.
```
go run main.go
```

2. Do [grpcurl](https://github.com/fullstorydev/grpcurl).
```
grpcurl -v -plaintext -d '{"name":"karrot"}' localhost:8080 helloworld.Greeter/SayHello
```

Finally, you can get logs like below.
```json
{
  "protocol": "grpc",
  "method": "/helloworld.Greeter/SayHello",
  "status": "OK",
  "time": "2021-12-09T13:02:09.644628Z",
  "elapsed(ms)": 0.013,
  "peer": "[::1]:49960",
  "metadata": "{\":authority\":[\"localhost:8080\"],\"user-agent\":[\"grpcurl/1.8.2 grpc-go/1.37.0\"]}",
  "req": "{\"name\":\"karrot\"}",
  "res": "{\"message\":\"Hello karrot\"}"
}
```
