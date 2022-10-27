package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client = DBConnect()
var Database string = os.Getenv("MONGODB_DB")

func DBConnect() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading the dotenv file.\nerror: %v\n", err.Error())
	}

	MongoDBURL := os.Getenv("MONGODB_URL") + os.Getenv("MONGODB_DB")

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDBURL))
	if err != nil {
		log.Fatalf("Error creating new client.\nerror: %v\n", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Error connecting to the mongodb database.\nerror: %v\n", err.Error())
	}

	return client
}

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database(Database).Collection(collectionName)
}
