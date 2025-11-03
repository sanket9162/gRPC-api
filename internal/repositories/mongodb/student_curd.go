package mongodb

import (
	"context"
	"fmt"

	"github.com/sanket9162/grpc-api/internal/models"

	pb "github.com/sanket9162/grpc-api/proto/gen"
	"github.com/sanket9162/grpc-api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddStudentToDb(ctx context.Context, studentFromReq []*pb.Student) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	newStudents := make([]*models.Student, len(studentFromReq))
	for i, pbStudent := range studentFromReq {
		newStudents[i] = MappbStudentToModelStudent(pbStudent)
	}

	var addedStudent []*pb.Student

	for _, student := range newStudents {
		result, err := client.Database("school").Collection("stucten").InsertOne(ctx, student)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding value to database")
		}
		objectId, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			student.Id = objectId.Hex()
		}

		pbStudent := MapModelStudentTopb(*student)
		addedStudent = append(addedStudent, pbStudent)
	}

	return addedStudent, nil
}

func GetStudentFromDB(ctx context.Context, sortOption bson.D, filter bson.M, pageNumber, pageSize int32) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("stucten")

	findOptions := options.Find()
	findOptions.SetSkip(int64((pageNumber - 1) * pageSize))
	findOptions.SetLimit(int64(pageSize))

	if len(sortOption) > 0 {
		findOptions.SetSort(sortOption)
	}
	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}
	defer cursor.Close(ctx)

	students, err := DecodeEntities(ctx, cursor, func() *pb.Student { return &pb.Student{} }, func() *models.Student {
		return &models.Student{}
	})
	if err != nil {
		return nil, err
	}
	return students, nil
}

func UpdateStudentInDB(ctx context.Context, pbstudent []*pb.Student) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal err")
	}
	defer client.Disconnect(ctx)

	var updatedStudents []*pb.Student

	for _, student := range pbstudent {
		modelStudent := MappbStudentToModelStudent(student)

		objId, err := primitive.ObjectIDFromHex(student.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid Id")
		}

		modelDoc, err := bson.Marshal(modelStudent)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}

		var updateDoc bson.M
		err = bson.Unmarshal(modelDoc, &updateDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}

		delete(updateDoc, "_id")

		_, err = client.Database("school").Collection("stucten").UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintln("error updating teacher id:", student.Id))
		}

		updatedStudent := MapModelStudentTopb(*modelStudent)

		updatedStudents = append(updatedStudents, updatedStudent)
	}
	return updatedStudents, nil
}

func DeleteStudentFromDB(ctx context.Context, studentIdsToDelete []string) ([]string, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	objectIds := make([]primitive.ObjectID, len(studentIdsToDelete))
	for i, id := range studentIdsToDelete {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("incorrect id%v", id))
		}
		objectIds[i] = objectId
	}
	filter := bson.M{"_id": bson.M{"$in": objectIds}}
	result, err := client.Database("school").Collection("stucten").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}

	if result.DeletedCount == 0 {
		return nil, utils.ErrorHandler(err, "no students were deleted")
	}

	deletedIds := make([]string, result.DeletedCount)
	for i, id := range objectIds {
		deletedIds[i] = id.Hex()
	}
	return deletedIds, nil
}
