package mongodb

import (
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateMongoClient() (*mongo.Client, error) {

	client, err := mongo.Connect(options.Client().ApplyURI("mongodb+srv://sanket:sanket@cluster0.hry3uhm.mongodb.net/"))
	if err != nil {
		log.Println("Error connecting to MongoDB.", err)
		return nil, err
	}

	log.Println("Connect to db")

	return client, nil
}
