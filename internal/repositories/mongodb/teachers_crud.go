package mongodb

import (
	"context"
	"fmt"

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
		newTeachers[i] = MappbTeacherToModelTeacher(pbTeacher)
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

		pbTeacher := MapModelTeacherTopb(*teacher)
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

	teachers, err := DecodeEntities(ctx, cursor, func() *pb.Teacher { return &pb.Teacher{} }, func() *models.Teacher {
		return &models.Teacher{}
	})
	if err != nil {
		return nil, err
	}
	return teachers, nil
}

func UpdateTeachersInDB(ctx context.Context, pbTeachers []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal err")
	}
	defer client.Disconnect(ctx)

	var updatedTeachers []*pb.Teacher

	for _, teacher := range pbTeachers {
		modelTeacher := MappbTeacherToModelTeacher(teacher)

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

		updatedTecher := MapModelTeacherTopb(*modelTeacher)

		updatedTeachers = append(updatedTeachers, updatedTecher)
	}
	return updatedTeachers, nil
}

func DeleteTeacherFromDB(ctx context.Context, teacherIdsToDelete []string) ([]string, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	objectIds := make([]primitive.ObjectID, len(teacherIdsToDelete))
	for i, id := range teacherIdsToDelete {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("incorrect id%v", id))
		}
		objectIds[i] = objectId
	}
	filter := bson.M{"_id": bson.M{"$in": objectIds}}
	result, err := client.Database("school").Collection("teacher").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}

	if result.DeletedCount == 0 {
		return nil, utils.ErrorHandler(err, "no teachers were deleted")
	}

	deletedIds := make([]string, result.DeletedCount)
	for i, id := range objectIds {
		deletedIds[i] = id.Hex()
	}
	return deletedIds, nil
}

func GetStudentCountByTeacherIDFromDB(ctx context.Context, teacherId string) (int64, error) {

	client, err := CreateMongoClient()
	if err != nil {
		return 0, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	objId, err := primitive.ObjectIDFromHex(teacherId)
	if err != nil {
		return 0, utils.ErrorHandler(err, "Invalid Teacher Id")
	}

	var teacher models.Teacher
	err = client.Database("school").Collection("teacher").FindOne(ctx, bson.M{"_id": objId}).Decode(&teacher)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, utils.ErrorHandler(err, "teacher not found")
		}
		return 0, utils.ErrorHandler(err, "internal error")
	}

	count, err := client.Database("school").Collection("stucten").CountDocuments(ctx, bson.M{"class": teacher.Class})
	if err != nil {
		return 0, utils.ErrorHandler(err, "internal error")
	}

	return count, nil
}

func GetStudentByTeacherIdFromDb(ctx context.Context, teacherId string) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	objId, err := primitive.ObjectIDFromHex(teacherId)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Invalid Teacher Id")
	}

	var teacher models.Teacher
	err = client.Database("school").Collection("teacher").FindOne(ctx, bson.M{"_id": objId}).Decode(&teacher)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, utils.ErrorHandler(err, "teacher not found")
		}
		return nil, utils.ErrorHandler(err, "internal error")
	}

	cursor, err := client.Database("school").Collection("stucten").Find(ctx, bson.M{"class": teacher.Class})
	defer cursor.Close(ctx)

	student, err := DecodeEntities(ctx, cursor, func() *pb.Student { return &pb.Student{} }, func() *models.Student { return &models.Student{} })
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}

	err = cursor.Err()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}

	return student, nil

}
