package db

import (
	"context"
	"errors"
	"testing"

	"os"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MockMongoClient simulates the MongoDB client behavior.
type MockMongoClient struct {
    mock.Mock
}

func (m *MockMongoClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
    args := m.Called(ctx, rp)
    return args.Error(0)
}

func (m *MockMongoClient) Database(name string, opts ...*options.DatabaseOptions) *mongo.Database {
    args := m.Called(name, opts)
    return args.Get(0).(*mongo.Database)
}

type MockDatabase struct {
    mock.Mock
}

func (m *MockDatabase) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
    args := m.Called(name, opts)
    return args.Get(0).(*mongo.Collection)
}

// MockCollection simulates a MongoDB collection.
type MockCollection struct {
    mock.Mock
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
    args := m.Called(ctx, document)
    return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollection) Find(ctx context.Context, filter interface{}) (*mongo.Cursor, error) {
    args := m.Called(ctx, filter)
    return args.Get(0).(*mongo.Cursor), args.Error(1)
}

func (m *MockCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
    args := m.Called(ctx, filter)
    return args.Get(0).(*mongo.SingleResult)
}

func (m *MockCollection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
    args := m.Called(ctx, filter)
    return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
    args := m.Called(ctx, filter, update)
    return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func TestMain(m *testing.M) {
	// Load environment variables from .env file for tests
	err := godotenv.Load("../.env") // Adjust the path as necessary
	if err != nil {
		panic("Error loading .env file")
	}

	os.Exit(m.Run())
}

func TestConnectDB_Success(t *testing.T) {
	originalMongoClient := MongoClient
	defer func() { MongoClient = originalMongoClient }()

	mockClient := new(MockMongoClient)
	mockClient.On("Ping", mock.Anything, mock.Anything).Return(errors.New("failed to ping MongoDB"))

	MongoClient = mockClient

	mongoURI := os.Getenv("MONGO_URI")
	ConnectDB(mongoURI)
}

func TestMongoClientWrapper_Ping(t *testing.T) {
	// Create a mock MongoDB client
	mockClient := new(MockMongoClient)
	mockClient.On("Ping", mock.Anything, mock.Anything).Return(nil)

	// Simulate the Ping method call
	err := mockClient.Ping(context.Background(), readpref.Primary())
	assert.NoError(t, err)

	// Assert that the Ping method was called
	mockClient.AssertCalled(t, "Ping", mock.Anything, mock.Anything)
}

// func TestMongoClientWrapper_Database(t *testing.T) {
// 	// Create a mock MongoDB client
// 	mockClient := new(MockMongoClient)
// 	mockDatabase := new(MockDatabase)

// 	// Setup the mock to return a database
// 	mockClient.On("Database", "pipeline_task", mock.Anything).Return(mockDatabase)

// 	// Call the Database method
// 	db := mockClient.Database("pipeline_task", nil)
// 	assert.NotNil(t, db)

// 	// Assert that the Database method was called with the correct arguments
// 	mockClient.AssertCalled(t, "Database", "pipeline_task", mock.Anything)
// }

// func TestConnectDB_ConnectionFailure(t *testing.T) {
// 	// Save the original log.Fatalf
// 	originalFatalf := log.Fatalf
// 	defer func() { log.Fatalf = originalFatalf }()

// 	// Mock log.Fatalf to panic instead of exiting
// 	log.Fatalf = func(format string, args ...interface{}) {
// 		panic(fmt.Sprintf(format, args...))
// 	}

// 	mongoURI := "mongodb://invalid-uri"

// 	// Expecting a panic due to log.Fatalf
// 	assert.PanicsWithValue(t, fmt.Sprintf("Failed to connect to MongoDB: %v", errors.New("connection failed")), func() {
// 		ConnectDB(mongoURI)
// 	})
// }

// func TestConnectDB_PingFailure(t *testing.T) {
// 	// Save the original log.Fatalf
// 	originalFatalf := log.Fatalf
// 	defer func() { log.Fatalf = originalFatalf }()

// 	// Mock log.Fatalf to panic instead of exiting
// 	log.Fatalf = func(format string, args ...interface{}) {
// 		panic(fmt.Sprintf(format, args...))
// 	}

// 	mockClient := new(MockMongoClient)
// 	mockClient.On("Ping", mock.Anything, mock.Anything).Return(errors.New("ping failed"))

// 	MongoClient = mockClient
// 	mongoURI := "mongodb://mock-uri"

// 	// Expecting a panic due to log.Fatalf
// 	assert.PanicsWithValue(t, fmt.Sprintf("Failed to ping MongoDB: %v", errors.New("ping failed")), func() {
// 		ConnectDB(mongoURI)
// 	})
// }

// func TestGetCollection(t *testing.T) {
// 	mockClient := new(MockMongoClient)
// 	mockDatabase := new(MockDatabase)
// 	mockCollection := new(mongo.Collection)

// 	// Setup the mock to return a database and collection
// 	mockClient.On("Database", "pipeline_task", mock.Anything).Return(mockDatabase)
// 	mockDatabase.On("Collection", "users", mock.Anything).Return(mockCollection)

// 	collection := GetCollection(mockClient)
// 	assert.NotNil(t, collection)

// 	// Verify that the methods were called
// 	mockClient.AssertCalled(t, "Database", "pipeline_task", mock.Anything)
// 	mockDatabase.AssertCalled(t, "Collection", "users", mock.Anything)
// }
