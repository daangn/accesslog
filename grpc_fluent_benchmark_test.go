package accesslog_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/daangn/accesslog"
	"github.com/daangn/accesslog/middleware"
	"github.com/daangn/accesslog/writer"
	"google.golang.org/grpc"
)

type benchmarkAuditWriter struct {
	w      io.Writer
	calls  int
	errors int
}

func (w *benchmarkAuditWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.calls++
	if err != nil || n != len(p) {
		w.errors++
	}
	return n, err
}

func BenchmarkGRPCUnaryFluentDefaultAsync(b *testing.B) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	drained := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			drained <- err
			return
		}
		defer conn.Close()
		n, err := io.Copy(io.Discard, conn)
		if err == nil && n == 0 {
			err = fmt.Errorf("forwarder sent no data")
		}
		drained <- err
	}()

	host, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		b.Fatal(err)
	}

	sink, err := writer.NewFluentLogWriter("alpha", host, port)
	if err != nil {
		b.Fatal(err)
	}
	defer sink.Close()
	audit := &benchmarkAuditWriter{w: sink}
	interceptor := middleware.UnaryServerInterceptor(accesslog.NewGRPCLogger(audit, accesslog.NewDefaultGRPCLogFormatter()))
	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/helloworld.Greeter/SayHello"}
	handler := func(context.Context, interface{}) (interface{}, error) { return nil, nil }

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := interceptor(ctx, nil, info, handler); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	if audit.calls != b.N || audit.errors != 0 {
		b.Fatalf("writes=%d expected=%d errors=%d", audit.calls, b.N, audit.errors)
	}
	if err := sink.Close(); err != nil {
		b.Fatal(err)
	}
	select {
	case err := <-drained:
		if err != nil {
			b.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		b.Fatal("forwarder did not finish draining")
	}
}
