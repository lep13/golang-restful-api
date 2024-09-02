package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    "github.com/joho/godotenv" 
    "golang-restful-api/handlers"
)

func main() {
    // Load environment variables from the .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Check if the MongoDB URI environment variable is loaded
    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" {
        log.Fatal("MongoDB URI is not set in environment variables")
    } else {
        log.Printf("MongoDB URI loaded: %s", mongoURI) // Debugging statement
    }

    // Set up the router
    r := mux.NewRouter()

    // Define routes for CRUD operations
    r.HandleFunc("/users", handlers.CreateUser).Methods("POST")
    r.HandleFunc("/users", handlers.GetUsers).Methods("GET")
    r.HandleFunc("/users/{id}", handlers.GetUser).Methods("GET")
    r.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT")
    r.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE")

    // Start the server
    port := os.Getenv("PORT")
    if port == "" {
        port = "3000"
    }

    log.Printf("Server is running on port %s...", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}
