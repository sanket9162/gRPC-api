package handlers

import (
	"context"

	"github.com/sanket9162/grpc-api/internal/models"
	"github.com/sanket9162/grpc-api/internal/repositories/mongodb"
	pb "github.com/sanket9162/grpc-api/proto/gen"
	"github.com/sanket9162/grpc-api/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AddExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {
	for _, teacher := range req.GetExecs() {
		if teacher.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "request is in incorrect format: non-empty ID fields are not allowed")
		}
	}

	addedExec, err := mongodb.AddExecsToDb(ctx, req.GetExecs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Execs{Execs: addedExec}, nil
}

func (s *Server) GetExecs(ctx context.Context, req *pb.GetExecsRequest) (*pb.Execs, error) {
	filter, err := filter(req.Exec, &models.Exec{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal err")
	}
	sortOption := sortOptions(req.GetSortBy())

	execs, err := mongodb.GetExecsFromDB(ctx, sortOption, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: execs}, nil
}
