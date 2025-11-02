package mongodb

import (
	"context"

	"github.com/sanket9162/grpc-api/internal/models"
	pb "github.com/sanket9162/grpc-api/proto/gen"
	"github.com/sanket9162/grpc-api/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		result, err := client.Database("school").Collection("execs").InsertOne(ctx, student)
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
