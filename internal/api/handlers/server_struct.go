package handlers

import pb "github.com/sanket9162/grpc-api/proto/gen"

type Server struct {
	pb.UnimplementedExecsServiceServer
	pb.UnimplementedStudentsServiceServer
	pb.UnimplementedTeachersServiceServer
}
