package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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

// MockCollectionInterface mocks the methods of a MongoDB collection.
type MockCollectionInterface struct {
	mock.Mock
}

func (m *MockCollectionInterface) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollectionInterface) Find(ctx context.Context, filter interface{}) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

func (m *MockCollectionInterface) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

func (m *MockCollectionInterface) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollectionInterface) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

// MockSingleResult mocks the behavior of a mongo.SingleResult.
type MockSingleResult struct {
	mock.Mock
}

func (m *MockSingleResult) Decode(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

// InjectMockCollection replaces the global collection with a mock.
func InjectMockCollection(mock db.MongoCollectionInterface) func() {
	original := collection
	collection = mock.(*mongo.Collection) // Cast back to original for runtime
	return func() { collection = original }
}

func TestCreateUser(t *testing.T) {
	// Setup
	mockCollection := new(MockCollectionInterface)
	defer InjectMockCollection(mockCollection)()

	user := models.User{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	userBytes, _ := json.Marshal(user)

	// Mock behaviors
	mockCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{InsertedID: primitive.NewObjectID()}, nil)

	// Execute
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(userBytes))
	assert.NoError(t, err)
	rr := httptest.NewRecorder()
	CreateUser(rr, req)

	// Verify
	assert.Equal(t, http.StatusOK, rr.Code)
	var createdUser models.User
	json.Unmarshal(rr.Body.Bytes(), &createdUser)
	assert.Equal(t, user.Name, createdUser.Name)
	assert.Equal(t, user.Email, createdUser.Email)
	mockCollection.AssertCalled(t, "InsertOne", mock.Anything, mock.Anything)
}

func TestGetUsers(t *testing.T) {
	// Setup
	mockCollection := new(MockCollectionInterface)
	defer InjectMockCollection(mockCollection)()

	// Mock cursor to simulate data retrieval
	cursor := new(mongo.Cursor)
	mockCollection.On("Find", mock.Anything, bson.M{}).Return(cursor, nil)

	// Execute
	req, err := http.NewRequest("GET", "/users", nil)
	assert.NoError(t, err)
	rr := httptest.NewRecorder()
	GetUsers(rr, req)

	// Verify
	assert.Equal(t, http.StatusOK, rr.Code)
	mockCollection.AssertCalled(t, "Find", mock.Anything, bson.M{})
}

func TestGetUser(t *testing.T) {
	// Setup
	mockCollection := new(MockCollectionInterface)
	defer InjectMockCollection(mockCollection)()

	user := models.User{
		ID:    primitive.NewObjectID(),
		Name:  "Jane Doe",
		Email: "jane@example.com",
	}

	mockSingleResult := new(MockSingleResult)
	mockSingleResult.On("Decode", mock.Anything).Return(func(v interface{}) error {
		*v.(*models.User) = user
		return nil
	})

	mockCollection.On("FindOne", mock.Anything, mock.Anything).Return(mockSingleResult)

	// Execute
	req, err := http.NewRequest("GET", "/users/{id}", nil)
	req = mux.SetURLVars(req, map[string]string{"id": user.ID.Hex()})
	assert.NoError(t, err)
	rr := httptest.NewRecorder()
	GetUser(rr, req)

	// Verify
	assert.Equal(t, http.StatusOK, rr.Code)
	var fetchedUser models.User
	json.Unmarshal(rr.Body.Bytes(), &fetchedUser)
	assert.Equal(t, user.ID, fetchedUser.ID)
	assert.Equal(t, user.Name, fetchedUser.Name)
	assert.Equal(t, user.Email, fetchedUser.Email)
	mockCollection.AssertCalled(t, "FindOne", mock.Anything, mock.Anything)
}

func TestDeleteUser(t *testing.T) {
	// Setup
	mockCollection := new(MockCollectionInterface)
	defer InjectMockCollection(mockCollection)()

	userID := primitive.NewObjectID()

	mockCollection.On("DeleteOne", mock.Anything, bson.M{"_id": userID}).Return(&mongo.DeleteResult{DeletedCount: 1}, nil)

	// Execute
	req, err := http.NewRequest("DELETE", "/users/{id}", nil)
	req = mux.SetURLVars(req, map[string]string{"id": userID.Hex()})
	assert.NoError(t, err)
	rr := httptest.NewRecorder()
	DeleteUser(rr, req)

	// Verify
	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "User deleted successfully", response["message"])
	mockCollection.AssertCalled(t, "DeleteOne", mock.Anything, bson.M{"_id": userID})
}

func TestUpdateUser(t *testing.T) {
	// Setup
	mockCollection := new(MockCollectionInterface)
	defer InjectMockCollection(mockCollection)()

	user := models.User{
		Name:  "Updated User",
		Email: "updated@example.com",
	}
	userID := primitive.NewObjectID()
	userBytes, _ := json.Marshal(user)

	mockCollection.On("UpdateOne", mock.Anything, bson.M{"_id": userID}, mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1}, nil)

	// Execute
	req, err := http.NewRequest("PUT", "/users/{id}", bytes.NewBuffer(userBytes))
	req = mux.SetURLVars(req, map[string]string{"id": userID.Hex()})
	assert.NoError(t, err)
	rr := httptest.NewRecorder()
	UpdateUser(rr, req)

	// Verify
	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "User updated successfully", response["message"])
	mockCollection.AssertCalled(t, "UpdateOne", mock.Anything, bson.M{"_id": userID}, mock.Anything)
}
