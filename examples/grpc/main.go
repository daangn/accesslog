package main

import (
	"context"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/daangn/accesslog"
	"github.com/daangn/accesslog/middleware"
)

type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	srv := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 15 * time.Second,
			Time:                  15 * time.Second,
			Timeout:               10 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(
			middleware.UnaryServerInterceptor(accesslog.NewGRPCLogger(os.Stdout, accesslog.NewDefaultGRPCLogFormatter(
				accesslog.WithRequestField(),
				accesslog.WithMetadataField(),
				accesslog.WithIgnoredMetadata("content-type"),
			))),
		),
	)
	reflection.Register(srv)
	pb.RegisterGreeterServer(srv, &server{})

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
