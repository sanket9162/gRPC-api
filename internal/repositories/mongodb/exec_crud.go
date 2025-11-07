package mongodb

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-mail/mail/v2"
	"github.com/sanket9162/grpc-api/internal/models"
	pb "github.com/sanket9162/grpc-api/proto/gen"
	"github.com/sanket9162/grpc-api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddExecsToDb(ctx context.Context, execFromReq []*pb.Exec) ([]*pb.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	newExecs := make([]*models.Exec, len(execFromReq))
	for i, pbExec := range execFromReq {
		newExecs[i] = MappbExecToModelExec(pbExec)
		hashedPassword, err := utils.HashPassword(newExecs[i].Password)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}

		newExecs[i].Password = hashedPassword
		currentTime := time.Now().Format(time.RFC3339)
		newExecs[i].UserCreatedAt = currentTime
		newExecs[i].InactiveStatus = false
	}

	var addedExec []*pb.Exec

	for _, exec := range newExecs {
		result, err := client.Database("school").Collection("execs").InsertOne(ctx, exec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding value to database")
		}
		objectId, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			exec.Id = objectId.Hex()
		}

		pbExec := MapModelExecTopb(*exec)
		addedExec = append(addedExec, pbExec)
	}

	return addedExec, nil
}

func GetExecsFromDB(ctx context.Context, sortOption bson.D, filter bson.M) ([]*pb.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("exec")
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

	execs, err := DecodeEntities(ctx, cursor, func() *pb.Exec { return &pb.Exec{} }, func() *models.Exec {
		return &models.Exec{}
	})
	if err != nil {
		return nil, err
	}
	return execs, nil
}

func UpdateExecsInDB(ctx context.Context, pbExec []*pb.Exec) ([]*pb.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal err")
	}
	defer client.Disconnect(ctx)

	var updatedExecs []*pb.Exec

	for _, exec := range pbExec {
		modelExec := MappbExecToModelExec(exec)

		objId, err := primitive.ObjectIDFromHex(exec.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid Id")
		}

		modelDoc, err := bson.Marshal(modelExec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}

		var updateDoc bson.M
		err = bson.Unmarshal(modelDoc, &updateDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}

		delete(updateDoc, "_id")

		_, err = client.Database("school").Collection("exec").UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintln("error updating teacher id:", exec.Id))
		}

		updatedExec := MapModelExecTopb(*modelExec)

		updatedExecs = append(updatedExecs, updatedExec)
	}
	return updatedExecs, nil
}

func DeleteExecsFromDB(ctx context.Context, execsIdsToDelete []string) ([]string, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	objectIds := make([]primitive.ObjectID, len(execsIdsToDelete))
	for i, id := range execsIdsToDelete {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("incorrect id%v", id))
		}
		objectIds[i] = objectId
	}
	filter := bson.M{"_id": bson.M{"$in": objectIds}}
	result, err := client.Database("school").Collection("exces").DeleteMany(ctx, filter)
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

func GetUserByUsername(ctx context.Context, req *pb.ExecLoginRequest) (*models.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal err")
	}
	defer client.Disconnect(ctx)

	filter := bson.M{"username": req.GetUsername()}
	var exec models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, filter).Decode(&exec)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, utils.ErrorHandler(err, "user not found, Incorrect username/password")
		}
		return nil, utils.ErrorHandler(err, "internal error")
	}
	return &exec, nil
}

func UpdatePassowordInDB(ctx context.Context, req *pb.UpdatePasswordRequest) (string, string, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return "", "", utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	objId, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return "", "", utils.ErrorHandler(err, "invalid Id")
	}

	var user models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, bson.M{"_id": objId}).Decode(&user)
	if err != nil {
		return "", "", utils.ErrorHandler(err, "user not found")
	}

	err = utils.VerifyPassword(req.GetCurrentPassword(), user.Password)
	if err != nil {
		return "", "", utils.ErrorHandler(err, "internal error")
	}

	hashedPassword, err := utils.HashPassword(req.GetNewPassword())
	if err != nil {
		return "", "", utils.ErrorHandler(err, "Internal error")
	}

	update := bson.M{
		"$set": bson.M{
			"password":           hashedPassword,
			"password_change_at": time.Now().Format(time.RFC3339),
		},
	}

	_, err = client.Database("school").Collection("execs").UpdateOne(ctx, bson.M{"_id": objId}, update)
	if err != nil {
		return "", "", utils.ErrorHandler(err, "failed to update the passowrd")
	}
	return user.Username, user.Role, nil
}

func DeactivateUserInDB(ctx context.Context, ids []string) (*mongo.UpdateResult, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	var objectIds []primitive.ObjectID
	for _, id := range ids {
		objId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "invalid Id")
		}
		objectIds = append(objectIds, objId)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIds}}
	update := bson.M{"$set": bson.M{"inactive_status": true}}
	result, err := client.Database("school").Collection("execs").UpdateMany(ctx, filter, update)
	if err != nil {
		return nil, utils.ErrorHandler(err, "failed to deactivate users")
	}
	return result, nil
}

func ForgotpasswordDb(ctx context.Context, email string) (string, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return "", utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	var exec models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, bson.M{"email": email}).Decode(&exec)
	if err != nil {
		if err != nil {
			return "", utils.ErrorHandler(err, "user not found")
		}
		return "", utils.ErrorHandler(err, "internal error")
	}

	tokenbytes := make([]byte, 32)
	_, err = rand.Read(tokenbytes)
	if err != nil {
		return "", utils.ErrorHandler(err, "Failed to send password reset email")
	}

	token := hex.EncodeToString(tokenbytes)
	hashedToken := sha256.Sum256(tokenbytes)
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXP_DURATION"))
	if err != nil {
		return "", utils.ErrorHandler(err, "Failed to send password reset email")
	}
	mins := time.Duration(duration)
	expiry := time.Now().Add(mins * time.Minute).Format(time.RFC3339)

	update := bson.M{
		"$set": bson.M{
			"password_reset_token":   hashedTokenString,
			"password_token_expires": expiry,
		},
	}

	_, err = client.Database("school").Collection("execs").UpdateOne(ctx, bson.M{"email": email}, update)
	if err != nil {
		return "", utils.ErrorHandler(err, "internal error")
	}

	resetUrl := fmt.Sprintf("http://localhost:50052/execs/resetpassword/reset/%s", token)
	message := fmt.Sprintf("Forget your password? Reset your passowrd using the following link \n%s\n Please ust the reset code:: %s along with your request to change passwrod.  If you didn't request a password reset. plase ignore this email. This link is only valid for %v minutes.", resetUrl, token, mins)
	subject := "Your password reset link"

	m := mail.NewMessage()
	m.SetHeader("From", "schooladmin@school.com")
	m.SetHeader("To", email)
	m.SetHeader("subject", subject)
	m.SetBody("text/plain", message)

	d := mail.NewDialer("localhost", 1025, "", "")
	err = d.DialAndSend(m)
	if err != nil {
		cleanup := bson.M{
			"$set": bson.M{
				"password_reset_token":   hashedTokenString,
				"password_token_expires": expiry,
			},
		}
		_, _ = client.Database("school").Collection("execs").UpdateOne(ctx, bson.M{"email": email}, cleanup)
		return "", utils.ErrorHandler(err, "Could not send passowrd reset email. Please try again ")
	}
	return message, nil
}
