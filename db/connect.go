package db

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoClientInterface defines the interface for MongoDB client methods used in our code.
type MongoClientInterface interface {
    Ping(ctx context.Context, rp *readpref.ReadPref) error
    Database(name string, opts ...*options.DatabaseOptions) *mongo.Database
}

// MongoClientWrapper wraps the actual MongoDB client to conform to our interface.
type MongoClientWrapper struct {
    Client *mongo.Client
}

func (m *MongoClientWrapper) Ping(ctx context.Context, rp *readpref.ReadPref) error {
    return m.Client.Ping(ctx, rp)
}

func (m *MongoClientWrapper) Database(name string, opts ...*options.DatabaseOptions) *mongo.Database {
    return m.Client.Database(name, opts...)
}

// MongoClient holds the actual MongoDB client or a mock for testing.
var MongoClient MongoClientInterface

// MongoDatabaseInterface defines the interface for MongoDB database methods.
type MongoDatabaseInterface interface {
    Collection(name string, opts ...*options.CollectionOptions) MongoCollectionInterface
}

// MongoCollectionInterface defines the interface for MongoDB collection methods.
type MongoCollectionInterface interface {
    InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error)
    Find(ctx context.Context, filter interface{}) (*mongo.Cursor, error)
    FindOne(ctx context.Context, filter interface{}) MongoSingleResultInterface
    DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error)
    UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error)
}

// MongoSingleResultInterface defines the interface for MongoDB single result methods.
type MongoSingleResultInterface interface {
    Decode(v interface{}) error
}

// ConnectDB connects to the MongoDB using the provided URI.
func ConnectDB(mongoURI string) MongoClientInterface {
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
    MongoClient = &MongoClientWrapper{Client: client}
    return MongoClient
}

// GetCollection returns a MongoDB collection from the "pipeline_task" database
func GetCollection(client MongoClientInterface) *mongo.Collection {
    db := client.Database("pipeline_task")
    collection := db.Collection("users")
    return collection
}
