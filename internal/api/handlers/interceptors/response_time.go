package interceptors

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ResponseTimeInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	log.Println("Response Time Interceptor run")
	start := time.Now()

	resp, err := handler(ctx, req)

	duration := time.Since(start)

	st, _ := status.FromError(err)
	fmt.Printf("Method: %s, Status: %d, Duration: %v\n", info.FullMethod, st.Code(), duration)

	md := metadata.Pairs("X-Tesponse-Time", duration.String())
	grpc.SetHeader(ctx, md)

	log.Println("sending response from Response time interceptor ")
	return resp, err
}
