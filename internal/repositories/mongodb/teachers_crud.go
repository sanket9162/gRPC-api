package mongodb

import (
	"context"
	"fmt"
	"reflect"

	pb "github.com/sanket9162/grpc-api/proto/gen"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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
		newTeachers[i] = MappbTeacherToModelTeacler(pbTeacher)
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

		pbTeacher := MapModelTeacherTopb(teacher)
		addedTeacher = append(addedTeacher, pbTeacher)
	}

	return addedTeacher, nil
}

func GetTeachersFromDB(ctx context.Context, sortOption bson.D, filter bson.M) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("teacher")
	var cursor *mongo.Cursor
	if len(sortOption) < 1 {

		cursor, err = coll.Find(ctx, filter)
	} else {
		cursor, err = coll.Find(ctx, filter, options.Find().SetSort(sortOption))
	}
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}
	defer cursor.Close(ctx)

	teachers, err := decodeEntities(ctx, cursor, func() *pb.Teacher { return &pb.Teacher{} }, newModel)
	if err != nil {
		return nil, err
	}
	return teachers, nil
}

func newModel() *models.Teacher {
	return &models.Teacher{}
}

func MapModelTeacherTopb(teacher *models.Teacher) *pb.Teacher {
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

func MappbTeacherToModelTeacler(pbTeacher *pb.Teacher) *models.Teacher {
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

func UpdateTeachersInDB(ctx context.Context, pbTeachers []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal err")
	}
	defer client.Disconnect(ctx)

	var updatedTeachers []*pb.Teacher

	for _, teacher := range pbTeachers {
		modelTeacher := MappbTeacherToModelTeacler(teacher)

		objId, err := primitive.ObjectIDFromHex(teacher.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid Id")
		}

		modelDoc, err := bson.Marshal(modelTeacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}

		var updateDoc bson.M
		err = bson.Unmarshal(modelDoc, &updateDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}

		delete(updateDoc, "_id")

		_, err = client.Database("school").Collection("teacher").UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintln("error updating teacher id:", teacher.Id))
		}

		updatedTecher := MapModelTeacherTopb(modelTeacher)

		updatedTeachers = append(updatedTeachers, updatedTecher)
	}
	return updatedTeachers, nil
}
