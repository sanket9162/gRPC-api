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

func (s *Server) AddStudents(ctx context.Context, req *pb.Students) (*pb.Students, error) {

	for _, student := range req.GetStudents() {
		if student.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "request is in incorrect format: non-empty ID fields are not allowed")
		}
	}

	addedStudent, err := mongodb.AddStudentToDb(ctx, req.GetStudents())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Students{Students: addedStudent}, nil

}

func (s *Server) GetStudents(ctx context.Context, req *pb.GetStudentsRequest) (*pb.Students, error) {
	filter, err := filter(req.Student, &models.Student{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal err")
	}
	sortOption := sortOptions(req.GetSortBy())

	pageNumber := req.GetPageNumber()
	pageSize := req.GetPageSize()

	if pageNumber < 1 {
		pageNumber = 1
	}

	if pageSize < 2 {
		pageSize = 10
	}

	students, err := mongodb.GetStudentFromDB(ctx, sortOption, filter, int32(pageNumber), int32(pageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: students}, nil
}

func (s *Server) UpdateStudents(ctx context.Context, req *pb.Students) (*pb.Students, error) {

	updatedStudents, err := mongodb.UpdateStudentInDB(ctx, req.Students)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Students{Students: updatedStudents}, nil
}

func (s *Server) DeleteStudents(ctx context.Context, req *pb.StudentIds) (*pb.DeleteStudentsConfirmation, error) {
	// ids := req.GetIds()
	// var studentIdsToDelete []string
	// for _, v := range ids {
	// 	studentIdsToDelete = append(studentIdsToDelete, req.GetIds())
	// }
	deletedIds, err := mongodb.DeleteTeacherFromDB(ctx, req.GetIds())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteStudentsConfirmation{
		Status:     "students successfully deleted",
		DeletedIds: deletedIds,
	}, nil

}
