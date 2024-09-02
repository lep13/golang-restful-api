package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "github.com/lep13/golang-restful-api/handlers"
)

func main() {
    // Load environment variables from the .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Get MongoDB URI and Port from environment variables
    mongoURI := os.Getenv("MONGO_URI")
    port := os.Getenv("PORT")

    // Set default port if not specified
    if port == "" {
        port = "3000"
    }

    // Initialize handlers with the MongoDB URI
    handlers.Initialize(mongoURI)

    // Set up the router
    r := mux.NewRouter()

    // Define routes for CRUD operations
    r.HandleFunc("/users", handlers.CreateUser).Methods("POST")
    r.HandleFunc("/users", handlers.GetUsers).Methods("GET")
    r.HandleFunc("/users/{id}", handlers.GetUser).Methods("GET")
    r.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT")
    r.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE")

    // Start the server
    log.Printf("Server is running on port %s...", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}
