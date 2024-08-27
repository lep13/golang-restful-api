package db

import (
    "context"
    "log"
    "os"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func ConnectDB() *mongo.Client {
    clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }
    return client
}

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
    db := client.Database(os.Getenv("DB_NAME")) 
    collection := db.Collection(collectionName)
    return collection
}
