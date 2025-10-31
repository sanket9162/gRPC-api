package handlers

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/sanket9162/grpc-api/internal/models"
	"github.com/sanket9162/grpc-api/internal/repositories/mongodb"
	pb "github.com/sanket9162/grpc-api/proto/gen"
	"github.com/sanket9162/grpc-api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	filter, err := filterTeacher(req.Teacher)
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

func filterTeacher(teacherObj *pb.Teacher) (bson.M, error) {
	filter := bson.M{}

	if teacherObj == nil {
		return filter, nil
	}

	var modelTeacher models.Teacher
	modelVal := reflect.ValueOf(&modelTeacher).Elem()
	modelType := modelVal.Type()

	reqVal := reflect.ValueOf(teacherObj).Elem()
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
			if bsonTag == "_id" {
				objID, err := primitive.ObjectIDFromHex(teacherObj.Id)
				if err != nil {
					return nil, utils.ErrorHandler(err, "Invalid Id")
				}
				filter[bsonTag] = objID
			} else {

				filter[bsonTag] = fieldVal.Interface().(string)
			}
		}
	}

	fmt.Println("Filter:", filter)
	return filter, nil
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
