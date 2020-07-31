package grpc

import (
	"context"
	"log"
	"net"

	"os"
	"os/signal"

	v1 "github.com/ryan0906/Memos/pkg/api/v1"
	"google.golang.org/grpc"
)

// RunServer runs gRPC service to publish Memo service
func RunServer(ctx context.Context, v1API v1.MemoServiceServer, port string) error {
	endpoint, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	v1.RegisterMemoServiceServer(server, v1API)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Println("shutting down gRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	log.Println("starting gRPC server...")
	return server.Serve(endpoint)
}
