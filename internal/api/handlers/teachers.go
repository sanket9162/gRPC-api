package handlers

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/sanket9162/grpc-api/internal/models"
	"github.com/sanket9162/grpc-api/internal/repositories/mongodb"
	pb "github.com/sanket9162/grpc-api/proto/gen"
	"go.mongodb.org/mongo-driver/v2/bson"
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
	filterTeacher(req)
	sortOptions(req.GetSortBy())

	return nil, nil
}

func filterTeacher(req *pb.GetTeachersRequest) {
	filter := bson.M{}

	var modelTeacher models.Teacher
	modelVal := reflect.ValueOf(&modelTeacher).Elem()
	modelType := modelVal.Type()

	reqVal := reflect.ValueOf(req.Teacher).Elem()
	reqType := reqVal.Type()

	for i := 0; i < reqVal.NumField(); i++ {
		fieldVal := reqVal.Field(i)
		fieldName := reqType.Field(i).Name

		if fieldVal.IsValid() && !fieldVal.IsZero() {
			modelField := modelVal.FieldByName(fieldName)
			if modelField.IsValid() && modelField.CanSet() {
				modelField.Set(fieldVal)
			}
		}
	}

	for i := 0; i < modelVal.NumField(); i++ {
		fieldVal := modelVal.Field(i)
		// fieldName := modelType.Field(i).Name

		if fieldVal.IsValid() && !fieldVal.IsZero() {
			bsonTag := modelType.Field(i).Tag.Get("bson")
			bsonTag = strings.TrimSuffix(bsonTag, ",omitempty")
			filter[bsonTag] = fieldVal.Interface().(string)
		}
	}

	fmt.Println("Filter:", filter)
}

func sortOptions(sortFields []*pb.SortField) bson.D {
	var sortOption bson.D

	for _, sortField := range sortFields {
		order := 1
		if sortField.GetOrder() == pb.Order_DESC {
			order = -1
		}
		sortOption = append(sortOption, bson.E{Key: sortField.Field, Value: order})
	}
	fmt.Println("Sort option :", sortOption)
	return sortOption
}
