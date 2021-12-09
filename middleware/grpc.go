package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/daangn/accesslog"
)

// UnaryServerInterceptor will write access log to the given grpc server.
func UnaryServerInterceptor(logger *accesslog.GRPCLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		le := logger.NewLogEntry(ctx, req, res, info, err)

		t := time.Now().UTC()
		defer func() {
			le.Write(t)
		}()

		res, err = handler(accesslog.SetLogEntry(ctx, le), req)
		return
	}
}
