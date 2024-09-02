package db

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// ConnectDB connects to the MongoDB using the provided URI.
func ConnectDB(mongoURI string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    clientOptions := options.Client().ApplyURI(mongoURI)
    var err error
    client, err = mongo.Connect(ctx, clientOptions)
    if err != nil {
        return err
    }

    err = client.Ping(ctx, nil)
    if err != nil {
        return err
    }

    log.Println("Connected to MongoDB!")
    return nil
}

// GetCollection returns a MongoDB collection from the connected database
func GetCollection() *mongo.Collection {
    if client == nil {
        log.Fatal("MongoDB client is not initialized")
    }
    // Hardcoded database name and collection name
    return client.Database("pipeline_task").Collection("users")
}
