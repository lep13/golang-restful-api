package db

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDBClient is an interface to abstract the mongo.Client
type MongoDBClient interface {
    Ping(ctx context.Context, rp *readpref.ReadPref) error
    Database(name string, opts ...*options.DatabaseOptions) *mongo.Database
}

// Client is a variable to hold the actual mongo client
var Client MongoDBClient

// ConnectDB connects to the MongoDB using the provided URI.
func ConnectDB(mongoURI string) MongoDBClient {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    clientOptions := options.Client().ApplyURI(mongoURI)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }

    err = client.Ping(ctx, readpref.Primary())
    if err != nil {
        log.Fatalf("Failed to ping MongoDB: %v", err)
    }

    log.Println("Connected to MongoDB!")
    Client = client
    return client
}

// GetCollection returns a MongoDB collection from the "pipeline_task" database
func GetCollection(client MongoDBClient) *mongo.Collection {
    db := client.Database("pipeline_task")
    collection := db.Collection("users")
    return collection
}
