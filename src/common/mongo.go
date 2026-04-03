package common

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectMongoDB() *mongo.Database {
	client, _ := mongo.Connect(options.Client().ApplyURI(GetEnv("MONGO_URL")))
	fmt.Println("Mongo connected!")
	return client.Database(GetEnv("DB_NAME"))
}