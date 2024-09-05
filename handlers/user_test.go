package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/lep13/golang-restful-api/db"
	"github.com/lep13/golang-restful-api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MockSingleResult implements db.MongoSingleResultInterface for unit testing
type MockSingleResult struct {
	mock.Mock
}

func (m *MockSingleResult) Decode(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

// MockCollection implements db.MongoCollectionInterface for unit testing
type MockCollection struct {
	mock.Mock
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollection) Find(ctx context.Context, filter interface{}) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter)
	// Handle nil case gracefully to prevent panic
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

func (m *MockCollection) FindOne(ctx context.Context, filter interface{}) db.MongoSingleResultInterface {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil // Gracefully return nil if no result is found
	}
	return args.Get(0).(db.MongoSingleResultInterface)
}

func (m *MockCollection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

type MockCursor struct {
	mock.Mock
}

// Simulate the Next method for iterating through documents.
func (m *MockCursor) Next(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

// Simulate the Decode method for retrieving documents.
func (m *MockCursor) Decode(val interface{}) error {
	args := m.Called(val)
	return args.Error(0)
}

// Simulate the Close method for closing the cursor.
func (m *MockCursor) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Simulate checking for errors in the cursor.
func (m *MockCursor) Err() error {
	args := m.Called()
	return args.Error(0)
}

// MongoCursorInterface defines the interface for working with MongoDB cursors
type MongoCursorInterface interface {
	Next(ctx context.Context) bool
	Decode(val interface{}) error
	Close(ctx context.Context) error
	Err() error
}

type FaultyResponseWriter struct {
	http.ResponseWriter
}

func (f *FaultyResponseWriter) Write(b []byte) (int, error) {
	// Simulate encoding failure
	return 0, fmt.Errorf("encoding failure")
}

// SetupMockCollection is used to inject a mock collection for testing
func SetupMockCollection(mock db.MongoCollectionInterface) func() {
	originalCollection := mongoCollection
	mongoCollection = mock
	return func() {
		mongoCollection = originalCollection
	}
}

func TestHealthCheck(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(HealthCheck)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "healthy")
}

func TestCreateUser(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	user := models.User{Name: "John Doe", Password: "password123"}
	user.ID = primitive.NewObjectID()

	mockCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{InsertedID: user.ID}, nil)

	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreateUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockCollection.AssertCalled(t, "InsertOne", mock.Anything, mock.Anything)
}

