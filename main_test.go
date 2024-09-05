package main

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/lep13/golang-restful-api/handlers"
	"github.com/lep13/golang-restful-api/db"
)

// MockCollection simulates a MongoDB collection for unit tests
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

// MockSingleResult simulates a MongoDB single result
type MockSingleResult struct {
	mock.Mock
}

func (m *MockSingleResult) Decode(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func TestRunServer(t *testing.T) {
	// Mock MongoDB collection and result
	mockCollection := new(MockCollection)
	mockSingleResult := new(MockSingleResult)
	mockCollection.On("FindOne", mock.Anything, mock.Anything).Return(mockSingleResult)

	// Initialize the handlers with the mock collection
	handlers.Initialize(mockCollection)

	// Set up router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Your API is up and running on port 5000!"))
	}).Methods("GET")
	r.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// Test the root endpoint
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Your API is up and running on port 5000!")

	// Test the health check endpoint
	req, _ = http.NewRequest("GET", "/health", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "healthy")
}
