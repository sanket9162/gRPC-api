package mongodb

import (
	"context"
	"log"

	"github.com/sanket9162/grpc-api/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateMongoClient() (*mongo.Client, error) {
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(""))
	if err != nil {
		return nil, utils.ErrorHandler(err, "unable to connect to database")
	}

	log.Println("Connect to db")

	return client, nil
}
