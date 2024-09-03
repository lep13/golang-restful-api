package db

import (
	"context"
	"errors"
	"testing"

	"os"

	"github.com/joho/godotenv"
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


