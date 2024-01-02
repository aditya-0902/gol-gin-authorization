package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const database = "auth"

var Client *mongo.Client = DBinstance()

func DBinstance() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error loading .env file")
	}
	MongoDb := os.Getenv("MONGODB_URL")
	if MongoDb == "" {
		MongoDb = "mongodb://localhost:27017"
	}
	fmt.Println(MongoDb)
	clientOpts := options.Client().ApplyURI(MongoDb)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal("error while connecting to mongo")
	}
	fmt.Println("mongo connection established")
	return client
}

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database(database).Collection(collectionName)
	return collection
}
