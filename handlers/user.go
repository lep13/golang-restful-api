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

// Initialize initializes the collection and database connection
func Initialize(mongoURI string) {
    client := db.ConnectDB(mongoURI)
    collection = db.GetCollection(client)
}

// CreateUser creates a new user in the database
func CreateUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Failed to decode request body", http.StatusBadRequest)
        return
    }
    user.ID = primitive.NewObjectID()
    _, err := collection.InsertOne(context.TODO(), user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if err := json.NewEncoder(w).Encode(user); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
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
    if err := json.NewEncoder(w).Encode(users); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}

// GetUser retrieves a single user by ID from the database
func GetUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, err := primitive.ObjectIDFromHex(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }
    var user models.User
    err = collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            http.Error(w, "User not found", http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }
    if err := json.NewEncoder(w).Encode(user); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}

// DeleteUser deletes a user by ID from the database
func DeleteUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, err := primitive.ObjectIDFromHex(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }
    res, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if res.DeletedCount == 0 {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    response := map[string]string{"message": "User deleted successfully"}
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}

// UpdateUser updates a user by ID in the database
func UpdateUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, err := primitive.ObjectIDFromHex(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Failed to decode request body", http.StatusBadRequest)
        return
    }
    update := bson.M{"$set": user}
    res, err := collection.UpdateOne(context.TODO(), bson.M{"_id": id}, update)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if res.MatchedCount == 0 {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    response := map[string]string{"message": "User updated successfully"}
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}