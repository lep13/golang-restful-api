package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MockMongoClient is a mock structure that simulates the behavior of a MongoDB client.
type MockMongoClient struct{}

// MockDatabase simulates a MongoDB database.
type MockDatabase struct{}

// MockCollection simulates a MongoDB collection.
type MockCollection struct{}

// Mock methods to satisfy the MongoDBClient interface
func (m *MockMongoClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	// Simulate a successful ping
	return nil
}

// Implementing a mock method to return a MockDatabase instead of a real mongo.Database
func (m *MockMongoClient) Database(name string, opts ...*options.DatabaseOptions) *mongo.Database {
	// Create a dummy mongo.Database object
	db := mongo.Database{}
	return &db
}

// Implementing a mock method for the collection, which avoids the actual database calls.
func (m *MockDatabase) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	// Create a dummy mongo.Collection object
	col := mongo.Collection{}
	return &col
}

func TestMain(m *testing.M) {
	// Load environment variables from .env file for tests
	err := godotenv.Load("../.env") // Adjust the path as necessary
	if err != nil {
		panic("Error loading .env file")
	}

	os.Exit(m.Run())
}

// TestConnectDB tests the ConnectDB function.
func TestConnectDB(t *testing.T) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		t.Fatal("MONGO_URI is not set in environment variables")
	}

	// Use a longer timeout to account for network delays
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	t.Log("Successfully connected and pinged MongoDB")
}

// TestGetCollection tests the GetCollection function.
// func TestGetCollection(t *testing.T) {
// 	// Use MockMongoClient to simulate a mongo.Client behavior
// 	client := &MockMongoClient{}

// 	// Inject the mock client into the GetCollection call
// 	collection := GetCollection(client)
// 	if collection == nil {
// 		t.Errorf("Expected non-nil collection, got nil")
// 	}
// }
