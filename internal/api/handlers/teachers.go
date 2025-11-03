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

func (s *Server) AddTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {

	for _, teacher := range req.GetTeachers() {
		if teacher.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "request is in incorrect format: non-empty ID fields are not allowed")
		}
	}

	addedTeacher, err := mongodb.AddTeachersToDb(ctx, req.GetTeachers())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Teachers{Teachers: addedTeacher}, nil

}

func (s *Server) GetTeachers(ctx context.Context, req *pb.GetTeachersRequest) (*pb.Teachers, error) {
	filter, err := filter(req.Teacher, &models.Teacher{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal err")
	}
	sortOption := sortOptions(req.GetSortBy())

	teachers, err := mongodb.GetTeachersFromDB(ctx, sortOption, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: teachers}, nil
}

func (s *Server) UpdateTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {

	updatedTeachers, err := mongodb.UpdateTeachersInDB(ctx, req.Teachers)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Teachers{Teachers: updatedTeachers}, nil
}

func (s *Server) DeleteExecs(ctx context.Context, req *pb.ExecIds) (*pb.DeleteExecsConfirmation, error) {
	ids := req.GetIds()
	var execsIdsToDelete []string
	for _, excesId := range ids {
		execsIdsToDelete = append(execsIdsToDelete, excesId.Id)
	}
	deletedIds, err := mongodb.DeleteTeacherFromDB(ctx, execsIdsToDelete)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteExecsConfirmation{
		Status:     "execs successfully deleted",
		DeletedIds: deletedIds,
	}, nil

}
