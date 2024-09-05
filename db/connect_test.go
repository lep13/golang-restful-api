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
	Client *mongo.Client // Simulate the actual mongo.Client struct
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

type MockCursor struct {
	mock.Mock
}

func (m *MockCursor) Next(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockCursor) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCursor) Decode(val interface{}) error {
	args := m.Called(val)
	return args.Error(0)
}

type MockSingleResult struct {
	mock.Mock
}

func (m *MockSingleResult) Decode(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
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

func TestGetCollection(t *testing.T) {
	mongoURI := os.Getenv("MONGO_URI") // Ensure this is set to a valid working MongoDB URI for the test
	client := ConnectDB(mongoURI)

	collection := GetCollection(client)
	assert.NotNil(t, collection)
	assert.Equal(t, "users", collection.Name()) // Verify the collection name is "users"
}

func TestNewMongoCollectionWrapper(t *testing.T) {
	mockCollection := new(mongo.Collection)
	wrapper := NewMongoCollectionWrapper(mockCollection)
	assert.NotNil(t, wrapper)
	assert.Equal(t, mockCollection, wrapper.(*MongoCollectionWrapper).collection)
}
