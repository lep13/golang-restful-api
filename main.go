package main

import (
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "github.com/lep13/golang-restful-api/db"
    "github.com/lep13/golang-restful-api/handlers"
)

// RunServer sets up and starts the server
func RunServer() {
    // Load environment variables
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Get MongoDB URI from environment
    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" {
        log.Fatal("MongoDB URI is not set in environment variables")
    }

    // Initialize MongoDB connection
    mongoClient := db.ConnectDB(mongoURI)

    // Get the collection from the MongoDB client
    collection := db.GetCollection(mongoClient)

    // Wrap the collection and pass it to the handlers
    wrappedCollection := db.NewMongoCollectionWrapper(collection)
    handlers.Initialize(wrappedCollection)

    // Set up router
    r := mux.NewRouter()

    // Root handler to display confirmation message
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, err := w.Write([]byte("Your API is up and running on port 5000!"))
        if err != nil {
            log.Printf("Error writing response: %v", err)
        }
    }).Methods("GET")

    // CRUD endpoints
    r.HandleFunc("/users", handlers.CreateUser).Methods("POST")
    r.HandleFunc("/users", handlers.GetUsers).Methods("GET")
    r.HandleFunc("/users/{id}", handlers.GetUser).Methods("GET")
    r.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT")
    r.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE")

    // Health check endpoint
    r.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

    // Get the port from environment variables or default to 5000
    port := os.Getenv("PORT")
    if port == "" {
        port = "5000"
    }

    // Server configuration
    srv := &http.Server{
        Handler:      r,
        Addr:         ":" + port,
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    // Start the server
    log.Printf("Server is running on port %s...", port)
    log.Fatal(srv.ListenAndServe())
}

func main() {
    RunServer()
}
