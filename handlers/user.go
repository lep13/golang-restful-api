package handlers

import (
    "context"
    "encoding/json"
    "github.com/lep13/golang-restful-api/db"
    "github.com/lep13/golang-restful-api/models"
    "log"
    "net/http"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

var collection *mongo.Collection

// Initialize is called to set up the database connection and collection
func Initialize(mongoURI string) {
    err := db.ConnectDB(mongoURI)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    collection = db.GetCollection()
    if collection == nil {
        log.Fatal("Failed to get collection. Collection is nil.")
    }
}

// CreateUser creates a new user in the database
func CreateUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }
    user.ID = primitive.NewObjectID()
    result, err := collection.InsertOne(context.TODO(), user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(result)
}

// GetUsers retrieves all users from the database
func GetUsers(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var users []models.User
    cur, err := collection.Find(context.TODO(), bson.M{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cur.Close(context.TODO())
    for cur.Next(context.TODO()) {
        var user models.User
        if err := cur.Decode(&user); err != nil {
            log.Println("Failed to decode user:", err)
            continue
        }
        users = append(users, user)
    }
    if err := cur.Err(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(users)
}

// GetUser retrieves a single user by ID from the database
func GetUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, err := primitive.ObjectIDFromHex(params["id"])
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    var user models.User
    if err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user); err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(user)
}

// DeleteUser deletes a user by ID from the database
func DeleteUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, err := primitive.ObjectIDFromHex(params["id"])
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    result, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    if result.DeletedCount == 0 {
        http.Error(w, "No user found to delete", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode("User deleted successfully")
}

// UpdateUser updates a user by ID in the database
func UpdateUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, err := primitive.ObjectIDFromHex(params["id"])
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }
    update := bson.M{"$set": user}
    result, err := collection.UpdateOne(context.TODO(), bson.M{"_id": id}, update)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if result.MatchedCount == 0 {
        http.Error(w, "No user found to update", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode("User updated successfully")
}
