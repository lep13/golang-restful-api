package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Client {
	mongoURI := os.Getenv("MONGO_URI")

	// Check if the URI is set
	if mongoURI == "" {
		log.Fatal("MongoDB URI is not set in environment variables")
	}

	// Parse the connection URI
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Use context to manage the connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping the database to ensure the connection is established
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB!")
	return client
}

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	db := client.Database(os.Getenv("DB_NAME"))
	collection := db.Collection(collectionName)
	return collection
}
