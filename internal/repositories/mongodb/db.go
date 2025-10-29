package mongodb

import (
	"log"

	"github.com/sanket9162/grpc-api/utils"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateMongoClient() (*mongo.Client, error) {

	client, err := mongo.Connect(options.Client().ApplyURI("mongodb+srv://sanket:sanket@cluster0.hry3uhm.mongodb.net/"))
	if err != nil {
		return nil, utils.ErrorHandler(err, "unable to connect to database")
	}

	log.Println("Connect to db")

	return client, nil
}
