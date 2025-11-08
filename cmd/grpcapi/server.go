package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sanket9162/grpc-api/internal/api/handlers"
	"github.com/sanket9162/grpc-api/internal/api/handlers/interceptors"
	"github.com/sanket9162/grpc-api/internal/repositories/mongodb"
	pb "github.com/sanket9162/grpc-api/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	mongodb.CreateMongoClient()

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors.ResponseTimeInterceptor, interceptors.NewRateLimiter(5, time.Minute).RateLimiterInterceprot, interceptors.AuthenticationInterceptor))

	pb.RegisterExecsServiceServer(s, &handlers.Server{})
	pb.RegisterStudentsServiceServer(s, &handlers.Server{})
	pb.RegisterTeachersServiceServer(s, &handlers.Server{})

	reflection.Register(s)

	port := os.Getenv("SERVER_PORT")

	fmt.Println("gRPC Server is running on port:", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error listening on specified port:", err)
	}

	err = s.Serve(lis)
	if err != nil {
		log.Fatal("Failed to serve:", err)
	}

}