func TestGetUser(t *testing.T) {
	mockCollection := new(MockCollection)
	mockSingleResult := new(MockSingleResult)
	defer SetupMockCollection(mockCollection)()

	// Define the expected user object
	expectedUser := models.User{
		ID:       primitive.NewObjectID(),
		Name:     "John Doe",
		Password: "password123",
	}

	// Set up expectations for the mock
	mockSingleResult.On("Decode", mock.AnythingOfType("*models.User")).Run(func(args mock.Arguments) {
		user := args.Get(0).(*models.User)
		user.ID = expectedUser.ID
		user.Name = expectedUser.Name
		user.Password = expectedUser.Password
	}).Return(nil)

	mockCollection.On("FindOne", mock.Anything, bson.M{"_id": expectedUser.ID}).Return(mockSingleResult)

	req, _ := http.NewRequest("GET", "/users/"+expectedUser.ID.Hex(), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": expectedUser.ID.Hex()})

	handler := http.HandlerFunc(GetUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockCollection.AssertCalled(t, "FindOne", mock.Anything, bson.M{"_id": expectedUser.ID})
	mockSingleResult.AssertExpectations(t)
}

func TestDeleteUser(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	userID := primitive.NewObjectID()
	mockCollection.On("DeleteOne", mock.Anything, bson.M{"_id": userID}).Return(&mongo.DeleteResult{DeletedCount: 1}, nil)

	req, _ := http.NewRequest("DELETE", "/users/"+userID.Hex(), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": userID.Hex()})

	handler := http.HandlerFunc(DeleteUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockCollection.AssertCalled(t, "DeleteOne", mock.Anything, bson.M{"_id": userID})
}

func TestUpdateUser(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	user := models.User{Name: "Jane Doe", Password: "lepakshi57983"}
	user.ID = primitive.NewObjectID()

	mockCollection.On("UpdateOne", mock.Anything, bson.M{"_id": user.ID}, bson.M{"$set": user}).Return(&mongo.UpdateResult{MatchedCount: 1}, nil)

	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("PUT", "/users/"+user.ID.Hex(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": user.ID.Hex()})

	handler := http.HandlerFunc(UpdateUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockCollection.AssertCalled(t, "UpdateOne", mock.Anything, bson.M{"_id": user.ID}, bson.M{"$set": user})
}

func TestCreateUser_InvalidBody(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	// Simulate invalid request body
	invalidBody := []byte("invalid body")
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreateUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to decode request body")
}

func TestUpdateUser_InvalidBody(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	userID := primitive.NewObjectID()

	// Simulate invalid request body
	invalidBody := []byte("invalid body")
	req, _ := http.NewRequest("PUT", "/users/"+userID.Hex(), bytes.NewBuffer(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": userID.Hex()})

	handler := http.HandlerFunc(UpdateUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to decode request body")
}

func TestDeleteUser_UserNotFound(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	userID := primitive.NewObjectID()

	// Simulate DeleteOne returning 0 deleted count
	mockCollection.On("DeleteOne", mock.Anything, bson.M{"_id": userID}).Return(&mongo.DeleteResult{DeletedCount: 0}, nil)

	req, _ := http.NewRequest("DELETE", "/users/"+userID.Hex(), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": userID.Hex()})

	handler := http.HandlerFunc(DeleteUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockCollection.AssertCalled(t, "DeleteOne", mock.Anything, bson.M{"_id": userID})
}

func TestCreateUser_InsertOneFailure(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	// Return an actual InsertOneResult and a simulated error
	mockCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{}, errors.New("insert error"))

	user := models.User{Name: "John Doe", Password: "password123"}
	body, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreateUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "insert error")
}

func TestUpdateUser_UpdateOneFailure(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	user := models.User{Name: "Jane Doe", Password: "password123"}
	userID := primitive.NewObjectID()

	// Return an actual UpdateResult and a simulated error
	mockCollection.On("UpdateOne", mock.Anything, bson.M{"_id": userID}, bson.M{"$set": user}).Return(&mongo.UpdateResult{}, errors.New("update error"))

	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("PUT", "/users/"+userID.Hex(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": userID.Hex()})

	handler := http.HandlerFunc(UpdateUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "update error")
}

func TestDeleteUser_DeleteOneFailure(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	userID := primitive.NewObjectID()

	// Return an actual DeleteResult and a simulated error
	mockCollection.On("DeleteOne", mock.Anything, bson.M{"_id": userID}).Return(&mongo.DeleteResult{}, errors.New("delete error"))

	req, _ := http.NewRequest("DELETE", "/users/"+userID.Hex(), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": userID.Hex()})

	handler := http.HandlerFunc(DeleteUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "delete error")
}

func TestGetUsers_FindFailure(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	// Return a non-nil cursor with an error
	mockCollection.On("Find", mock.Anything, mock.Anything).Return(new(mongo.Cursor), errors.New("find error"))

	req, _ := http.NewRequest("GET", "/users", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetUsers)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "find error")
}

func TestInitialize(t *testing.T) {
	mockCollection := new(MockCollection)
	Initialize(mockCollection)

	// Assert that mongoCollection was properly set
	assert.Equal(t, mockCollection, mongoCollection)
}

func TestGetUser_InvalidIDFormat(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users/invalid-id", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid ID format")
}

func TestDeleteUser_InvalidIDFormat(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/users/invalid-id", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DeleteUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid ID format")
}

func TestUpdateUser_InvalidIDFormat(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/users/invalid-id", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpdateUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid ID format")
}

func TestUpdateUser_UserNotFound(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	userID := primitive.NewObjectID()
	mockCollection.On("UpdateOne", mock.Anything, mock.Anything, mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 0}, nil)

	req, _ := http.NewRequest("PUT", "/users/"+userID.Hex(), bytes.NewBuffer([]byte(`{"name":"John"}`)))
	rr := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"id": userID.Hex()})
	handler := http.HandlerFunc(UpdateUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "User not found")
}

func TestGetUser_UserNotFound(t *testing.T) {
	mockCollection := new(MockCollection)
	defer SetupMockCollection(mockCollection)()

	userID := primitive.NewObjectID()
	mockResult := new(MockSingleResult)
	mockResult.On("Decode", mock.Anything).Return(mongo.ErrNoDocuments)
	mockCollection.On("FindOne", mock.Anything, mock.Anything).Return(mockResult)

	req, _ := http.NewRequest("GET", "/users/"+userID.Hex(), nil)
	rr := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"id": userID.Hex()})
	handler := http.HandlerFunc(GetUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "User not found")
}

///////////////////////////////
///////////////////////////////
///////////////////////////////

// func TestGetUsers_CursorOperations(t *testing.T) {
//     mockCollection := new(MockCollection)
//     defer SetupMockCollection(mockCollection)()

//     mockCursor := new(MockCursor)

//     // Simulate 'Find' returning the mock cursor
//     mockCollection.On("Find", mock.Anything, mock.Anything).Return(mockCursor, nil)

//     // Simulate cursor behavior
//     mockCursor.On("Next", mock.Anything).Return(true).Once()  // Simulate one user found
//     mockCursor.On("Decode", mock.Anything).Return(nil).Once()  // Successfully decode one user
//     mockCursor.On("Next", mock.Anything).Return(false).Once()  // End of cursor
//     mockCursor.On("Close", mock.Anything).Return(nil).Once()   // Simulate closing the cursor
//     mockCursor.On("Err").Return(nil)                           // No cursor error

//     req, _ := http.NewRequest("GET", "/users", nil)
//     rr := httptest.NewRecorder()

//     handler := http.HandlerFunc(GetUsers)
//     handler.ServeHTTP(rr, req)

//     // Check if the response status is OK and the body contains expected data
//     assert.Equal(t, http.StatusOK, rr.Code)
//     assert.Contains(t, rr.Body.String(), "[]") // Assuming no users in the mock
// }

// func TestGetUsers_CursorError(t *testing.T) {
//     mockCollection := new(MockCollection)
//     defer SetupMockCollection(mockCollection)()

//     mockCursor := new(MockCursor)

//     // Simulate 'Find' returning the mock cursor
//     mockCollection.On("Find", mock.Anything, mock.Anything).Return(mockCursor, nil)

//     // Simulate cursor behavior
//     mockCursor.On("Next", mock.Anything).Return(true).Once()
//     mockCursor.On("Decode", mock.Anything).Return(nil).Once()  // Decode successfully
//     mockCursor.On("Next", mock.Anything).Return(false).Once()  // End of cursor
//     mockCursor.On("Close", mock.Anything).Return(nil).Once()   // Simulate closing the cursor
//     mockCursor.On("Err").Return(fmt.Errorf("cursor error"))    // Simulate cursor error

//     req, _ := http.NewRequest("GET", "/users", nil)
//     rr := httptest.NewRecorder()

//     handler := http.HandlerFunc(GetUsers)
//     handler.ServeHTTP(rr, req)

//     // Check if the test properly handles the cursor error
//     assert.Equal(t, http.StatusInternalServerError, rr.Code)
//     assert.Contains(t, rr.Body.String(), "cursor error")
// }

// func TestGetUser_InternalServerError(t *testing.T) {
//     mockCollection := new(MockCollection)
//     defer SetupMockCollection(mockCollection)()

//     // Simulate FindOne returning an internal server error
//     mockCollection.On("FindOne", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("some internal error"))

//     req, _ := http.NewRequest("GET", "/users/valid-id", nil)
//     rr := httptest.NewRecorder()

//     handler := http.HandlerFunc(GetUser)
//     handler.ServeHTTP(rr, req)

//     // Check if the test properly returns the internal server error
//     assert.Equal(t, http.StatusInternalServerError, rr.Code)
//     assert.Contains(t, rr.Body.String(), "some internal error")
// }
