package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lep13/golang-restful-api/db"
	"github.com/lep13/golang-restful-api/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MockMongoClient for testing purposes
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

// MockCollection for testing purposes
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

func (m *MockCollection) FindOne(ctx context.Context, filter interface{}) db.MongoSingleResultInterface {
	args := m.Called(ctx, filter)
	return args.Get(0).(db.MongoSingleResultInterface)
}

func (m *MockCollection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

// MockDatabase for testing purposes
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	args := m.Called(name, opts)
	return args.Get(0).(*mongo.Collection)
}


func TestHealthCheck(t *testing.T) {
	// Create a request to pass to the handler
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	// Call the handler directly
	handler := http.HandlerFunc(handlers.HealthCheck)
	handler.ServeHTTP(rr, req)

	// Check the response code is 200
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response body
	expected := `{"status":"healthy"}`
	assert.JSONEq(t, expected, rr.Body.String())
}

func TestCreateUser_Success(t *testing.T) {
	// Mock the collection
	mockCollection := new(MockCollection)
	mockCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{}, nil)

	// Initialize the handler with the mocked collection
	handlers.Initialize(mockCollection)

	// Create a request
	userData := `{"name":"John Doe","email":"johndoe@example.com","password":"1234"}`
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString(userData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(handlers.CreateUser)
	handler.ServeHTTP(rr, req)

	// Check response code and body
	assert.Equal(t, http.StatusOK, rr.Code)
	mockCollection.AssertCalled(t, "InsertOne", mock.Anything, mock.Anything)
}
