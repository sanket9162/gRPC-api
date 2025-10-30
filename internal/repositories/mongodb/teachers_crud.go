package mongodb

import (
	"context"
	"reflect"

	pb "github.com/sanket9162/grpc-api/proto/gen"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sanket9162/grpc-api/internal/models"
	"github.com/sanket9162/grpc-api/utils"
)

func AddTeachersToDb(ctx context.Context, teachersFromreq []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	newTeachers := make([]*models.Teacher, len(teachersFromreq))
	for i, pbTeacher := range teachersFromreq {
		newTeachers[i] = mappbTeacherToModelTeacler(pbTeacher)
	}

	var addedTeacher []*pb.Teacher

	for _, teacher := range newTeachers {
		result, err := client.Database("school").Collection("teacher").InsertOne(ctx, teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding value to database")
		}
		objectId, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			teacher.Id = objectId.Hex()
		}

		pbTeacher := mapModelTeacherTopb(teacher)
		addedTeacher = append(addedTeacher, pbTeacher)
	}

	return addedTeacher, nil
}

func mapModelTeacherTopb(teacher *models.Teacher) *pb.Teacher {
	pbTeacher := &pb.Teacher{}
	modelVal := reflect.ValueOf(*teacher)
	pbVal := reflect.ValueOf(pbTeacher).Elem()

	for i := 0; i < modelVal.NumField(); i++ {
		modelField := modelVal.Field(i)
		modelFieldType := modelVal.Type().Field(i)

		pbField := pbVal.FieldByName(modelFieldType.Name)
		if pbField.IsValid() && pbField.CanSet() {
			pbField.Set(modelField)
		}
	}
	return pbTeacher
}

func mappbTeacherToModelTeacler(pbTeacher *pb.Teacher) *models.Teacher {
	modelTeacher := models.Teacher{}
	pbVal := reflect.ValueOf(pbTeacher).Elem()
	modelVal := reflect.ValueOf(&modelTeacher).Elem()

	for i := 0; i < pbVal.NumField(); i++ {
		pbField := pbVal.Field(i)
		fieldName := pbVal.Type().Field(i).Name

		modelField := modelVal.FieldByName(fieldName)
		if modelField.IsValid() && modelField.CanSet() {
			modelField.Set(pbField)
		}

	}

	return &modelTeacher
}
